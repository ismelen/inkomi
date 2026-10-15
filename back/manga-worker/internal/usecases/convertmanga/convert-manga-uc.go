package convertmanga

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sync/errgroup"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/common/fsutil"
	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
)

type ConvertMangaUC struct {
	queue        queue.IQueue
	cloudStorage cloud.ICloudStorage
	extractors   map[models.SourceFormat]ports.IPageExtractor
	processor    ports.IPageProcessor
	builder      ports.IMangaEpubBuilder
	concurrency  int
}

// NewConvertMangaUC builds the UC. concurrency bounds how many pages of a job are loaded in
// memory at once.
func NewConvertMangaUC(
	queue queue.IQueue,
	cloudStorage cloud.ICloudStorage,
	extractors map[models.SourceFormat]ports.IPageExtractor,
	processor ports.IPageProcessor,
	builder ports.IMangaEpubBuilder,
	concurrency int,
) *ConvertMangaUC {
	return &ConvertMangaUC{
		queue:        queue,
		cloudStorage: cloudStorage,
		extractors:   extractors,
		processor:    processor,
		builder:      builder,
		concurrency:  max(1, concurrency),
	}
}

// Execute converts the cbz or pdf of msg into a fixed-layout epub stored as msg.Id/<title>.epub,
// publishes it on the next step and deletes the source.
func (c *ConvertMangaUC) Execute(ctx context.Context, msg jobs.Msg, _ jobrun.Delivery) error {
	var rawCfg []byte
	rtl := false
	if msg.Manga != nil {
		rawCfg, rtl = msg.MangaCfg, msg.Rtl
	}
	cfg, err := models.NewMangaConfig(rawCfg)
	if err != nil {
		return jobrun.Permanent(err)
	}

	format, err := models.SourceFormatFromFilename(msg.Filename)
	if err != nil {
		return jobrun.Permanent(fmt.Errorf("%s: %w", msg.Filename, err))
	}
	extractor, ok := c.extractors[format]
	if !ok {
		return jobrun.Permanent(fmt.Errorf("%s: %w", msg.Filename, models.ErrUnsupportedFormat))
	}

	title := cmp.Or(msg.Title, strings.TrimSuffix(msg.Filename, filepath.Ext(msg.Filename)))
	epubName, err := fsutil.SanitizeFilename(title + ".epub")
	if err != nil {
		return jobrun.Permanent(fmt.Errorf("epub filename %q: %w", title, err))
	}

	if err := c.queue.Publish(ctx, string(jobs.SubjectStart), msg); err != nil {
		return err
	}

	dir, err := os.MkdirTemp("", "manga-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		// best effort: a leftover temp dir doesn't affect the result
		_ = os.RemoveAll(dir)
	}()

	srcPath := filepath.Join(dir, "source"+filepath.Ext(msg.Filename))
	if err := c.cloudStorage.DownloadFile(ctx, msg.Id, msg.Filename, srcPath); err != nil {
		if c.alreadyConverted(ctx, msg.Id, epubName) {
			return nil
		}
		return err
	}

	pages, err := c.convert(ctx, extractor, srcPath, dir, cfg, rtl)
	if err != nil {
		return err
	}

	epubPath := filepath.Join(dir, "book.epub")
	meta := models.EpubMeta{Title: title, RTL: rtl, Width: cfg.TargetWidth, Height: cfg.TargetHeight}
	if err := c.builder.Build(epubPath, meta, pages); err != nil {
		return jobrun.Permanent(fmt.Errorf("build epub: %w", err))
	}

	size, err := c.cloudStorage.UploadFile(ctx, msg.Id, epubName, epubPath)
	if err != nil {
		return err
	}

	original := msg.Filename
	msg.Filename, msg.Size, msg.Manga = epubName, size, nil
	if err := c.queue.Publish(ctx, string(NextSubject(msg)), msg); err != nil {
		return err
	}

	// The epub is already published: retrying because the source couldn't be deleted would
	// redo the whole conversion, so the failure is only logged.
	if err := c.cloudStorage.Delete(ctx, msg.Id, original); err != nil {
		log.Printf("convert manga %s: delete %s: %v", msg.Id, original, err)
	}
	return nil
}

// convert extracts the pages of srcPath and processes them in parallel, returning the processed
// pages in reading order.
func (c *ConvertMangaUC) convert(
	ctx context.Context,
	extractor ports.IPageExtractor,
	srcPath, dir string,
	cfg models.MangaConfig,
	rtl bool,
) ([]models.EpubPage, error) {
	rawDir, outDir := filepath.Join(dir, "raw"), filepath.Join(dir, "out")
	for _, d := range []string{rawDir, outDir} {
		if err := os.Mkdir(d, 0o755); err != nil {
			return nil, fmt.Errorf("create %s: %w", d, err)
		}
	}

	paths, err := extractor.Extract(ctx, srcPath, rawDir)
	if err != nil {
		return nil, permanentUnlessCanceled(fmt.Errorf("extract pages: %w", err))
	}
	if len(paths) == 0 {
		return nil, jobrun.Permanent(models.ErrNoPages)
	}

	processed := make([][]models.EpubPage, len(paths))
	group, gctx := errgroup.WithContext(ctx)
	group.SetLimit(c.concurrency)
	for i, path := range paths {
		group.Go(func() error {
			pages, err := c.processPage(gctx, path, outDir, i, cfg, rtl)
			processed[i] = pages
			return err
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}

	var pages []models.EpubPage
	for _, p := range processed {
		pages = append(pages, p...)
	}
	return pages, nil
}

func (c *ConvertMangaUC) processPage(
	ctx context.Context,
	path, outDir string,
	index int,
	cfg models.MangaConfig,
	rtl bool,
) ([]models.EpubPage, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read page %d: %w", index+1, err)
	}

	processed, err := c.processor.Process(ctx, data, cfg, rtl)
	if err != nil {
		return nil, permanentUnlessCanceled(fmt.Errorf("process page %d: %w", index+1, err))
	}

	pages := make([]models.EpubPage, len(processed))
	for j, p := range processed {
		outPath := filepath.Join(outDir, fmt.Sprintf("%04d-%d%s", index+1, j, p.Format.Ext()))
		if err := os.WriteFile(outPath, p.Data, 0o644); err != nil {
			return nil, fmt.Errorf("write page %d: %w", index+1, err)
		}
		pages[j] = models.EpubPage{Path: outPath, Width: p.Width, Height: p.Height, Split: p.Split, Format: p.Format}
	}
	return pages, nil
}

// alreadyConverted reports whether a redelivered job was finished by a previous attempt that
// deleted the source after uploading the epub.
func (c *ConvertMangaUC) alreadyConverted(ctx context.Context, id, epubName string) bool {
	_, err := c.cloudStorage.Size(ctx, id, epubName)
	return err == nil
}

// permanentUnlessCanceled marks err as permanent, since a corrupt archive or image fails again
// on retry, unless it was caused by the job being canceled.
func permanentUnlessCanceled(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	return jobrun.Permanent(err)
}

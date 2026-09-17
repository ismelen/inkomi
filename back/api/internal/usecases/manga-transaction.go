package usecases

import (
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain/convert"
	"ismelen/inkomi/internal/domain/manga"
	"ismelen/inkomi/internal/infra/epub"
	"ismelen/inkomi/internal/infra/fs"
	"ismelen/inkomi/internal/shared/uid"
	"os"
	"path/filepath"
	"runtime"

	"golang.org/x/sync/errgroup"
)

type GetChapterFunc = func(file string, chaptersDir string) (*manga.Chapter, error)

type MangaTransactionUC struct {
	BaseTransactionUC
	imageSettings *manga.ImageSettings
	getChapter    GetChapterFunc
}

func NewMangaTransactionUC(
	pushNotifier convert.PushNotifier,
	cloud convert.CloudStorage,
) *MangaTransactionUC {
	t := &MangaTransactionUC{
		BaseTransactionUC: BaseTransactionUC{
			pushNotifier: pushNotifier,
			cloud:        cloud,
		},
		imageSettings: manga.NewDefaultImageSettings(),
	}
	t.processor = t
	return t
}

func (c MangaTransactionUC) GetImageSettings() *manga.ImageSettings {
	return c.imageSettings
}

func (c MangaTransactionUC) GetChapter(file string, chaptersDir string) (*manga.Chapter, error) {
	ext := filepath.Ext(file)

	var filename, chapterPath string
	var pages []string
	var err error

	switch ext {
	case ".pdf":
		filename, chapterPath, pages, err = fs.UnzipPdf(file, chaptersDir)
	case ".cbz":
		filename, chapterPath, pages, err = fs.UnzipFile(file, chaptersDir)
	default:
		return nil, fmt.Errorf("Not valid format")
	}

	if err != nil {
		return nil, err
	}
	return manga.NewChapter(filename, chapterPath, pages), nil
}

func (c MangaTransactionUC) Process(file *convert.TransactionFile, tran *convert.Transaction, transPath string) *convert.TransactionResultFile {
	defer os.RemoveAll(filepath.Join(file.SrcPath))
	dir := filepath.Dir(file.SrcPath)
	builder := epub.NewEpubBuilder().
		SetSettings(c.imageSettings, tran.Config.Profile).
		Start(file.Name, dir)

	workers := max(1, runtime.NumCPU()*3/4)

	chaptersDir := filepath.Join(dir, "chapters")
	defer os.RemoveAll(chaptersDir)

	chapter, err := c.GetChapter(file.SrcPath, chaptersDir)
	if err != nil {
		file.SetError(err)
		return nil
	}

	group, gctx := errgroup.WithContext(context.Background())
	group.SetLimit(workers)

	pages := chapter.GetOrderedPagePaths(".jpg", ".jpeg", ".png", ".webp")
	processedPages := make([]*manga.Page, len(pages))

	for pIdx, pagePath := range pages {
		idx, path := pIdx, pagePath
		group.Go(func() error {
			if err := gctx.Err(); err != nil {
				return fmt.Errorf("Job canceled")
			}
			page, err := c.ProcessPage(path, idx+1, tran.Config.Profile, c.imageSettings)
			if err != nil {
				return err
			}
			processedPages[idx] = page
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		file.SetError(err)
		return nil
	}

	for i, page := range processedPages {
		builder.AddPage(page, i == 0)
	}

	path, err := builder.Build()
	if err != nil {
		file.SetError(err)
		return nil
	}
	filename := filepath.Base(path)

	return convert.NewTransactionResultFile(uid.GetRandomID(6), filename, path, file.Size, []*convert.TransactionFile{file})
}

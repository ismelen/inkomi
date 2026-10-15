package joinepubs

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/shared/fsutil"
)

var errMissingParent = errors.New("join job without parent or total")

type JoinEpubsUC struct {
	queue        ports.Queue
	joinStore    ports.JoinStore
	cloudStorage ports.CloudStorage
	merger       ports.EpubMerger
	maxBlockSize int64
}

func NewJoinEpubsUC(
	queue ports.Queue,
	joinStore ports.JoinStore,
	cloudStorage ports.CloudStorage,
	merger ports.EpubMerger,
	maxBlockSize int64,
) *JoinEpubsUC {
	return &JoinEpubsUC{
		queue:        queue,
		joinStore:    joinStore,
		cloudStorage: cloudStorage,
		merger:       merger,
		maxBlockSize: maxBlockSize,
	}
}

type publication struct {
	subject jobs.Subject
	msg     jobs.Msg
}

func (j *JoinEpubsUC) Execute() error {
	return j.queue.Subscribe(string(jobs.SubjectJoin), func(ctx context.Context, subject string, payload []byte) error {
		return j.handle(ctx, payload)
	})
}

func (j *JoinEpubsUC) handle(ctx context.Context, payload []byte) error {
	var msg jobs.Msg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return j.fail(ctx, msg, err)
	}
	if msg.Parent == nil || msg.Parent.Total <= 0 {
		return j.fail(ctx, msg, errMissingParent)
	}

	msgs, err := j.joinStore.Add(ctx, msg.Parent.Id, msg.Parent.Total, msg)
	if err != nil {
		return err
	}
	if msgs == nil {
		return nil
	}

	if err := j.join(ctx, msg.Parent, msgs); err != nil {
		if releaseErr := j.joinStore.Release(ctx, msg.Parent.Id); releaseErr != nil {
			return errors.Join(err, releaseErr)
		}
		return err
	}

	return j.joinStore.Clear(ctx, msg.Parent.Id)
}

func (j *JoinEpubsUC) join(ctx context.Context, parent *jobs.Parent, msgs []jobs.Msg) error {
	items, err := j.orderedItems(ctx, msgs)
	if err != nil {
		return err
	}

	blocks := SplitBlocks(items, j.maxBlockSize)
	publications := make([]publication, 0, len(blocks)+len(msgs))
	for i, block := range blocks {
		result, err := j.processBlock(ctx, parent, block, i, len(blocks))
		if err != nil {
			return err
		}
		publications = append(publications, result)
	}

	for _, child := range items {
		publications = append(publications, publication{
			subject: jobs.SubjectJoined,
			msg:     jobs.Msg{Id: child.Msg.Id, UserId: child.Msg.UserId},
		})
	}

	for _, p := range publications {
		if err := j.queue.Publish(ctx, string(p.subject), p.msg); err != nil {
			return err
		}
	}
	return nil
}

func (j *JoinEpubsUC) orderedItems(ctx context.Context, msgs []jobs.Msg) ([]models.JoinItem, error) {
	sorted := slices.Clone(msgs)
	slices.SortStableFunc(sorted, func(a, b jobs.Msg) int {
		return a.Parent.Index - b.Parent.Index
	})

	items := make([]models.JoinItem, len(sorted))
	for i, msg := range sorted {
		size, err := j.cloudStorage.Size(ctx, msg.Id, msg.Filename)
		if err != nil {
			return nil, fmt.Errorf("size of %s/%s: %w", msg.Id, msg.Filename, err)
		}
		items[i] = models.JoinItem{Msg: msg, Size: size}
	}
	return items, nil
}

// processBlock merges and uploads one block. A merge failure is reported as a failed result
// instead of an error, because retrying would fail again on the same files.
func (j *JoinEpubsUC) processBlock(ctx context.Context, parent *jobs.Parent, block []models.JoinItem, index, total int) (publication, error) {
	first := block[0].Msg
	title := models.BlockTitle(parent.Title, block, index, total)
	filename, err := fsutil.SanitizeFilename(title + ".epub")
	if err != nil {
		return publication{}, fmt.Errorf("result filename %q: %w", title, err)
	}

	result := jobs.Msg{
		Id:       ResultId(parent.Id, index),
		UserId:   first.UserId,
		Title:    title,
		Filename: filename,
		Kepubify: first.Kepubify,
		ToCloud:  first.ToCloud,
		Result:   &jobs.Result{ParentId: parent.Id, Index: index, Total: total},
	}

	dir, err := os.MkdirTemp("", "join-epubs-*")
	if err != nil {
		return publication{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer func() {
		// best effort: a leftover temp dir doesn't affect the result
		_ = os.RemoveAll(dir)
	}()

	books, err := j.download(ctx, dir, block)
	if err != nil {
		return publication{}, err
	}

	outputPath := filepath.Join(dir, filename)
	meta := models.EpubMeta{Title: title, RTL: parent.Rtl}
	if err := j.merger.Merge(outputPath, meta, books); err != nil {
		result.ErrorMsg = fmt.Sprintf("merge %s: %v", title, err)
		return publication{subject: jobs.SubjectFailed, msg: result}, nil
	}

	result.Size, err = j.upload(ctx, result.Id, filename, outputPath)
	if err != nil {
		return publication{}, err
	}

	if result.Kepubify {
		return publication{subject: jobs.SubjectKepub, msg: result}, nil
	}
	return publication{subject: jobs.SubjectDone, msg: result}, nil
}

func (j *JoinEpubsUC) download(ctx context.Context, dir string, block []models.JoinItem) ([]models.EpubBook, error) {
	books := make([]models.EpubBook, len(block))
	for i, item := range block {
		books[i] = models.EpubBook{
			Path:  filepath.Join(dir, fmt.Sprintf("%04d.epub", i)),
			Title: cmp.Or(item.Msg.Title, strings.TrimSuffix(item.Msg.Filename, filepath.Ext(item.Msg.Filename))),
		}
		if err := j.downloadFile(ctx, item.Msg, books[i].Path); err != nil {
			return nil, err
		}
	}
	return books, nil
}

func (j *JoinEpubsUC) downloadFile(ctx context.Context, msg jobs.Msg, path string) error {
	reader, err := j.cloudStorage.Download(ctx, msg.Id, msg.Filename)
	if err != nil {
		return fmt.Errorf("download %s/%s: %w", msg.Id, msg.Filename, err)
	}
	defer reader.Close()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if _, err := io.Copy(file, reader); err != nil {
		return errors.Join(fmt.Errorf("write %s: %w", path, err), file.Close())
	}
	return file.Close()
}

func (j *JoinEpubsUC) upload(ctx context.Context, id, filename, path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}

	if err := j.cloudStorage.Upload(ctx, id, filename, file); err != nil {
		return 0, fmt.Errorf("upload %s/%s: %w", id, filename, err)
	}
	return info.Size(), nil
}

// fail reports an unprocessable job; returning the publish result acks the message so it isn't retried.
func (j *JoinEpubsUC) fail(ctx context.Context, msg jobs.Msg, err error) error {
	return j.queue.Publish(ctx, string(jobs.SubjectFailed), jobs.Msg{
		Id:       msg.Id,
		UserId:   msg.UserId,
		ErrorMsg: err.Error(),
	})
}

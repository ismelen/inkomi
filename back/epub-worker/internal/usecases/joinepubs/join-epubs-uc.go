package joinepubs

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/shared/fsutil"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/cloudfile"
)

var (
	errMissingParent   = errors.New("join job without parent or total")
	errNoValidChildren = errors.New("no child of the folder is a plain .epub")
)

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

// Execute buffers a child of a join folder and, once the whole folder has arrived, merges it.
func (j *JoinEpubsUC) Execute(ctx context.Context, msg jobs.Msg, delivery jobrun.Delivery) error {
	if msg.Parent == nil || msg.Parent.Total <= 0 {
		return jobrun.Permanent(errMissingParent)
	}

	msgs, err := j.joinStore.Add(ctx, msg.Parent.Id, msg.Parent.Total, msg)
	if err != nil {
		return err
	}
	if msgs == nil {
		return nil
	}

	merged, err := j.join(ctx, msg.Parent, msgs)
	if err != nil {
		return j.giveBack(ctx, delivery, msg, err)
	}

	// The results are already published: retrying because of a cleanup failure would fail
	// on what was already cleaned, so these failures are only logged.
	if err := j.joinStore.Clear(ctx, msg.Parent.Id); err != nil {
		log.Printf("join %s: clear group: %v", msg.Parent.Id, err)
	}
	if err := j.deleteSources(ctx, merged); err != nil {
		log.Printf("join %s: delete merged sources: %v", msg.Parent.Id, err)
	}
	return nil
}

// giveBack releases the group so a redelivery can retry it or, when no retry is left,
// clears it and reports the failure on the parent folder.
func (j *JoinEpubsUC) giveBack(ctx context.Context, delivery jobrun.Delivery, msg jobs.Msg, err error) error {
	err = jobrun.For(jobs.Msg{Id: msg.Parent.Id, UserId: msg.UserId}, err)
	if !delivery.LastAttempt && !jobrun.IsPermanent(err) {
		return errors.Join(err, j.joinStore.Release(ctx, msg.Parent.Id))
	}
	if clearErr := j.joinStore.Clear(ctx, msg.Parent.Id); clearErr != nil {
		log.Printf("join %s: clear group: %v", msg.Parent.Id, clearErr)
	}
	return err
}

// join merges the group and publishes its results. It returns the children whose block
// was merged, so their source files can be deleted.
func (j *JoinEpubsUC) join(ctx context.Context, parent *jobs.Parent, msgs []jobs.Msg) ([]jobs.Msg, error) {
	valid, err := j.rejectInvalid(ctx, msgs)
	if err != nil {
		return nil, err
	}

	items, err := j.orderedItems(ctx, valid)
	if err != nil {
		return nil, err
	}

	blocks := SplitBlocks(items, j.maxBlockSize)
	publications := make([]publication, 0, len(blocks)+len(items))
	var merged []jobs.Msg
	for i, block := range blocks {
		result, err := j.processBlock(ctx, parent, block, i, len(blocks))
		if err != nil {
			return nil, err
		}
		publications = append(publications, result)
		if result.subject != jobs.SubjectFailed {
			for _, item := range block {
				merged = append(merged, item.Msg)
			}
		}
	}

	for _, child := range items {
		publications = append(publications, publication{
			subject: jobs.SubjectJoined,
			msg:     jobs.Msg{Id: child.Msg.Id, UserId: child.Msg.UserId},
		})
	}

	for _, p := range publications {
		if err := j.queue.Publish(ctx, string(p.subject), p.msg); err != nil {
			return nil, err
		}
	}
	return merged, nil
}

func (j *JoinEpubsUC) deleteSources(ctx context.Context, msgs []jobs.Msg) error {
	var errs []error
	for _, msg := range msgs {
		if err := j.cloudStorage.Delete(ctx, msg.Id, msg.Filename); err != nil {
			errs = append(errs, fmt.Errorf("delete %s/%s: %w", msg.Id, msg.Filename, err))
		}
	}
	return errors.Join(errs...)
}

// rejectInvalid reports the children that aren't plain .epub files as failed and returns the rest.
func (j *JoinEpubsUC) rejectInvalid(ctx context.Context, msgs []jobs.Msg) ([]jobs.Msg, error) {
	valid := make([]jobs.Msg, 0, len(msgs))
	for _, child := range msgs {
		err := models.CheckEpubFilename(child.Filename)
		if err == nil {
			valid = append(valid, child)
			continue
		}
		child.ErrorMsg = fmt.Sprintf("%s: %v", child.Filename, err)
		if err := j.queue.Publish(ctx, string(jobs.SubjectFailed), child); err != nil {
			return nil, err
		}
	}
	if len(valid) == 0 {
		return nil, jobrun.Permanent(errNoValidChildren)
	}
	return valid, nil
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

	result.Size, err = cloudfile.Upload(ctx, j.cloudStorage, result.Id, filename, outputPath)
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
		if err := cloudfile.Download(ctx, j.cloudStorage, item.Msg.Id, item.Msg.Filename, books[i].Path); err != nil {
			return nil, err
		}
	}
	return books, nil
}

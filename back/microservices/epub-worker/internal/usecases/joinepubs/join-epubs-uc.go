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

	"github.com/ismelen/inkomi/back/microservices/common/cloud"
	"github.com/ismelen/inkomi/back/microservices/common/fsutil"
	"github.com/ismelen/inkomi/back/microservices/common/jobrun"
	"github.com/ismelen/inkomi/back/microservices/common/jobs"
	"github.com/ismelen/inkomi/back/microservices/common/queue"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/domain/ports"
)

var (
	errMissingParent   = errors.New("join job without parent or total")
	errNoValidChildren = errors.New("no child of the folder is a plain .epub")
)

type JoinEpubsUC struct {
	queue        queue.IQueue
	joinStore    ports.IJoinStore
	cloudStorage cloud.ICloudStorage
	merger       ports.IEpubMerger
	maxBlockSize int64
}

func NewJoinEpubsUC(
	queue queue.IQueue,
	joinStore ports.IJoinStore,
	cloudStorage cloud.ICloudStorage,
	merger ports.IEpubMerger,
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

	if err := j.join(ctx, msg.Parent, msgs); err != nil {
		return j.giveBack(ctx, delivery, msg, err)
	}

	// The results are already published: retrying because of a cleanup failure would fail
	// on what was already cleaned, so it's only logged.
	if err := j.joinStore.Clear(ctx, msg.Parent.Id); err != nil {
		log.Printf("join %s: clear group: %v", msg.Parent.Id, err)
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

// join merges the group block by block and publishes its results. The chapters of a block are
// deleted as soon as the block is stored, so the folder never needs its chapters and all its
// blocks on disk at once.
func (j *JoinEpubsUC) join(ctx context.Context, parent *jobs.Parent, msgs []jobs.Msg) error {
	valid, err := j.rejectInvalid(ctx, msgs)
	if err != nil {
		return err
	}

	items, err := j.orderedItems(ctx, valid)
	if err != nil {
		return err
	}

	blocks := SplitBlocks(items, j.maxBlockSize)
	publications := make([]publication, 0, len(blocks)+len(items))
	for i, block := range blocks {
		result, err := j.processBlock(ctx, parent, block, i, len(blocks))
		if err != nil {
			return err
		}
		publications = append(publications, result)
		if result.subject != jobs.SubjectFailed {
			// the block is stored: a retry reuses it instead of these chapters
			if err := j.deleteSources(ctx, block); err != nil {
				log.Printf("join %s: delete merged chapters: %v", parent.Id, err)
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
			return err
		}
	}
	return nil
}

func (j *JoinEpubsUC) deleteSources(ctx context.Context, block []models.JoinItem) error {
	var errs []error
	for _, item := range block {
		if err := j.cloudStorage.Delete(ctx, item.Msg.Id, item.Msg.Filename); err != nil {
			errs = append(errs, fmt.Errorf("delete %s/%s: %w", item.Msg.Id, item.Msg.Filename, err))
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
		// the size travels with the job, so a retry splits the same blocks even when the chapters
		// of the blocks already stored were deleted
		size := msg.Size
		if size <= 0 {
			var err error
			if size, err = j.cloudStorage.Size(ctx, msg.Id, msg.Filename); err != nil {
				return nil, fmt.Errorf("size of %s/%s: %w", msg.Id, msg.Filename, err)
			}
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

	// a previous attempt already stored this block before failing on a later one
	if size, err := j.cloudStorage.Size(ctx, result.Id, filename); err == nil && size > 0 {
		result.Size = size
		return j.stored(result), nil
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

	result.Size, err = j.cloudStorage.UploadFile(ctx, result.Id, filename, outputPath)
	if err != nil {
		return publication{}, err
	}
	return j.stored(result), nil
}

// stored is the publication of a block already in storage: on to kepub or done.
func (j *JoinEpubsUC) stored(result jobs.Msg) publication {
	if result.Kepubify {
		return publication{subject: jobs.SubjectKepub, msg: result}
	}
	return publication{subject: jobs.SubjectDone, msg: result}
}

func (j *JoinEpubsUC) download(ctx context.Context, dir string, block []models.JoinItem) ([]models.EpubBook, error) {
	books := make([]models.EpubBook, len(block))
	for i, item := range block {
		books[i] = models.EpubBook{
			Path:  filepath.Join(dir, fmt.Sprintf("%04d.epub", i)),
			Title: cmp.Or(item.Msg.Title, strings.TrimSuffix(item.Msg.Filename, filepath.Ext(item.Msg.Filename))),
		}
		if err := j.cloudStorage.DownloadFile(ctx, item.Msg.Id, item.Msg.Filename, books[i].Path); err != nil {
			return nil, err
		}
	}
	return books, nil
}

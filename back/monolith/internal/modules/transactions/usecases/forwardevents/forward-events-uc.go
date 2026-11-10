package forwardevents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
)

type ForwardEventsUC struct {
	hub        ports.ISocketHub
	queue      ports.IQueue
	sourceRepo ports.ISourceRepository
	quota      ports.IStorageQuota
}

func NewForwardEventsUC(
	hub ports.ISocketHub,
	queue ports.IQueue,
	sourceRepo ports.ISourceRepository,
	quota ports.IStorageQuota,
) *ForwardEventsUC {
	return &ForwardEventsUC{
		hub:        hub,
		queue:      queue,
		sourceRepo: sourceRepo,
		quota:      quota,
	}
}

func (f *ForwardEventsUC) Execute() error {
	return f.queue.Subscribe("job.step.>", func(ctx context.Context, subject string, payload []byte) error {
		var job jobs.Msg
		if err := json.Unmarshal(payload, &job); err != nil {
			return err
		}

		status := models.StatusFromSubject(jobs.Subject(subject))
		if job.Result != nil {
			return f.forwardResult(ctx, job, status)
		}

		msg := map[string]any{
			"id":     job.Id,
			"status": status,
		}

		if job.Filename != "" || job.Size > 0 {
			if err := f.sourceRepo.UpdateFile(ctx, job.Id, job.UserId, job.Filename, job.Size); err != nil {
				return err
			}
			if job.Filename != "" {
				msg["filename"] = job.Filename
			}
			if job.Size > 0 {
				msg["size"] = job.Size
			}
		}

		var err error
		if status == models.SourceStatusFailed && job.ErrorMsg != "" {
			err = f.sourceRepo.UpdateErrorMsg(ctx, job.Id, job.UserId, job.ErrorMsg)
			msg["error"] = job.ErrorMsg
		} else {
			err = f.sourceRepo.UpdateStatus(ctx, job.Id, job.UserId, status)
		}

		if err != nil {
			return err
		}
		logTransition(job, status)
		f.releaseIfFinished(ctx, job, status)

		return f.send(job.UserId, msg)
	})
}

func logTransition(job jobs.Msg, status models.SourceStatus) {
	if job.ErrorMsg != "" {
		log.Printf("source %s user %d: %s: %s", job.Id, job.UserId, status, job.ErrorMsg)
		return
	}
	log.Printf("source %s user %d: %s", job.Id, job.UserId, status)
}

// send pushes msg to the user's socket; a user without the app open isn't an error, the app
// catches up with GET /uploads.
func (f *ForwardEventsUC) send(userId int, msg map[string]any) error {
	if err := f.hub.Send(userId, msg); err != nil && !errors.Is(err, models.ErrClientNotConnected) {
		return err
	}
	return nil
}

func (f *ForwardEventsUC) forwardResult(ctx context.Context, job jobs.Msg, status models.SourceStatus) error {
	// a cancelled folder takes no new results from a join that was already running
	if parent, err := f.sourceRepo.GetByIdAndUserId(ctx, job.Result.ParentId, job.UserId); err == nil &&
		parent != nil && parent.Status == models.SourceStatusCancelled {
		return nil
	}
	if err := f.sourceRepo.UpsertResult(ctx, models.NewResultSource(job, status)); err != nil {
		return err
	}

	result, err := f.sourceRepo.GetByIdAndUserId(ctx, job.Id, job.UserId)
	if err != nil {
		return err
	}

	msg := map[string]any{
		"id":     result.Id,
		"status": result.Status,
		"result": result,
	}
	if result.ErrorMsg != nil {
		msg["error"] = *result.ErrorMsg
	}
	logTransition(job, status)
	if err := f.send(job.UserId, msg); err != nil {
		return err
	}

	if err := f.finishParent(ctx, job); err != nil {
		return err
	}
	f.releaseIfFinished(ctx, job, status)
	return nil
}

// finishParent sets the parent folder status once all its results are finished.
func (f *ForwardEventsUC) finishParent(ctx context.Context, job jobs.Msg) error {
	parentId := job.Result.ParentId
	results, err := f.sourceRepo.ListResultsByParentIdAndUserId(ctx, parentId, job.UserId)
	if err != nil {
		return err
	}
	if len(results) < job.Result.Total {
		return nil
	}

	failed := 0
	for i := range results {
		if !results[i].IsFinished() {
			return nil
		}
		if results[i].Status == models.SourceStatusFailed {
			failed++
		}
	}

	msg := map[string]any{"id": parentId}
	if failed > 0 {
		errorMsg := fmt.Sprintf("%d of %d results failed", failed, len(results))
		err = f.sourceRepo.UpdateErrorMsg(ctx, parentId, job.UserId, errorMsg)
		msg["status"] = models.SourceStatusFailed
		msg["error"] = errorMsg
	} else {
		err = f.sourceRepo.UpdateStatus(ctx, parentId, job.UserId, models.SourceStatusDone)
		msg["status"] = models.SourceStatusDone
	}
	if err != nil {
		return err
	}

	return f.send(job.UserId, msg)
}

// releaseIfFinished frees the disk reservation of the job's upload once its whole tree reached a
// final status. A failure only delays the release until the quota sweep, so it isn't an error.
func (f *ForwardEventsUC) releaseIfFinished(ctx context.Context, job jobs.Msg, status models.SourceStatus) {
	switch status {
	case models.SourceStatusDone, models.SourceStatusSent, models.SourceStatusFailed, models.SourceStatusJoined:
	default:
		return
	}
	if err := f.quota.ReleaseFinished(ctx, job.Id, job.UserId); err != nil {
		log.Printf("storage: release after %s %s: %v", job.Id, status, err)
	}
}

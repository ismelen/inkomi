package forwaredevents

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type ForwardEventsUC struct {
	hub        ports.SocketHub
	queue      ports.Queue
	sourceRepo ports.SourceRepository
}

func NewForwardEventsUC(
	hub ports.SocketHub,
	queue ports.Queue,
	sourceRepo ports.SourceRepository,
) *ForwardEventsUC {
	return &ForwardEventsUC{
		hub,
		queue,
		sourceRepo,
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

		return f.hub.Send(job.UserId, msg)
	})
}

func (f *ForwardEventsUC) forwardResult(ctx context.Context, job jobs.Msg, status models.SourceStatus) error {
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
	if err := f.hub.Send(job.UserId, msg); err != nil {
		return err
	}

	return f.finishParent(ctx, job)
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

	return f.hub.Send(job.UserId, msg)
}

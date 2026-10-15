package forwaredevents

import (
	"context"
	"encoding/json"

	"github.com/ismelen/inkomi/back/contracts/jobs"
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
		msg := map[string]any{
			"id":     job.Id,
			"status": status,
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

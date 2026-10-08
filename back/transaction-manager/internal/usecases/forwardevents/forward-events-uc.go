package forwaredevents

import (
	"context"
	"encoding/json"

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

func (f *ForwardEventsUC) Execute() {
	f.queue.Subscribe("job.step.>", func(ctx context.Context, subject string, payload []byte) error {
		var step stepMsgDTO
		if err := json.Unmarshal(payload, &step); err != nil {
			return err
		}

		stepKey := subject[9:] // len("job.step.")
		status := f.getStatus(stepKey)
		msg := map[string]any{
			"id":     step.Id,
			"status": status,
		}

		var errorStr *string
		if step.Error != "" {
			errorStr = &step.Error
			msg["error"] = step.Error
		}

		if err := f.sourceRepo.UpdateStatus(ctx, step.Id, step.UserId, status, errorStr); err != nil {
			return err
		}

		return f.hub.Send(step.UserId, msg)
	})
}

func (f *ForwardEventsUC) getStatus(key string) models.SourceStatus {
	switch key {
	case "start":
		return models.SourceStatusProcessing
	case "epub.kepub":
		return models.SourceStatusKepubifying
	case "epub.join":
		return models.SourceStatusJoining
	case "done":
		return models.SourceStatusDone
	case "sent":
		return models.SourceStatusSent
	case "error":
		return models.SourceStatusFailed
	default:
		return models.SourceStatusFailed
	}
}

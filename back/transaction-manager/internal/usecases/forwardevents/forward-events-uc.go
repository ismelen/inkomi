package forwardevents

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

func (f *ForwardEventsUC) Execute() error {
	return f.queue.Subscribe("job.step.>", func(ctx context.Context, subject string, payload []byte) error {
		stepKey := subject[9:] // len("job.step.")
		switch stepKey {
		case "error":
			return f.handleError(ctx, payload)
		case "waitjoin":
			return f.handleWaitJoin(ctx, payload)
		}

		return nil
	})
}

func (f *ForwardEventsUC) handleWaitJoin(ctx context.Context, payload []byte) error {
	var msg WaitJoinMsgDTO
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}

	if err := f.sourceRepo.UpdateStatus(ctx, msg.Id, msg.UserId, models.SourceStatusWaitingJoin); err != nil {
		return err
	}

	return f.hub.Send(msg.UserId, map[string]any{
		"id":     msg.Id,
		"status": models.SourceStatusWaitingJoin,
	})
}

func (f *ForwardEventsUC) handleError(ctx context.Context, payload []byte) error {
	var msg ErrorMsgDTO
	if err := json.Unmarshal(payload, &msg); err != nil {
		return err
	}

	if err := f.sourceRepo.SetError(ctx, msg.Id, msg.UserId, msg.Error); err != nil {
		return err
	}

	return f.hub.Send(msg.UserId, map[string]any{
		"id":     msg.Id,
		"status": models.SourceStatusFailed,
		"error":  msg.Error,
	})
}

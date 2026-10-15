package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/contracts/jobs"
)

type EventHandler func(ctx context.Context, subject string, payload []byte) error

type Queue interface {
	Publish(ctx context.Context, subject jobs.Subject, payload []byte) error
	Subscribe(subject jobs.Subject, handler EventHandler) error
	Close() error
}

package queue

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
)

type IQueue interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(subject string, handler jobrun.DeliveryHandler) error
	Close() error
}

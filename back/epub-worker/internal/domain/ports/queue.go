package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/common/jobrun"
)

type EventHandler func(ctx context.Context, delivery jobrun.Delivery) error

type Queue interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(subject string, handler EventHandler) error
	Close() error
}

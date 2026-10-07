package ports

import "context"

type EventHandler func(ctx context.Context, payload []byte) error

type Queue interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	Subscribe(ctx context.Context, subject string, handler EventHandler) error
	Close() error
}


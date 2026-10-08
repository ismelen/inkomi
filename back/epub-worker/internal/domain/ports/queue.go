package ports

import "context"

type EventHandler func(ctx context.Context, subject string, payload []byte) error

type Queue interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(subject string, handler EventHandler) error
	Close() error
}

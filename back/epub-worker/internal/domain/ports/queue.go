package ports

import "context"

// Delivery is a message received from the queue. LastAttempt is true when a failure
// won't be redelivered, so the handler must report it instead of relying on a retry.
type Delivery struct {
	Subject     string
	Payload     []byte
	LastAttempt bool
}

type EventHandler func(ctx context.Context, delivery Delivery) error

type Queue interface {
	Publish(ctx context.Context, subject string, payload any) error
	Subscribe(subject string, handler EventHandler) error
	Close() error
}

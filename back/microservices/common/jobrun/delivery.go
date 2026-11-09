package jobrun

import "context"

// Delivery is a message received from the queue. LastAttempt is true when a failure
// won't be redelivered, so it must be reported instead of relying on a retry.
type Delivery struct {
	Subject     string
	Payload     []byte
	LastAttempt bool
}

type DeliveryHandler func(ctx context.Context, delivery Delivery) error

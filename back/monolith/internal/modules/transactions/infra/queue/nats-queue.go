package queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsQueue struct {
	nc   *nats.Conn
	js   jetstream.JetStream
	subs []*nats.Subscription
}

// NewNatsQueue uses the connection nc, which the caller owns and closes.
func NewNatsQueue(nc *nats.Conn) (*NatsQueue, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	return &NatsQueue{
		nc: nc,
		js: js,
	}, nil
}

func (n *NatsQueue) ensureStream(ctx context.Context, subject jobs.Subject) error {
	streamName := strings.ReplaceAll(string(subject), ".", "_")
	streamName = strings.ReplaceAll(streamName, "*", "all")
	streamName = strings.ReplaceAll(streamName, ">", "all")

	_, err := n.js.Stream(ctx, streamName)
	if err != nil {
		_, err = n.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     streamName,
			Subjects: []string{string(subject)},
		})
		if err != nil {
			return fmt.Errorf("failed to create stream for subject %s: %w", subject, err)
		}
	}
	return nil
}

func (n *NatsQueue) Publish(ctx context.Context, subject jobs.Subject, payload []byte) error {
	if err := n.ensureStream(ctx, subject); err != nil {
		return err
	}

	_, err := n.js.Publish(ctx, string(subject), payload)
	if err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", subject, err)
	}

	return nil
}

func (n *NatsQueue) Subscribe(subject jobs.Subject, handler ports.EventHandler) error {
	sub, err := n.nc.Subscribe(string(subject), func(msg *nats.Msg) {
		hCtx := context.Background()
		_ = handler(hCtx, msg.Subject, msg.Data)
	})

	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	n.subs = append(n.subs, sub)

	return nil
}

func (n *NatsQueue) Close() error {
	for _, sub := range n.subs {
		// the subscription dies with the connection anyway
		_ = sub.Unsubscribe()
	}
	return nil
}

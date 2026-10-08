package queue

import (
	"context"
	"fmt"
	"strings"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsQueue struct {
	nc   *nats.Conn
	js   jetstream.JetStream
	subs []*nats.Subscription
}

func NewNatsQueue(url string) (*NatsQueue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	return &NatsQueue{
		nc: nc,
		js: js,
	}, nil
}

func (n *NatsQueue) ensureStream(ctx context.Context, subject string) error {
	streamName := strings.ReplaceAll(subject, ".", "_")
	streamName = strings.ReplaceAll(streamName, "*", "all")
	streamName = strings.ReplaceAll(streamName, ">", "all")

	_, err := n.js.Stream(ctx, streamName)
	if err != nil {
		_, err = n.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
			Name:     streamName,
			Subjects: []string{subject},
		})
		if err != nil {
			return fmt.Errorf("failed to create stream for subject %s: %w", subject, err)
		}
	}
	return nil
}

func (n *NatsQueue) Publish(ctx context.Context, subject string, payload []byte) error {
	if err := n.ensureStream(ctx, subject); err != nil {
		return err
	}

	_, err := n.js.Publish(ctx, subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", subject, err)
	}

	return nil
}

func (n *NatsQueue) Subscribe(subject string, handler ports.EventHandler) error {
	sub, err := n.nc.Subscribe(subject, func(msg *nats.Msg) {
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
		_ = sub.Unsubscribe()
	}
	n.nc.Close()
	return nil
}

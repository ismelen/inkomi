package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsQueue struct {
	nc      *nats.Conn
	js      jetstream.JetStream
	subs    []jetstream.ConsumeContext
	ackWait time.Duration
}

// NewNatsQueue connects to NATS; ackWait is how long a handler may run before the message is redelivered.
func NewNatsQueue(url string, ackWait time.Duration) (*NatsQueue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	return &NatsQueue{
		nc:      nc,
		js:      js,
		ackWait: ackWait,
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

func (n *NatsQueue) Publish(ctx context.Context, subject string, payload any) error {
	if err := n.ensureStream(ctx, subject); err != nil {
		return err
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = n.js.Publish(ctx, subject, bytes)
	if err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", subject, err)
	}

	return nil
}

func (n *NatsQueue) Subscribe(subject string, handler ports.EventHandler) error {
	ctx := context.Background()
	if err := n.ensureStream(ctx, subject); err != nil {
		return err
	}

	streamName := strings.ReplaceAll(subject, ".", "_")
	streamName = strings.ReplaceAll(streamName, "*", "all")
	streamName = strings.ReplaceAll(streamName, ">", "all")

	consumer, err := n.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:   streamName,
		AckPolicy: jetstream.AckExplicitPolicy,
		AckWait:   n.ackWait,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer for %s: %w", subject, err)
	}

	subCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		hCtx := context.Background()
		err = handler(hCtx, msg.Subject(), msg.Data())
		if err != nil {
			msg.Nak()
		} else {
			msg.Ack()
		}
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	n.subs = append(n.subs, subCtx)
	return nil
}

func (n *NatsQueue) Close() error {
	for _, sub := range n.subs {
		sub.Stop()
	}
	n.nc.Close()
	return nil
}

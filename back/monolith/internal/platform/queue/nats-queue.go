package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const retryDelay = 10 * time.Second

var _ IQueue = (*NatsQueue)(nil)

var streamNameReplacer = strings.NewReplacer(".", "_", "*", "all", ">", "all")

type NatsQueue struct {
	nc         *nats.Conn
	js         jetstream.JetStream
	subs       []jetstream.ConsumeContext
	ackWait    time.Duration
	maxDeliver int
}

// NewNatsQueue uses the connection nc, which the caller owns and closes. ackWait is how long a
// handler may run before the message is redelivered and maxDeliver how many times a failing message
// is delivered in total.
func NewNatsQueue(nc *nats.Conn, ackWait time.Duration, maxDeliver int) (*NatsQueue, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	return &NatsQueue{
		nc:         nc,
		js:         js,
		ackWait:    ackWait,
		maxDeliver: maxDeliver,
	}, nil
}

func streamName(subject string) string {
	return streamNameReplacer.Replace(subject)
}

func (n *NatsQueue) ensureStream(ctx context.Context, subject string) error {
	name := streamName(subject)
	if _, err := n.js.Stream(ctx, name); err == nil {
		return nil
	}

	_, err := n.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     name,
		Subjects: []string{subject},
	})
	if err != nil {
		return fmt.Errorf("failed to create stream for subject %s: %w", subject, err)
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

	if _, err := n.js.Publish(ctx, subject, bytes); err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", subject, err)
	}
	return nil
}

// admissionPoll is how often a job waiting for SubscribeOptions.Ready checks again.
const admissionPoll = 2 * time.Second

// SubscribeOptions tunes how many jobs of a subject run at once.
type SubscribeOptions struct {
	// Concurrency is how many messages are handled at once (1 when not positive). It also bounds
	// how many are fetched ahead, so a message never waits in the client longer than a job while
	// its ack wait runs.
	Concurrency int
	// Ready, when set, gates the start of every job (e.g. on free memory). A message waiting for
	// it keeps reporting progress, so it isn't redelivered.
	Ready func() bool
}

func (n *NatsQueue) Subscribe(subject string, handler jobrun.DeliveryHandler) error {
	return n.SubscribeWith(subject, SubscribeOptions{Concurrency: 1}, handler)
}

func (n *NatsQueue) SubscribeWith(subject string, opts SubscribeOptions, handler jobrun.DeliveryHandler) error {
	ctx := context.Background()
	if err := n.ensureStream(ctx, subject); err != nil {
		return err
	}
	concurrency := max(1, opts.Concurrency)

	name := streamName(subject)
	consumer, err := n.js.CreateOrUpdateConsumer(ctx, name, jetstream.ConsumerConfig{
		Durable:    name,
		AckPolicy:  jetstream.AckExplicitPolicy,
		AckWait:    n.ackWait,
		MaxDeliver: n.maxDeliver,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer for %s: %w", subject, err)
	}

	slots := make(chan struct{}, concurrency)
	subCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		// blocking here stops the delivery of more messages until a job ends
		slots <- struct{}{}
		for opts.Ready != nil && !opts.Ready() {
			if err := msg.InProgress(); err != nil {
				log.Printf("nats: %s in progress: %v", msg.Subject(), err)
			}
			time.Sleep(admissionPoll)
		}
		go func() {
			defer func() { <-slots }()
			n.handle(msg, handler)
		}()
	}, jetstream.PullMaxMessages(concurrency))
	if err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", subject, err)
	}

	n.subs = append(n.subs, subCtx)
	return nil
}

func (n *NatsQueue) handle(msg jetstream.Msg, handler jobrun.DeliveryHandler) {
	attempt := uint64(1)
	if meta, err := msg.Metadata(); err == nil {
		attempt = meta.NumDelivered
	}

	err := handler(context.Background(), jobrun.Delivery{
		Subject:     msg.Subject(),
		Payload:     msg.Data(),
		LastAttempt: attempt >= uint64(n.maxDeliver),
	})
	var ackErr error
	if err != nil {
		log.Printf("nats: %s attempt %d failed: %v", msg.Subject(), attempt, err)
		ackErr = msg.NakWithDelay(time.Duration(attempt) * retryDelay)
	} else {
		ackErr = msg.Ack()
	}
	if ackErr != nil {
		log.Printf("nats: %s ack: %v", msg.Subject(), ackErr)
	}
}

func (n *NatsQueue) Close() error {
	for _, sub := range n.subs {
		sub.Stop()
	}
	return nil
}

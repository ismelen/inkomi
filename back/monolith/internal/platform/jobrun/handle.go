package jobrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
)

// IPublisher is the part of a queue jobrun needs to report failures.
type IPublisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

type Handler func(ctx context.Context, msg jobs.Msg, delivery Delivery) error

// ICancellations tells which jobs were cancelled and interrupts the running ones.
type ICancellations interface {
	IsCancelled(ids ...string) bool
	// Watch calls cancel when any of ids is cancelled, until stop is called.
	Watch(ids []string, cancel func()) (stop func())
}

type Option func(*options)

type options struct {
	cancellations ICancellations
}

// WithCancellations skips the jobs of cancelled uploads and interrupts the running ones: their
// context is cancelled and the failure that follows isn't reported.
func WithCancellations(c ICancellations) Option {
	return func(o *options) { o.cancellations = c }
}

// cancelIds are the ids whose cancellation stops msg: its own and its folder's.
func cancelIds(msg jobs.Msg) []string {
	ids := []string{msg.Id}
	if msg.Parent != nil {
		ids = append(ids, msg.Parent.Id)
	}
	if msg.Result != nil {
		ids = append(ids, msg.Result.ParentId)
	}
	return ids
}

// Handle adapts a job handler to a queue handler: it parses the jobs.Msg, runs h and reports
// failures on SubjectFailed. Permanent failures are reported right away and transient ones on
// the last attempt; otherwise the error is returned so the queue redelivers the job.
func Handle(pub IPublisher, h Handler, opts ...Option) DeliveryHandler {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return func(ctx context.Context, delivery Delivery) error {
		var msg jobs.Msg
		err := json.Unmarshal(delivery.Payload, &msg)
		if err == nil && o.cancellations != nil {
			ids := cancelIds(msg)
			if o.cancellations.IsCancelled(ids...) {
				log.Printf("job %s %s user %d: skipped, cancelled", delivery.Subject, msg.Id, msg.UserId)
				return nil
			}
			var cancel context.CancelFunc
			ctx, cancel = context.WithCancel(ctx)
			defer cancel()
			defer o.cancellations.Watch(ids, cancel)()
		}
		if err != nil {
			err = Permanent(fmt.Errorf("parse job: %w", err))
		} else {
			log.Printf("job %s %s user %d: start %q", delivery.Subject, msg.Id, msg.UserId, msg.Filename)
			start := time.Now()
			err = h(ctx, msg, delivery)
			if err == nil {
				log.Printf("job %s %s user %d: done in %s", delivery.Subject, msg.Id, msg.UserId, time.Since(start).Round(time.Millisecond))
			}
		}
		if err != nil && o.cancellations != nil && o.cancellations.IsCancelled(cancelIds(msg)...) {
			log.Printf("job %s %s user %d: interrupted, cancelled", delivery.Subject, msg.Id, msg.UserId)
			return nil
		}
		if err != nil {
			log.Printf("job %s %s user %d: failed (last attempt %t): %v", delivery.Subject, msg.Id, msg.UserId, delivery.LastAttempt, err)
		}
		return report(ctx, pub, delivery, msg, err)
	}
}

func report(ctx context.Context, pub IPublisher, delivery Delivery, msg jobs.Msg, err error) error {
	if err == nil {
		return nil
	}

	var e *Error
	permanent := errors.As(err, &e) && e.permanent
	if !permanent && !delivery.LastAttempt {
		return err
	}

	failed := msg
	if e != nil && e.msg != nil {
		failed = *e.msg
	}
	failed.ErrorMsg = err.Error()

	if pubErr := pub.Publish(ctx, string(jobs.SubjectFailed), failed); pubErr != nil {
		return errors.Join(err, pubErr)
	}
	return nil
}

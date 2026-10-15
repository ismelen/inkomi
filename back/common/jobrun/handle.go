package jobrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ismelen/inkomi/back/common/jobs"
)

// IPublisher is the part of a queue jobrun needs to report failures.
type IPublisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

type Handler func(ctx context.Context, msg jobs.Msg, delivery Delivery) error

// Handle adapts a job handler to a queue handler: it parses the jobs.Msg, runs h and reports
// failures on SubjectFailed. Permanent failures are reported right away and transient ones on
// the last attempt; otherwise the error is returned so the queue redelivers the job.
func Handle(pub IPublisher, h Handler) DeliveryHandler {
	return func(ctx context.Context, delivery Delivery) error {
		var msg jobs.Msg
		err := json.Unmarshal(delivery.Payload, &msg)
		if err != nil {
			err = Permanent(fmt.Errorf("parse job: %w", err))
		} else {
			err = h(ctx, msg, delivery)
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

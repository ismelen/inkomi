package jobfail

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
)

// Error marks how a use case failure must be reported.
type Error struct {
	err       error
	msg       *jobs.Msg
	permanent bool
}

func (e *Error) Error() string { return e.err.Error() }
func (e *Error) Unwrap() error { return e.err }

// Permanent marks err as one that a retry can't fix, so it's reported right away.
func Permanent(err error) error {
	return wrap(err, func(e *Error) { e.permanent = true })
}

// For reports err on msg instead of on the delivered job.
func For(msg jobs.Msg, err error) error {
	return wrap(err, func(e *Error) { e.msg = &msg })
}

func wrap(err error, apply func(*Error)) error {
	if err == nil {
		return nil
	}
	var e *Error
	if !errors.As(err, &e) {
		e = &Error{err: err}
	}
	apply(e)
	return e
}

// Report publishes SubjectFailed for permanent errors and for any error on the last attempt,
// and acks them. Other errors are returned so the queue redelivers the job.
func Report(ctx context.Context, queue ports.Queue, delivery ports.Delivery, err error) error {
	if err == nil {
		return nil
	}

	var e *Error
	permanent := errors.As(err, &e) && e.permanent
	if !permanent && !delivery.LastAttempt {
		return err
	}

	var failed jobs.Msg
	if e != nil && e.msg != nil {
		failed = *e.msg
	} else {
		// best effort: an unparsable payload is reported without job id
		_ = json.Unmarshal(delivery.Payload, &failed)
	}
	failed.ErrorMsg = err.Error()

	if pubErr := queue.Publish(ctx, string(jobs.SubjectFailed), failed); pubErr != nil {
		return errors.Join(err, pubErr)
	}
	return nil
}

// IsPermanent reports whether err was marked with Permanent.
func IsPermanent(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.permanent
}

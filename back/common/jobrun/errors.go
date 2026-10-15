package jobrun

import (
	"errors"

	"github.com/ismelen/inkomi/back/common/jobs"
)

// Error marks how a job failure must be reported.
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

// IsPermanent reports whether err was marked with Permanent.
func IsPermanent(err error) bool {
	var e *Error
	return errors.As(err, &e) && e.permanent
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

package jobrun_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
)

type fakePublisher struct {
	subjects []string
	msgs     []jobs.Msg
	err      error
}

func (p *fakePublisher) Publish(ctx context.Context, subject string, payload any) error {
	p.subjects = append(p.subjects, subject)
	p.msgs = append(p.msgs, payload.(jobs.Msg))
	return p.err
}

func payloadOf(t *testing.T, msg jobs.Msg) []byte {
	t.Helper()

	bytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func TestHandle(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("boom")
	cases := []struct {
		name          string
		payload       []byte
		handlerErr    error
		lastAttempt   bool
		expectRetry   bool
		expectedJobId *string
	}{
		{"Success", nil, nil, false, false, nil},
		{"Transient with retries left", nil, baseErr, false, true, nil},
		{"Transient on last attempt", nil, baseErr, true, false, new("job-1")},
		{"Permanent", nil, jobrun.Permanent(baseErr), false, false, new("job-1")},
		{"For other job", nil, jobrun.Permanent(jobrun.For(jobs.Msg{Id: "parent-1"}, baseErr)), false, false, new("parent-1")},
		{"Invalid payload", []byte("{"), nil, false, false, new("")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			pub := &fakePublisher{}
			payload := tc.payload
			if payload == nil {
				payload = payloadOf(t, jobs.Msg{Id: "job-1", UserId: 1})
			}
			var received jobs.Msg
			handler := jobrun.Handle(pub, func(ctx context.Context, msg jobs.Msg, delivery jobrun.Delivery) error {
				received = msg
				return tc.handlerErr
			})

			// Act
			err := handler(context.Background(), jobrun.Delivery{Payload: payload, LastAttempt: tc.lastAttempt})

			// Assert
			if tc.expectRetry != (err != nil) {
				t.Fatalf("expected retry %v, got error %v", tc.expectRetry, err)
			}
			if tc.expectRetry && !errors.Is(err, baseErr) {
				t.Fatalf("expected %v, got %v", baseErr, err)
			}
			if tc.payload == nil && received.Id != "job-1" {
				t.Fatalf("handler received %+v", received)
			}
			if tc.expectedJobId == nil {
				if len(pub.msgs) != 0 {
					t.Fatalf("expected no publish, got %+v", pub.msgs)
				}
				return
			}
			if len(pub.msgs) != 1 || pub.subjects[0] != string(jobs.SubjectFailed) {
				t.Fatalf("expected one failed publish, got %v %+v", pub.subjects, pub.msgs)
			}
			if pub.msgs[0].Id != *tc.expectedJobId || pub.msgs[0].ErrorMsg == "" {
				t.Fatalf("unexpected failed msg %+v", pub.msgs[0])
			}
		})
	}
}

func TestHandle_PublishFails_ReturnsErrorToRetry(t *testing.T) {
	t.Parallel()

	// Arrange
	pubErr := errors.New("nats down")
	pub := &fakePublisher{err: pubErr}
	handler := jobrun.Handle(pub, func(ctx context.Context, msg jobs.Msg, delivery jobrun.Delivery) error {
		return jobrun.Permanent(errors.New("boom"))
	})

	// Act
	err := handler(context.Background(), jobrun.Delivery{Payload: payloadOf(t, jobs.Msg{Id: "job-1"})})

	// Assert
	if !errors.Is(err, pubErr) {
		t.Fatalf("expected %v, got %v", pubErr, err)
	}
}

func TestIsPermanent(t *testing.T) {
	t.Parallel()

	base := errors.New("boom")
	if jobrun.IsPermanent(base) || jobrun.IsPermanent(jobrun.For(jobs.Msg{}, base)) {
		t.Fatal("unmarked errors must not be permanent")
	}
	if !jobrun.IsPermanent(jobrun.For(jobs.Msg{}, jobrun.Permanent(base))) {
		t.Fatal("For must keep the permanent mark")
	}
	if jobrun.Permanent(nil) != nil {
		t.Fatal("Permanent(nil) must be nil")
	}
}

package jobfail_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/jobfail"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReport(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("boom")
	payload, err := json.Marshal(jobs.Msg{Id: "job-1", UserId: 1, Filename: "a.epub"})
	require.NoError(t, err)

	cases := []struct {
		name          string
		err           error
		lastAttempt   bool
		expectRetry   bool
		expectedJobId string
	}{
		{"No error", nil, false, false, ""},
		{"Transient with retries left", baseErr, false, true, ""},
		{"Transient on last attempt", baseErr, true, false, "job-1"},
		{"Permanent", jobfail.Permanent(baseErr), false, false, "job-1"},
		{"For other job", jobfail.Permanent(jobfail.For(jobs.Msg{Id: "parent-1"}, baseErr)), false, false, "parent-1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var published []jobs.Msg
			queue := &mocks.MockQueue{PublishFn: func(ctx context.Context, subject string, p any) error {
				assert.Equal(t, string(jobs.SubjectFailed), subject)
				published = append(published, p.(jobs.Msg))
				return nil
			}}
			delivery := ports.Delivery{Payload: payload, LastAttempt: tc.lastAttempt}

			// Act
			got := jobfail.Report(context.Background(), queue, delivery, tc.err)

			// Assert
			if tc.expectRetry {
				assert.ErrorIs(t, got, baseErr)
			} else {
				assert.NoError(t, got)
			}
			if tc.expectedJobId == "" {
				assert.Empty(t, published)
				return
			}
			require.Len(t, published, 1)
			assert.Equal(t, tc.expectedJobId, published[0].Id)
			assert.Equal(t, "boom", published[0].ErrorMsg)
		})
	}
}

func TestReport_PublishFails_ReturnsErrorToRetry(t *testing.T) {
	t.Parallel()

	// Arrange
	pubErr := errors.New("nats down")
	queue := &mocks.MockQueue{PublishFn: func(ctx context.Context, subject string, p any) error { return pubErr }}

	// Act
	err := jobfail.Report(context.Background(), queue, ports.Delivery{}, jobfail.Permanent(errors.New("boom")))

	// Assert
	assert.ErrorIs(t, err, pubErr)
}

package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/queue"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	platformqueue "github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNatsQueue_PublishAndSubscribe_ValidMessage_HandlesEvent(t *testing.T) {
	t.Parallel()

	// Arrange
	srv, err := platformqueue.StartEmbeddedNATS(t.TempDir())
	require.NoError(t, err)
	defer srv.Shutdown()
	nc, err := srv.Connect()
	require.NoError(t, err)
	defer nc.Close()

	q, err := queue.NewNatsQueue(nc)
	require.NoError(t, err)
	defer q.Close()

	var subject jobs.Subject = "test.event.subject"
	payload := []byte("hello world")
	received := make(chan string, 1)

	// Act
	err = q.Subscribe(subject, func(ctx context.Context, subject string, data []byte) error {
		received <- string(data)
		return nil
	})
	require.NoError(t, err)

	err = q.Publish(context.Background(), subject, payload)
	require.NoError(t, err)

	// Assert
	select {
	case msg := <-received:
		assert.Equal(t, "hello world", msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for the message")
	}
}

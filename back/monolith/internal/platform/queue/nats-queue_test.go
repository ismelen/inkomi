package queue_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// connect starts an embedded server with a fresh store, so streams left by other runs are not
// redelivered, and returns a connection to it.
func connect(t *testing.T, storeDir string) *nats.Conn {
	t.Helper()

	srv, err := queue.StartEmbeddedNATS(storeDir)
	require.NoError(t, err)
	t.Cleanup(srv.Shutdown)

	nc, err := srv.Connect()
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	return nc
}

func TestNatsQueue_PublishAndSubscribe_ValidMessage_HandlesEvent(t *testing.T) {
	t.Parallel()

	// Arrange
	q, err := queue.NewNatsQueue(connect(t, t.TempDir()), time.Minute, 3)
	require.NoError(t, err)
	defer q.Close()

	subject := "test.event.subject"
	payload := map[string]string{"msg": "hello world"}
	received := make(chan string, 1)

	// Act
	err = q.Subscribe(subject, func(ctx context.Context, delivery jobrun.Delivery) error {
		received <- string(delivery.Payload)
		return nil
	})
	require.NoError(t, err)

	err = q.Publish(context.Background(), subject, payload)
	require.NoError(t, err)

	// Assert
	select {
	case msg := <-received:
		assert.JSONEq(t, `{"msg":"hello world"}`, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for the message")
	}
}

func TestNatsQueue_Subscribe_FirstDelivery_ReportsLastAttempt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		maxDeliver  int
		lastAttempt bool
	}{
		{"Retries left", 3, false},
		{"No retries left", 1, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			q, err := queue.NewNatsQueue(connect(t, t.TempDir()), time.Minute, tc.maxDeliver)
			require.NoError(t, err)
			defer q.Close()

			received := make(chan jobrun.Delivery, 1)
			require.NoError(t, q.Subscribe("test.attempt", func(ctx context.Context, delivery jobrun.Delivery) error {
				received <- delivery
				return nil
			}))

			// Act
			require.NoError(t, q.Publish(context.Background(), "test.attempt", map[string]string{"id": "1"}))

			// Assert
			select {
			case delivery := <-received:
				assert.Equal(t, "test.attempt", delivery.Subject)
				assert.Equal(t, tc.lastAttempt, delivery.LastAttempt)
			case <-time.After(2 * time.Second):
				t.Fatal("timeout waiting for the message")
			}
		})
	}
}

func TestEmbeddedNATS_Restart_PendingMessage_IsRedelivered(t *testing.T) {
	t.Parallel()

	// Arrange
	storeDir := t.TempDir()
	srv, err := queue.StartEmbeddedNATS(storeDir)
	require.NoError(t, err)
	nc, err := srv.Connect()
	require.NoError(t, err)
	q, err := queue.NewNatsQueue(nc, time.Minute, 3)
	require.NoError(t, err)
	require.NoError(t, q.Publish(context.Background(), "test.persist", map[string]string{"id": "1"}))
	nc.Close()
	srv.Shutdown()

	// Act
	restarted, err := queue.NewNatsQueue(connect(t, storeDir), time.Minute, 3)
	require.NoError(t, err)
	defer restarted.Close()
	received := make(chan string, 1)
	require.NoError(t, restarted.Subscribe("test.persist", func(ctx context.Context, delivery jobrun.Delivery) error {
		received <- string(delivery.Payload)
		return nil
	}))

	// Assert
	select {
	case msg := <-received:
		assert.JSONEq(t, `{"id":"1"}`, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for the persisted message")
	}
}

func TestNatsQueue_SubscribeWith_Concurrency_RunsJobsAtOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	q, err := queue.NewNatsQueue(connect(t, t.TempDir()), time.Minute, 3)
	require.NoError(t, err)
	defer q.Close()
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	require.NoError(t, q.SubscribeWith("test.parallel", queue.SubscribeOptions{Concurrency: 2}, func(ctx context.Context, d jobrun.Delivery) error {
		started <- struct{}{}
		<-release
		return nil
	}))

	// Act
	for i := range 2 {
		require.NoError(t, q.Publish(context.Background(), "test.parallel", map[string]int{"n": i}))
	}

	// Assert
	for range 2 {
		select {
		case <-started:
		case <-time.After(2 * time.Second):
			t.Fatal("the second job didn't start while the first one ran")
		}
	}
	close(release)
}

func TestNatsQueue_SubscribeWith_NotReady_WaitsToStart(t *testing.T) {
	t.Parallel()

	// Arrange
	q, err := queue.NewNatsQueue(connect(t, t.TempDir()), time.Minute, 3)
	require.NoError(t, err)
	defer q.Close()
	var ready atomic.Bool
	started := make(chan struct{}, 1)
	require.NoError(t, q.SubscribeWith("test.gate", queue.SubscribeOptions{Concurrency: 1, Ready: ready.Load}, func(ctx context.Context, d jobrun.Delivery) error {
		started <- struct{}{}
		return nil
	}))

	// Act
	require.NoError(t, q.Publish(context.Background(), "test.gate", map[string]int{"n": 1}))

	// Assert
	select {
	case <-started:
		t.Fatal("the job started before the gate was ready")
	case <-time.After(500 * time.Millisecond):
	}
	ready.Store(true)
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("the job didn't start once the gate was ready")
	}
}

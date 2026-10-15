package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runJetStreamServer starts an embedded NATS server with JetStream enabled on a random port
// and a fresh store, so streams left by other runs are not redelivered.
func runJetStreamServer(t *testing.T) *server.Server {
	t.Helper()

	opts := test.DefaultTestOptions
	opts.Port = -1 // random free port
	opts.JetStream = true
	opts.StoreDir = t.TempDir()
	return test.RunServer(&opts)
}

func TestNatsQueue_NewNatsQueue_InvalidURL_ReturnsError(t *testing.T) {
	// Arrange
	url := "invalid-url://localhost:99999"

	// Act
	q, err := queue.NewNatsQueue(url, time.Minute, 3)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, q)
}

func TestNatsQueue_PublishAndSubscribe_ValidMessage_HandlesEvent(t *testing.T) {
	// Arrange
	// embedded NATS server
	s := runJetStreamServer(t)
	defer s.Shutdown()

	q, err := queue.NewNatsQueue(s.ClientURL(), time.Minute, 3)
	assert.NoError(t, err)
	defer q.Close()

	subject := "test.event.subject"
	payload := map[string]string{"msg": "hello world"}
	received := make(chan string, 1)

	// Act
	err = q.Subscribe(subject, func(ctx context.Context, delivery jobrun.Delivery) error {
		received <- string(delivery.Payload)
		return nil
	})
	assert.NoError(t, err)

	err = q.Publish(context.Background(), subject, payload)
	assert.NoError(t, err)

	// Assert
	select {
	case msg := <-received:
		assert.JSONEq(t, `{"msg":"hello world"}`, msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for the message")
	}
}

func TestNatsQueue_Subscribe_FirstDelivery_ReportsLastAttempt(t *testing.T) {
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
			// Arrange
			s := runJetStreamServer(t)
			defer s.Shutdown()

			q, err := queue.NewNatsQueue(s.ClientURL(), time.Minute, tc.maxDeliver)
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

package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/queue"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
)

// runJetStreamServer starts an embedded NATS server with JetStream enabled on a random port.
func runJetStreamServer() *server.Server {
	opts := test.DefaultTestOptions
	opts.Port = -1 // Asigna un puerto aleatorio disponible
	opts.JetStream = true
	return test.RunServer(&opts)
}

func TestNatsQueue_NewNatsQueue_InvalidURL_ReturnsError(t *testing.T) {
	// Arrange
	url := "invalid-url://localhost:99999"

	// Act
	q, err := queue.NewNatsQueue(url)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, q)
}

func TestNatsQueue_PublishAndSubscribe_ValidMessage_HandlesEvent(t *testing.T) {
	// Arrange
	// Levantamos el servidor NATS de prueba en memoria
	s := runJetStreamServer()
	defer s.Shutdown()

	// Conectamos nuestra cola al servidor en memoria
	q, err := queue.NewNatsQueue(s.ClientURL())
	assert.NoError(t, err)
	defer q.Close()

	subject := "test.event.subject"
	payload := []byte("hello world")
	received := make(chan string, 1)

	// Act
	err = q.Subscribe(subject, func(ctx context.Context, subject string, data []byte) error {
		received <- string(data)
		return nil
	})
	assert.NoError(t, err)

	err = q.Publish(context.Background(), subject, payload)
	assert.NoError(t, err)

	// Assert
	select {
	case msg := <-received:
		assert.Equal(t, "hello world", msg)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout esperando el mensaje")
	}
}

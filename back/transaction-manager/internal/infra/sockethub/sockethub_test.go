package sockethub_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/sockethub"
	"github.com/stretchr/testify/assert"
)

func TestSocketHub_Register_ValidConnection_RegistersClient(t *testing.T) {
	// Arrange
	hub := sockethub.NewSocketHub()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := hub.Register(1, w, r)
		assert.NoError(t, err)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Act
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	defer conn.Close()
}

func TestSocketHub_Send_ExistingClient_SendsMessage(t *testing.T) {
	// Arrange
	hub := sockethub.NewSocketHub()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = hub.Register(2, w, r)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// Wait for client to be registered in hub
	time.Sleep(100 * time.Millisecond)

	type testMessage struct {
		Text string `json:"text"`
	}
	msg := testMessage{Text: "hello"}

	// Act
	err = hub.Send(2, msg)

	// Assert
	assert.NoError(t, err)

	var received testMessage
	err = conn.ReadJSON(&received)
	assert.NoError(t, err)
	assert.Equal(t, "hello", received.Text)
}

func TestSocketHub_Send_NonExistingClient_ReturnsError(t *testing.T) {
	// Arrange
	hub := sockethub.NewSocketHub()
	type testMessage struct {
		Text string `json:"text"`
	}

	// Act
	err := hub.Send(999, testMessage{Text: "miss"})

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cliente 999 no encontrado")
}


package sockethub

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ismelen/inkomi/back/microservices/common/httpauth"
	"github.com/ismelen/inkomi/back/microservices/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/microservices/transaction-manager/internal/domain/ports"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Browsers send the token as the subprotocol list "bearer, <token>" and need it echoed.
	Subprotocols: []string{httpauth.WebSocketProtocol},
	CheckOrigin: func(r *http.Request) bool {
		// the token, not the origin, authenticates the connection
		return true
	},
}

type SocketHub struct {
	mu      sync.RWMutex
	clients map[int]*client
}

func NewSocketHub() *SocketHub {
	return &SocketHub{
		clients: make(map[int]*client),
	}
}

var _ ports.ISocketHub = NewSocketHub()

func (h *SocketHub) Register(id int, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws: user %d: upgrade: %v", id, err)
		return err
	}

	client := &client{
		Id:   id,
		Conn: conn,
	}

	h.mu.Lock()
	if existing, ok := h.clients[id]; ok {
		existing.Conn.Close()
	}
	h.clients[id] = client
	h.mu.Unlock()
	log.Printf("ws: user %d connected", id)

	go h.readPump(client)
	return nil
}

func (h *SocketHub) Send(id int, v any) error {
	h.mu.RLock()
	client, ok := h.clients[id]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("%w: user %d", models.ErrClientNotConnected, id)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	return client.Conn.WriteJSON(v)
}

func (h *SocketHub) readPump(c *client) {
	defer func() {
		h.unregister(c.Id)
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("ws: user %d: %v", c.Id, err)
			}
			log.Printf("ws: user %d disconnected", c.Id)
			break
		}
	}
}

func (h *SocketHub) unregister(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[id]; ok {
		delete(h.clients, id)
		c.Conn.Close()
	}
}

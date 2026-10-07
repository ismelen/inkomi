package sockethub

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite cualquier origen por defecto
	},
}

type SocketHub struct {
	mu      sync.RWMutex
	clients map[string]*client
}

func NewSocketHub() *SocketHub {
	return &SocketHub{
		clients: make(map[string]*client),
	}
}

var _ domain.SocketHub = NewSocketHub()

func (h *SocketHub) Register(id string, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error actualizando a websocket para el cliente %s: %v", id, err)
		return err
	}

	client := &client{
		ID:   id,
		Conn: conn,
	}

	h.mu.Lock()
	if existing, ok := h.clients[id]; ok {
		existing.Conn.Close()
	}
	h.clients[id] = client
	h.mu.Unlock()

	go h.readPump(client)
	return nil
}

func (h *SocketHub) Send(id string, v any) error {
	h.mu.RLock()
	client, ok := h.clients[id]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cliente %s no encontrado", id)
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	return client.Conn.WriteJSON(v)
}

func (h *SocketHub) readPump(c *client) {
	defer func() {
		h.unregister(c.ID)
		c.Conn.Close()
	}()

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error en conexión websocket del cliente %s: %v", c.ID, err)
			}
			break
		}
	}
}

func (h *SocketHub) unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[id]; ok {
		delete(h.clients, id)
		c.Conn.Close()
	}
}

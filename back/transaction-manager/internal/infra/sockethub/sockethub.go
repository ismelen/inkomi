package sockethub

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Browsers send the token as the subprotocol list "bearer, <token>" and need it echoed.
	Subprotocols: []string{httpauth.WebSocketProtocol},
	CheckOrigin: func(r *http.Request) bool {
		return true // Permite cualquier origen por defecto
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
		log.Printf("Error actualizando a websocket para el cliente %d: %v", id, err)
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

	go h.readPump(client)
	return nil
}

func (h *SocketHub) Send(id int, v any) error {
	h.mu.RLock()
	client, ok := h.clients[id]
	h.mu.RUnlock()

	if !ok {
		return fmt.Errorf("cliente %d no encontrado", id)
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error en conexión websocket del cliente %d: %v", c.Id, err)
			}
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

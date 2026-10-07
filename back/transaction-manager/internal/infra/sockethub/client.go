package sockethub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type client struct {
	ID   string
	Conn *websocket.Conn
	mu   sync.Mutex // Protege las escrituras concurrentes en la conexión
}

package queue

import (
	"fmt"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const embeddedStartTimeout = 10 * time.Second

// EmbeddedNATS is a JetStream server running inside the process. It doesn't listen on any port:
// clients connect in-process with Connect.
type EmbeddedNATS struct {
	srv *server.Server
}

// StartEmbeddedNATS starts the server keeping the JetStream streams in storeDir, so pending jobs
// survive a restart.
func StartEmbeddedNATS(storeDir string) (*EmbeddedNATS, error) {
	srv, err := server.NewServer(&server.Options{
		ServerName: "inkomi-monolith",
		JetStream:  true,
		StoreDir:   storeDir,
		DontListen: true,
		NoSigs:     true,
		NoLog:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("create embedded nats: %w", err)
	}

	go srv.Start()
	if !srv.ReadyForConnections(embeddedStartTimeout) {
		srv.Shutdown()
		return nil, fmt.Errorf("embedded nats not ready after %s", embeddedStartTimeout)
	}
	return &EmbeddedNATS{srv: srv}, nil
}

func (e *EmbeddedNATS) Connect() (*nats.Conn, error) {
	nc, err := nats.Connect("", nats.InProcessServer(e.srv))
	if err != nil {
		return nil, fmt.Errorf("connect to embedded nats: %w", err)
	}
	return nc, nil
}

func (e *EmbeddedNATS) Shutdown() {
	e.srv.Shutdown()
	e.srv.WaitForShutdown()
}

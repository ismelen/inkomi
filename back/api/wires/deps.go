// Package wires builds every module and registers its routes and job subscribers; main.go is the
// only caller.
package wires

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"

	"github.com/ismelen/inkomi/back/api/internal/modules/epub/infra/store"
	"github.com/ismelen/inkomi/back/api/internal/platform/cloud"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobcancel"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/platform/queue"
	"github.com/ismelen/inkomi/back/api/internal/platform/storage"
)

// TimeoutMiddleware bounds every route except the WebSocket, which is long-lived, and the preview,
// which has its own limit.
var TimeoutMiddleware = middleware.Timeout(60 * time.Second)

// Deps are the shared pieces every module is wired with.
type Deps struct {
	Ctx            context.Context
	NC             *nats.Conn
	Queue          *queue.NatsQueue
	Cloud          cloud.ICloudStorage
	AuthMiddleware func(http.Handler) http.Handler
	// Local is the server's disk storage, nil when the files live in R2.
	Local  *storage.LocalStorage
	Signer *storage.URLSigner
	// Ready holds back heavy jobs while the free memory or disk is below its reserve.
	Ready func() bool
	// MangaConcurrency is how many manga jobs run at once.
	MangaConcurrency int
	// Cancels skips and interrupts the jobs of cancelled uploads.
	Cancels   *jobcancel.Registry
	JoinStore *store.SQLiteJoinStore
	// DiscoverDB holds the wishlists and the recommendations (discover.db).
	DiscoverDB  *sql.DB
	OpenLibrary *openlibrary.Client
}

// CloseLogged adapts a Close method to a closer that logs its error.
func CloseLogged(name string, closeFn func() error) func() {
	return func() {
		if err := closeFn(); err != nil {
			log.Printf("close %s: %v", name, err)
		}
	}
}

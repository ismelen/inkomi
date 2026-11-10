package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/store"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/cloud"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httpauth"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httplog"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobcancel"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/resources"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/sqlitedb"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/storage"
	"github.com/ismelen/inkomi/back/monolith/migrations"
)

// timeoutMiddleware bounds every route except the WebSocket, which is long-lived, and the preview,
// which has its own limit.
var timeoutMiddleware = middleware.Timeout(60 * time.Second)

// deps are the shared pieces every module is wired with.
type deps struct {
	ctx            context.Context
	nc             *nats.Conn
	queue          *queue.NatsQueue
	cloud          cloud.ICloudStorage
	authMiddleware func(http.Handler) http.Handler
	// local is the server's disk storage, nil when the files live in R2.
	local  *storage.LocalStorage
	signer *storage.URLSigner
	// ready holds back heavy jobs while the free memory or disk is below its reserve.
	ready func() bool
	// mangaConcurrency is how many manga jobs run at once.
	mangaConcurrency int
	// cancels skips and interrupts the jobs of cancelled uploads.
	cancels   *jobcancel.Registry
	joinStore *store.SQLiteJoinStore
}

func main() {
	config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := os.MkdirAll(config.Env.DataDir, 0o750); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	natsServer, err := queue.StartEmbeddedNATS(filepath.Join(config.Env.DataDir, "nats"))
	if err != nil {
		log.Fatal(err)
	}
	defer natsServer.Shutdown()
	nc, err := natsServer.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	jobQueue, err := queue.NewNatsQueue(nc, config.Env.NATSAckWait, config.Env.NATSMaxDeliver)
	if err != nil {
		log.Fatal(err)
	}
	defer jobQueue.Close()

	authMiddleware, err := httpauth.NewAuthMiddleware(config.Env.JwtPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	// jobs.db holds the join groups and the cancelled jobs
	jobsDB, err := sqlitedb.Open(filepath.Join(config.Env.DataDir, "jobs.db"), migrations.Joins)
	if err != nil {
		log.Fatal(err)
	}
	defer closeLogged("jobs.db", jobsDB.Close)()
	cancels, err := jobcancel.NewRegistry(jobsDB, time.Now)
	if err != nil {
		log.Fatal(err)
	}
	cancels.Start(ctx, 6*time.Hour)

	memory := resources.NewMemory()
	d := &deps{
		cancels:          cancels,
		joinStore:        store.NewSQLiteJoinStore(jobsDB, config.Env.JoinTTL, config.Env.JoinLockTTL, time.Now),
		ctx:              ctx,
		nc:               nc,
		queue:            jobQueue,
		authMiddleware:   authMiddleware,
		mangaConcurrency: mangaConcurrency(memory),
	}

	router := chi.NewRouter()
	router.Use(httplog.AccessLog("/healthz"), middleware.Recoverer)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// outside the timeout group: an upload from a phone can take minutes
	wireStorage(d, router)
	memoryGate := resources.NewMemoryGate(memory, config.Env.MemoryReserve)
	d.ready = func() bool { return memoryGate.Ready() && diskReady(d) }

	// closers release the module databases once the server stopped, in reverse order
	var closers []func()
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}()

	// the WebSocket stays outside the timeout group and the preview has its own limits
	closers = append(closers, wireTransactions(d, router))
	wireManga(d, router)

	router.Group(func(r chi.Router) {
		r.Use(timeoutMiddleware)
		closeAuth, auth := wireAuth(d, r)
		closers = append(closers, closeAuth)
		wireEpub(d)
		wireLibrary(d, r)
		wireRelease(r)
		wirePush(d, auth)
	})

	server := &http.Server{
		Addr:              ":" + config.Env.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    64 << 10,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("Starting at port %s", config.Env.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// closeLogged adapts a Close method to a closer that logs its error.
func closeLogged(name string, closeFn func() error) func() {
	return func() {
		if err := closeFn(); err != nil {
			log.Printf("close %s: %v", name, err)
		}
	}
}

// mangaConcurrency is MANGA_CONCURRENCY or, when unset, min(CPUs, free memory / memory per job).
func mangaConcurrency(memory *resources.Memory) int {
	if config.Env.MangaConcurrency > 0 {
		return config.Env.MangaConcurrency
	}
	available, ok := memory.Available()
	return resources.Concurrency(runtime.NumCPU(), available, ok, config.Env.MangaJobMemory)
}

// diskReady reports whether the local storage has more free space than the reserve; it is
// always true with R2 or when the free space can't be read.
func diskReady(d *deps) bool {
	if d.local == nil {
		return true
	}
	free, ok := storage.FreeBytes(d.local.Root())
	return !ok || free > config.Env.DiskReserve
}

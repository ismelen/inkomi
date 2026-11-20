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

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/epub/infra/store"
	"github.com/ismelen/inkomi/back/api/internal/platform/httpauth"
	"github.com/ismelen/inkomi/back/api/internal/platform/httplog"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobcancel"
	"github.com/ismelen/inkomi/back/api/internal/platform/queue"
	"github.com/ismelen/inkomi/back/api/internal/platform/resources"
	"github.com/ismelen/inkomi/back/api/internal/platform/sqlitedb"
	"github.com/ismelen/inkomi/back/api/internal/platform/storage"
	"github.com/ismelen/inkomi/back/api/migrations"
	"github.com/ismelen/inkomi/back/api/wires"
)

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
	defer wires.CloseLogged("jobs.db", jobsDB.Close)()
	cancels, err := jobcancel.NewRegistry(jobsDB, time.Now)
	if err != nil {
		log.Fatal(err)
	}
	cancels.Start(ctx, 6*time.Hour)

	memory := resources.NewMemory()
	d := &wires.Deps{
		Cancels:          cancels,
		JoinStore:        store.NewSQLiteJoinStore(jobsDB, config.Env.JoinTTL, config.Env.JoinLockTTL, time.Now),
		Ctx:              ctx,
		NC:               nc,
		Queue:            jobQueue,
		AuthMiddleware:   authMiddleware,
		MangaConcurrency: mangaConcurrency(memory),
	}

	router := chi.NewRouter()
	router.Use(httplog.AccessLog("/healthz"), middleware.Recoverer)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	// outside the timeout group: an upload from a phone can take minutes
	wires.Storage(d, router)
	memoryGate := resources.NewMemoryGate(memory, config.Env.MemoryReserve)
	d.Ready = func() bool { return memoryGate.Ready() && diskReady(d) }

	// closers release the module databases once the server stopped, in reverse order
	var closers []func()
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			closers[i]()
		}
	}()

	// the WebSocket stays outside the timeout group and the preview has its own limits
	closers = append(closers, wires.Transactions(d, router))
	wires.Manga(d, router)

	router.Group(func(r chi.Router) {
		r.Use(wires.TimeoutMiddleware)
		closeAuth, auth := wires.Auth(d, r)
		closers = append(closers, closeAuth)
		wires.Epub(d)
		wires.Library(d, r)
		wires.Release(r)
		wires.Push(d, auth)
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
func diskReady(d *wires.Deps) bool {
	if d.Local == nil {
		return true
	}
	free, ok := storage.FreeBytes(d.Local.Root())
	return !ok || free > config.Env.DiskReserve
}

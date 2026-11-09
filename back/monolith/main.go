package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/nats-io/nats.go"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/cloud"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httpauth"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httplog"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
)

// timeoutMiddleware bounds every route except the WebSocket, which is long-lived, and the preview,
// which has its own limit.
var timeoutMiddleware = middleware.Timeout(60 * time.Second)

// deps are the shared pieces every module is wired with.
type deps struct {
	ctx            context.Context
	nc             *nats.Conn
	queue          *queue.NatsQueue
	cloud          *cloud.CloudflareR2Storage
	authMiddleware func(http.Handler) http.Handler
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

	cloudStorage, err := cloud.NewCloudflareR2Storage(
		ctx,
		config.Env.R2AccountId,
		config.Env.R2AccessKeyId,
		config.Env.R2SecretAccessKey,
		config.Env.R2BucketName,
	)
	if err != nil {
		log.Fatal(err)
	}

	authMiddleware, err := httpauth.NewAuthMiddleware(config.Env.JwtPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	d := &deps{ctx: ctx, nc: nc, queue: jobQueue, cloud: cloudStorage, authMiddleware: authMiddleware}

	router := chi.NewRouter()
	router.Use(httplog.AccessLog("/healthz"), middleware.Recoverer)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

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
		closers = append(closers, closeAuth, wireEpub(d))
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

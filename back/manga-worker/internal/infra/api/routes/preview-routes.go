package routes

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/middlewares"
)

const (
	maxQueryLength  = 2048
	concurrencyWait = 2 * time.Second
)

type PreviewLimits struct {
	RateLimiter   *middlewares.IPRateLimiter
	MaxConcurrent int
	Timeout       time.Duration
}

// SetupPreviewRoutes registers the public preview. Only /processed is expensive, so it's the one
// behind the rate and concurrency limits; the rate limit goes first so rejected requests never
// take a slot.
func SetupPreviewRoutes(api chi.Router, handler *handlers.PreviewHandler, limits PreviewLimits) {
	api.Route("/preview", func(r chi.Router) {
		r.Use(middlewares.MaxQueryLength(maxQueryLength))
		r.Get("/", handler.HandlePage)
		r.Get("/original", handler.HandleOriginal)
		r.With(
			limits.RateLimiter.Middleware,
			middlewares.ConcurrencyLimit(limits.MaxConcurrent, concurrencyWait),
			middlewares.Timeout(limits.Timeout),
		).Get("/processed", handler.HandleProcessed)
	})
}

package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
)

// SetupInternalRoutes registers the routes for other services, protected by serviceMiddleware.
func SetupInternalRoutes(api chi.Router, handler *handlers.InternalHandler, serviceMiddleware func(http.Handler) http.Handler) {
	api.Route("/internal/users/{userId}", func(r chi.Router) {
		r.Use(serviceMiddleware)
		r.Get("/dropbox", requtil.Wrap(handler.HandleGetDropbox))
		r.Get("/fcm-token", requtil.Wrap(handler.HandleGetFCMToken))
	})
}

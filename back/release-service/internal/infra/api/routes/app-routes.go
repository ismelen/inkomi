package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/requtil"
)

// SetupAppRoutes registers the public routes of the app releases: no token, the update check
// runs before sign in.
func SetupAppRoutes(api chi.Router, handler *handlers.AppHandler) {
	api.Route("/app", func(r chi.Router) {
		r.Get("/latest", requtil.Wrap(handler.HandleLatest))
		r.Get("/download", requtil.Wrap(handler.HandleDownload))
	})
}

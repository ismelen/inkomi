package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
)

// SetupUserRoutes registers the routes of the signed in user, protected by authMiddleware.
func SetupUserRoutes(api chi.Router, handler *handlers.UsersHandler, authMiddleware func(http.Handler) http.Handler) {
	api.Route("/users/me", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Get("/", requtil.Wrap(handler.HandleGetProfile))
		r.Put("/fcm-token", requtil.Wrap(handler.HandleUpdateFCMToken))
		r.Put("/dropbox", requtil.Wrap(handler.HandleLinkDropbox))
		r.Delete("/dropbox", requtil.Wrap(handler.HandleUnlinkDropbox))
		r.Put("/dropbox/folder", requtil.Wrap(handler.HandleUpdateDropboxFolder))
	})
}

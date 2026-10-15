package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/api/requtil"
)

func SetupAuthRoutes(api chi.Router, handler *handlers.AuthHandler) {
	api.Route("/auth", func(r chi.Router) {
		r.Post("/register", requtil.Wrap(handler.HandleRegister))
		r.Post("/login", requtil.Wrap(handler.HandleLogin))
		r.Post("/google", requtil.Wrap(handler.HandleGoogleLogin))
		r.Post("/dropbox", requtil.Wrap(handler.HandleDropboxLogin))
		r.Post("/refresh", requtil.Wrap(handler.HandleRefresh))
		r.Post("/logout", requtil.Wrap(handler.HandleLogout))
	})
}

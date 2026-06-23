package routes

import (
	"ismelen/inkomi/internal/infra/api/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupLibgenRoutes(api *chi.Mux, handler *handlers.LibgenHandler) {
	r := chi.NewRouter()
	api.Mount("/books", r)

	r.Get("/search", Wrap(handler.HandleSearchBook))
	r.Post("/download", Wrap(handler.HandleDownloadBook))
}

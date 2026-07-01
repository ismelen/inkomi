package routes

import (
	"ismelen/inkomi/internal/infra/api/handlers"
	"ismelen/inkomi/internal/infra/api/requtil"

	"github.com/go-chi/chi/v5"
)

func SetupLibgenRoutes(api *chi.Mux, handler *handlers.LibgenHandler) {
	r := chi.NewRouter()
	api.Mount("/books", r)

	r.Get("/search", requtil.Wrap(handler.HandleSearchBook))
	r.Get("/download/{md5}", requtil.Wrap(handler.HandleDownloadBook))
}

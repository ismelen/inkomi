package routes

import (
	"ismelen/ermc/v2/infra/api/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupConvertRoutes(api *chi.Mux, handler *handlers.ConvertHandler) {
	api.Post("/convert", Wrap(handler.Convert))
	api.Get("/status/{id}", Wrap(handler.CheckStatus))
	api.Get("/download/{id}", Wrap(handler.Download))
	api.Put("/cancel/{id}", Wrap(handler.Cancel))
}

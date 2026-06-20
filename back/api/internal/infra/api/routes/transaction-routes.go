package routes

import (
	"ismelen/inkomi/internal/infra/api/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupConvertRoutes(api *chi.Mux, handler *handlers.TransactionHandler) {
	r := chi.NewRouter()
	api.Mount("/transaction", r)

	r.Post("/convert", Wrap(handler.HandleConvert))
	r.Get("/status/{id}", Wrap(handler.HandleCheckStatus))
	r.Get("/download/{id}", Wrap(handler.HandleDownload))
	r.Put("/cancel/{id}", Wrap(handler.HandleCancel))
}

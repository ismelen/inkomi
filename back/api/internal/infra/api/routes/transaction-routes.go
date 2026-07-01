package routes

import (
	"ismelen/inkomi/internal/infra/api/handlers"
	"ismelen/inkomi/internal/infra/api/requtil"

	"github.com/go-chi/chi/v5"
)

func SetupConvertRoutes(api *chi.Mux, handler *handlers.TransactionHandler) {
	r := chi.NewRouter()
	api.Mount("/transaction", r)

	r.Post("/convert", requtil.Wrap(handler.HandleConvert))
	r.Get("/status/{id}", requtil.Wrap(handler.HandleCheckStatus))
	r.Get("/download/{id}", requtil.Wrap(handler.HandleDownload))
	r.Put("/cancel/{id}", requtil.Wrap(handler.HandleCancel))
}

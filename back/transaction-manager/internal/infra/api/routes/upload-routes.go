package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/dtos"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/handlers"
	requtil "github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/requtils"
)

func SetupUploadRoutes(api *chi.Mux, handler *handlers.UploadsHandler) {
	r := chi.NewRouter()
	api.Mount("/uploads", r)

	r.Get("/", requtil.Wrap(handler.HandleList))
	r.Post("/request", requtil.Wrap[[]dtos.SourceDTO](handler.HandleNewRequest))
	r.Patch("/done/{sourceId}", requtil.Wrap(handler.HandleOnComplete))
	r.Get("/{sourceId}/download", requtil.Wrap(handler.HandleDownload))
}

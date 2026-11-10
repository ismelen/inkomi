package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/dtos"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/handlers"
	requtil "github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/requtils"
)

func SetupUploadRoutes(api chi.Router, handler *handlers.UploadsHandler) {
	r := chi.NewRouter()
	api.Mount("/uploads", r)

	r.Get("/", requtil.Wrap(handler.HandleList))
	r.Post("/request", requtil.Wrap[[]dtos.SourceDTO](handler.HandleNewRequest))
	r.Patch("/done/{sourceId}", requtil.Wrap(handler.HandleOnComplete))
	r.Get("/{sourceId}/download", requtil.Wrap(handler.HandleDownload))
	r.Post("/{sourceId}/retry", requtil.Wrap(handler.HandleRetry))
	r.Post("/{sourceId}/cancel", requtil.Wrap(handler.HandleCancel))
}

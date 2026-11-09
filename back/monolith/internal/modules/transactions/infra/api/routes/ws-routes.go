package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/handlers"
	requtil "github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/requtils"
)

func SetupWsRoutes(api chi.Router, handler *handlers.WSHandler) {
	r := chi.NewRouter()
	api.Mount("/ws", r)

	r.Get("/", requtil.Wrap[any](handler.HandleConnect))
}

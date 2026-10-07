package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/handlers"
	requtil "github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/requtils"
)

func SetupWsRoutes(api *chi.Mux, handler *handlers.WSHandler) {
	r := chi.NewRouter()
	api.Mount("/ws", r)

	r.Get("/", requtil.Wrap[any](handler.HandleConnect))
}

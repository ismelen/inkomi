package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/microservices/transaction-manager/internal/infra/api/handlers"
)

func SetupWsRoutes(api *chi.Mux, handler *handlers.WSHandler) {
	r := chi.NewRouter()
	api.Mount("/ws", r)

	r.Get("/", handler.HandleConnect)
}

package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/ui/handlers"
)

func SetupWsRoutes(api chi.Router, handler *handlers.WSHandler) {
	r := chi.NewRouter()
	api.Mount("/ws", r)

	r.Get("/", handler.HandleConnect)
}

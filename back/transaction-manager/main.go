package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/middlewares"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/sockethub"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	api := chi.NewRouter()
	api.Use(middlewares.AuthMiddleware)

	sockethub := sockethub.NewSocketHub()

	wsHandler := handlers.NewWSHandler(sockethub)
	routes.SetupWsRoutes(api, wsHandler)

}

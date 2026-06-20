package main

import (
	"ismelen/inkomi/internal/infra/api/handlers"
	"ismelen/inkomi/internal/infra/api/routes"
	"ismelen/inkomi/internal/infra/notifications"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	api := chi.NewRouter()

	api.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
		middleware.RequestSize(250<<20),
		cors.AllowAll().Handler,
	)

	pushNotifier := notifications.FirebasePushNotifier{}
	pushNotifier.Init()

	convertHandler := handlers.NewConvertHandler(&pushNotifier)
	routes.SetupConvertRoutes(api, convertHandler)

	log.Println("Starting at port 3000")
	if err := http.ListenAndServe(":3000", api); err != nil {
		log.Fatal(err)
	}
}

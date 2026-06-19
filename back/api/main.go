package main

import (
	"ismelen/ermc/v2/infra/api/handlers"
	"ismelen/ermc/v2/infra/api/routes"
	"ismelen/ermc/v2/infra/notifications"
	"ismelen/ermc/v2/usecases"
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

	convertUC := usecases.NewConvertMangaUC(&pushNotifier)
	convertHandler := handlers.NewConvertHandler(convertUC)
	routes.SetupConvertRoutes(api, convertHandler)

	if err := http.ListenAndServe(":3000", api); err != nil {
		log.Fatal(err)
	}
}

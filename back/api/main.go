package main

import (
	"ismelen/ermc/v2/infra/api/handlers"
	"ismelen/ermc/v2/infra/api/routes"
	"ismelen/ermc/v2/infra/notifications"
	"ismelen/ermc/v2/usecases"
	"log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	api := chi.NewRouter()

	api.Use(middleware.RequestID)
	api.Use(middleware.Logger)
	api.Use(middleware.Recoverer)
	api.Use(cors.Handler(cors.Options{
    // AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
    AllowedOrigins:   []string{"https://*", "http://*"},
    // AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: false,
    MaxAge:           300, // Maximum value not ignored by any of major browsers
  }))
	// api := echo.New()
	// api.Use(middleware.RequestLogger())
	// api.Use(middleware.CORS())
	// api.Use(middleware.BodyLimit("200M"))

	pushNotifier := notifications.FirebasePushNotifier{}
	pushNotifier.Init()

	convertUC := usecases.NewConvertMangaUC(&pushNotifier)
	convertHandler := handlers.NewConvertHandler(convertUC)
	routes.SetupConvertRoutes(api, convertHandler)

	if err := api.Start(":3000"); err != nil {
		log.Fatal(err)
	}
}

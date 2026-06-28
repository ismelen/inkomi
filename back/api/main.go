package main

import (
	"context"
	"ismelen/inkomi/internal/infra/api/handlers"
	"ismelen/inkomi/internal/infra/api/routes"
	"ismelen/inkomi/internal/infra/libgen"
	"ismelen/inkomi/internal/infra/notifications"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	api := chi.NewRouter()

	api.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
		middleware.RequestSize(250<<20),
		cors.AllowAll().Handler,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	libgenServ := libgen.New()
	libgenServ.StartDiscovery(ctx, 12*time.Hour)

	pushNotifier := notifications.FirebasePushNotifier{}
	pushNotifier.Init()

	convertHandler := handlers.NewConvertHandler(&pushNotifier)
	routes.SetupConvertRoutes(api, convertHandler)

	libgenhandler := handlers.NewLibgenHandler(libgenServ)
	routes.SetupLibgenRoutes(api, libgenhandler)

	log.Println("Starting at port 3000")
	if err := http.ListenAndServe(":3000", api); err != nil {
		log.Fatal(err)
	}
}

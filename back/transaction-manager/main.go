package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/middlewares"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/cloud"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/datasources"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/queue"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/sockethub"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	godotenv.Load()

	api := chi.NewRouter()
	api.Use(middlewares.AuthMiddleware)

	sockethub := sockethub.NewSocketHub()
	queue, err := queue.NewNatsQueue(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	wsHandler := handlers.NewWSHandler(sockethub)
	routes.SetupWsRoutes(api, wsHandler)

	db, err := datasources.NewSQLiteDatasource("transactions.db", "./migrations/trasnactions.sql")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sourceRepo := repositories.NewSQLiteSourceRepository(db.DB)
	configRepo := repositories.NewSQLiteConfigRepository(db.DB)
	cloudStorage := cloud.NewCloudflareR2Storage()

	newUploadRequestUC := usecases.NewNewUploadRequestUC(db, sourceRepo, configRepo, cloudStorage)
	uploadDoneUC := usecases.NewUploadDoneUC(sourceRepo, queue)
	uploadsHandler := handlers.NewUploadsHandler(newUploadRequestUC, uploadDoneUC)
	routes.SetupUploadRoutes(api, uploadsHandler)

	log.Println("Starting at port 3000")
	if err := http.ListenAndServe(":3000", api); err != nil {
		log.Fatal(err)
	}

}

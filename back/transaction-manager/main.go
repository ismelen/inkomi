package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/config"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/cloud"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/datasources"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/queue"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/repositories"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/sockethub"
	forwaredevents "github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/forwardevents"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/newupload"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/uploaddone"
)

func main() {
	config.Load()

	authMiddleware, err := httpauth.NewAuthMiddleware(config.Env.JwtPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	api := chi.NewRouter()
	api.Use(authMiddleware)

	sockethub := sockethub.NewSocketHub()
	queue, err := queue.NewNatsQueue(config.Env.NATSUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	wsHandler := handlers.NewWSHandler(sockethub)
	routes.SetupWsRoutes(api, wsHandler)

	db, err := datasources.NewSQLiteDatasource(config.Env.DBPath, config.Env.SchemaPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sourceRepo := repositories.NewSQLiteSourceRepository(db.DB)
	configRepo := repositories.NewSQLiteConfigRepository(db.DB)

	cloudStorage, err := cloud.NewCloudflareR2Storage(
		config.Env.R2AccountId,
		config.Env.R2AccessKeyId,
		config.Env.R2SecretAccessKey,
		config.Env.R2BucketName,
	)
	if err != nil {
		log.Fatal(err)
	}

	newUploadRequestUC := newupload.NewNewUploadRequestUC(db, sourceRepo, configRepo, cloudStorage)
	uploadDoneUC := uploaddone.NewUploadDoneUC(sourceRepo, cloudStorage, queue)
	uploadsHandler := handlers.NewUploadsHandler(newUploadRequestUC, uploadDoneUC)
	routes.SetupUploadRoutes(api, uploadsHandler)

	forwardEventsUc := forwaredevents.NewForwardEventsUC(sockethub, queue, sourceRepo)
	if err := forwardEventsUc.Execute(); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting at port 3000")
	if err := http.ListenAndServe(":3000", api); err != nil {
		log.Fatal(err)
	}

}

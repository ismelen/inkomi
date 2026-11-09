package main

import (
	"log"
	"path/filepath"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/cloud"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/datasources"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/queue"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/repositories"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/sockethub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/forwardevents"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/getdownload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/listuploads"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/newupload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/uploaddone"
	"github.com/ismelen/inkomi/back/monolith/migrations"
)

func wireTransactions(d *deps, router chi.Router) func() {
	db, err := datasources.NewSQLiteDatasource(filepath.Join(config.Env.DataDir, "transactions.db"), migrations.Transactions)
	if err != nil {
		log.Fatal(err)
	}

	eventQueue, err := queue.NewNatsQueue(d.nc)
	if err != nil {
		log.Fatal(err)
	}

	cloudStorage, err := cloud.NewCloudflareR2Storage(
		config.Env.R2AccountId,
		config.Env.R2AccessKeyId,
		config.Env.R2SecretAccessKey,
		config.Env.R2BucketName,
	)
	if err != nil {
		log.Fatal(err)
	}

	sourceRepo := repositories.NewSQLiteSourceRepository(db.DB)
	configRepo := repositories.NewSQLiteConfigRepository(db.DB)
	hub := sockethub.NewSocketHub()

	uploadsHandler := handlers.NewUploadsHandler(
		newupload.NewNewUploadRequestUC(db, sourceRepo, configRepo, cloudStorage),
		uploaddone.NewUploadDoneUC(sourceRepo, cloudStorage, eventQueue),
		listuploads.NewListUploadsUC(sourceRepo),
		getdownload.NewGetDownloadUC(sourceRepo, cloudStorage),
	)

	if err := forwardevents.NewForwardEventsUC(hub, eventQueue, sourceRepo).Execute(); err != nil {
		log.Fatal(err)
	}

	router.Group(func(r chi.Router) {
		r.Use(d.authMiddleware)
		routes.SetupWsRoutes(r, handlers.NewWSHandler(hub))
		r.Group(func(r chi.Router) {
			r.Use(timeoutMiddleware)
			routes.SetupUploadRoutes(r, uploadsHandler)
		})
	})

	return func() {
		closeLogged("event queue", eventQueue.Close)()
		closeLogged("transactions.db", db.Close)()
	}
}

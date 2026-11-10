package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/cloud"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/datasources"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/queue"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/repositories"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/sockethub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/storagequota"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/canceltransaction"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/forwardevents"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/getdownload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/listuploads"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/newupload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/retryupload"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/uploaddone"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/storage"
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

	cloudStorage, quota := transactionsStorage(d, db)

	sourceRepo := repositories.NewSQLiteSourceRepository(db.DB)
	configRepo := repositories.NewSQLiteConfigRepository(db.DB)
	hub := sockethub.NewSocketHub()

	uploadsHandler := handlers.NewUploadsHandler(
		newupload.NewNewUploadRequestUC(db, sourceRepo, configRepo, cloudStorage, quota, models.StorageEstimateOptions{
			EpubRatio:       config.Env.StorageEpubRatio,
			LibraryEstimate: config.Env.LibraryEstimate,
			Concurrency:     d.mangaConcurrency,
			MaxJoinBlock:    config.Env.JoinMaxBlockSize,
		}),
		uploaddone.NewUploadDoneUC(sourceRepo, cloudStorage, eventQueue),
		listuploads.NewListUploadsUC(sourceRepo),
		getdownload.NewGetDownloadUC(sourceRepo, cloudStorage),
		retryupload.NewRetryUploadUC(sourceRepo, cloudStorage),
		canceltransaction.NewCancelTransactionUC(
			db, sourceRepo, repositories.NewSQLiteStorageReservationRepository(db.DB),
			d.cancels, fileCleaner(d), d.joinStore, hub,
		),
	)

	if err := forwardevents.NewForwardEventsUC(hub, eventQueue, sourceRepo, quota).Execute(); err != nil {
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

// transactionsStorage returns the signed urls and the space quota of the storage in use: urls of
// /files with a disk quota for the local storage, presigned urls without quota for R2.
func transactionsStorage(d *deps, db *datasources.SQLiteDatasource) (ports.ICloudStorage, ports.IStorageQuota) {
	if d.local == nil {
		r2, err := cloud.NewCloudflareR2Storage(
			config.Env.R2AccountId,
			config.Env.R2AccessKeyId,
			config.Env.R2SecretAccessKey,
			config.Env.R2BucketName,
		)
		if err != nil {
			log.Fatal(err)
		}
		return r2, storagequota.NoQuota{}
	}

	quota := storagequota.NewDiskQuota(
		repositories.NewSQLiteStorageReservationRepository(db.DB),
		d.local,
		func() (uint64, bool) { return storage.FreeBytes(d.local.Root()) },
		int64(config.Env.DiskReserve),
		time.Now,
	)
	quota.Start(d.ctx, 5*time.Minute)
	return cloud.NewLocalSignedStorage(d.local, d.signer), quota
}

// noFiles cleans nothing: R2 objects expire through the bucket's lifecycle.
type noFiles struct{}

func (noFiles) DeleteAll(string) error { return nil }

func fileCleaner(d *deps) ports.IFileCleaner {
	if d.local == nil {
		return noFiles{}
	}
	return d.local
}

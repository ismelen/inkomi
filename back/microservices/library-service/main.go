package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/microservices/common/cloud"
	"github.com/ismelen/inkomi/back/microservices/common/httpauth"
	"github.com/ismelen/inkomi/back/microservices/common/jobrun"
	"github.com/ismelen/inkomi/back/microservices/common/jobs"
	"github.com/ismelen/inkomi/back/microservices/common/queue"
	"github.com/ismelen/inkomi/back/microservices/library-service/internal/config"
	"github.com/ismelen/inkomi/back/microservices/library-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/microservices/library-service/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/microservices/library-service/internal/infra/libgen"
	"github.com/ismelen/inkomi/back/microservices/library-service/internal/usecases/downloadbook"
	"github.com/ismelen/inkomi/back/microservices/library-service/internal/usecases/searchbooks"
)

func main() {
	config.Load()
	ctx := context.Background()

	queue, err := queue.NewNatsQueue(config.Env.NATSUrl, config.Env.NATSAckWait, config.Env.NATSMaxDeliver)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	cloudStorage, err := cloud.NewCloudflareR2Storage(
		ctx,
		config.Env.R2AccountId,
		config.Env.R2AccessKeyId,
		config.Env.R2SecretAccessKey,
		config.Env.R2BucketName,
	)
	if err != nil {
		log.Fatal(err)
	}

	slumURL := config.Env.SlumURL
	if slumURL == "" {
		slumURL = libgen.SlumURL
	}
	library := libgen.NewLibgenLibrary(
		libgen.NewSlumMirrorList(slumURL, libgen.FallbackMirrors()),
		config.Env.MirrorRefreshCooldown,
	)
	library.Start(ctx, config.Env.MirrorRefreshInterval)

	downloadBookUC := downloadbook.NewDownloadBookUC(queue, library, cloudStorage)
	if err := queue.Subscribe(string(jobs.SubjectLibrary), jobrun.Handle(queue, downloadBookUC.Execute)); err != nil {
		log.Fatal(err)
	}

	authMiddleware, err := httpauth.NewAuthMiddleware(config.Env.JwtPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	api := chi.NewRouter()
	api.Use(authMiddleware)
	routes.SetupBookRoutes(api, handlers.NewBooksHandler(searchbooks.NewSearchBooksUC(library)))

	log.Printf("Starting at port %s", config.Env.Port)
	if err := http.ListenAndServe(":"+config.Env.Port, api); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/ismelen/inkomi/back/manga-worker/internal/config"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/middlewares"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/archive"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/epub"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/samples"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/vips"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/convertmanga"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/getsample"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/previewpage"
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

	processor := vips.NewBimgPageProcessor(ctx, config.Env.Workers)
	extractors := map[models.SourceFormat]ports.IPageExtractor{
		models.SourceFormatCBZ: archive.NewCBZExtractor(),
		models.SourceFormatPDF: archive.NewPDFExtractor(),
	}
	convertMangaUC := convertmanga.NewConvertMangaUC(
		queue,
		cloudStorage,
		extractors,
		processor,
		epub.NewFixedLayoutBuilder(),
		config.Env.Workers,
	)

	if err := queue.Subscribe(string(jobs.SubjectManga), jobrun.Handle(queue, convertMangaUC.Execute)); err != nil {
		log.Fatal(err)
	}

	sampleSource := samples.NewCloudSampleSource(
		cloudStorage,
		config.Env.PreviewSamplesId,
		map[models.PreviewSample]string{
			models.PreviewSampleColor: config.Env.PreviewColorSample,
			models.PreviewSampleBW:    config.Env.PreviewBWSample,
		},
		config.Env.PreviewMaxSampleSize,
	)
	previewHandler := handlers.NewPreviewHandler(
		previewpage.NewPreviewPageUC(sampleSource, processor),
		getsample.NewGetSampleUC(sampleSource),
	)
	rateLimiter := middlewares.NewIPRateLimiter(
		config.Env.PreviewRatePerMinute,
		config.Env.PreviewBurst,
		config.Env.PreviewTrustProxy,
	)
	rateLimiter.Start(ctx, time.Minute)

	api := chi.NewRouter()
	routes.SetupPreviewRoutes(api, previewHandler, routes.PreviewLimits{
		RateLimiter:   rateLimiter,
		MaxConcurrent: config.Env.PreviewMaxConcurrent,
		Timeout:       config.Env.PreviewTimeout,
	})

	// a public endpoint: slow or oversized clients must not hold connections open
	server := &http.Server{
		Addr:              ":" + config.Env.Port,
		Handler:           api,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      config.Env.PreviewTimeout + 10*time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}

	log.Printf("Starting at port %s", config.Env.Port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

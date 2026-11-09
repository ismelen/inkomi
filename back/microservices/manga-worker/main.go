package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/microservices/common/cloud"
	"github.com/ismelen/inkomi/back/microservices/common/jobrun"
	"github.com/ismelen/inkomi/back/microservices/common/jobs"
	commonqueue "github.com/ismelen/inkomi/back/microservices/common/queue"
	"github.com/ismelen/inkomi/back/microservices/common/resources"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/config"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/api/middlewares"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/archive"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/epub"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/samples"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/infra/vips"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/usecases/convertmanga"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/usecases/getsample"
	"github.com/ismelen/inkomi/back/microservices/manga-worker/internal/usecases/previewpage"
)

func main() {
	config.Load()
	ctx := context.Background()

	queue, err := commonqueue.NewNatsQueue(config.Env.NATSUrl, config.Env.NATSAckWait, config.Env.NATSMaxDeliver)
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

	memory := resources.NewMemory()
	concurrency := config.Env.MangaConcurrency
	if concurrency <= 0 {
		available, ok := memory.Available()
		concurrency = resources.Concurrency(runtime.NumCPU(), available, ok, config.Env.MangaJobMemory)
	}
	log.Printf("manga: %d jobs at once, %d page workers", concurrency, config.Env.Workers)
	opts := commonqueue.SubscribeOptions{
		Concurrency: concurrency,
		Ready:       resources.NewMemoryGate(memory, config.Env.MemoryReserve).Ready,
	}
	if err := queue.SubscribeWith(string(jobs.SubjectManga), opts, jobrun.Handle(queue, convertMangaUC.Execute)); err != nil {
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

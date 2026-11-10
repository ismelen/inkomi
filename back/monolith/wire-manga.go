package main

import (
	"log"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/api/middlewares"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/archive"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/epub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/samples"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/infra/vips"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/usecases/convertmanga"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/usecases/getsample"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/manga/usecases/previewpage"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
)

func wireManga(d *deps, router chi.Router) {
	processor := vips.NewBimgPageProcessor(d.ctx, config.Env.Workers)
	extractors := map[models.SourceFormat]ports.IPageExtractor{
		models.SourceFormatCBZ: archive.NewCBZExtractor(),
		models.SourceFormatPDF: archive.NewPDFExtractor(),
	}
	convertMangaUC := convertmanga.NewConvertMangaUC(
		d.queue,
		d.cloud,
		extractors,
		processor,
		epub.NewFixedLayoutBuilder(),
		config.Env.Workers,
	)
	log.Printf("manga: %d jobs at once, %d page workers", d.mangaConcurrency, config.Env.Workers)
	opts := queue.SubscribeOptions{Concurrency: d.mangaConcurrency, Ready: d.ready}
	if err := d.queue.SubscribeWith(string(jobs.SubjectManga), opts, jobrun.Handle(d.queue, convertMangaUC.Execute, jobrun.WithCancellations(d.cancels))); err != nil {
		log.Fatal(err)
	}

	sampleSource := samples.NewCloudSampleSource(
		d.cloud,
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
	rateLimiter.Start(d.ctx, time.Minute)

	routes.SetupPreviewRoutes(router, previewHandler, routes.PreviewLimits{
		RateLimiter:   rateLimiter,
		MaxConcurrent: config.Env.PreviewMaxConcurrent,
		Timeout:       config.Env.PreviewTimeout,
	})
}

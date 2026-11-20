package wires

import (
	"log"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/infra/archive"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/infra/epub"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/infra/samples"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/infra/vips"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/ui/middlewares"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/usecases/convertmanga"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/usecases/getsample"
	"github.com/ismelen/inkomi/back/api/internal/modules/manga/usecases/previewpage"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/api/internal/platform/queue"
)

func Manga(d *Deps, router chi.Router) {
	processor := vips.NewBimgPageProcessor(d.Ctx, config.Env.Workers)
	extractors := map[models.SourceFormat]ports.IPageExtractor{
		models.SourceFormatCBZ: archive.NewCBZExtractor(),
		models.SourceFormatPDF: archive.NewPDFExtractor(),
	}
	convertMangaUC := convertmanga.NewConvertMangaUC(
		d.Queue,
		d.Cloud,
		extractors,
		processor,
		epub.NewFixedLayoutBuilder(),
		config.Env.Workers,
	)
	log.Printf("manga: %d jobs at once, %d page workers", d.MangaConcurrency, config.Env.Workers)
	opts := queue.SubscribeOptions{Concurrency: d.MangaConcurrency, Ready: d.Ready}
	if err := d.Queue.SubscribeWith(string(jobs.SubjectManga), opts, jobrun.Handle(d.Queue, convertMangaUC.Execute, jobrun.WithCancellations(d.Cancels))); err != nil {
		log.Fatal(err)
	}

	sampleSource := samples.NewCloudSampleSource(
		d.Cloud,
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
	rateLimiter.Start(d.Ctx, time.Minute)

	routes.SetupPreviewRoutes(router, previewHandler, routes.PreviewLimits{
		RateLimiter:   rateLimiter,
		MaxConcurrent: config.Env.PreviewMaxConcurrent,
		Timeout:       config.Env.PreviewTimeout,
	})
}

package main

import (
	"context"
	"log"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/ismelen/inkomi/back/manga-worker/internal/config"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/ports"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/archive"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/epub"
	"github.com/ismelen/inkomi/back/manga-worker/internal/infra/vips"
	"github.com/ismelen/inkomi/back/manga-worker/internal/usecases/convertmanga"
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

	extractors := map[models.SourceFormat]ports.IPageExtractor{
		models.SourceFormatCBZ: archive.NewCBZExtractor(),
		models.SourceFormatPDF: archive.NewPDFExtractor(),
	}
	convertMangaUC := convertmanga.NewConvertMangaUC(
		queue,
		cloudStorage,
		extractors,
		vips.NewBimgPageProcessor(ctx, config.Env.Workers),
		epub.NewFixedLayoutBuilder(),
		config.Env.Workers,
	)

	if err := queue.Subscribe(string(jobs.SubjectManga), jobrun.Handle(queue, convertMangaUC.Execute)); err != nil {
		log.Fatal(err)
	}

	select {}
}

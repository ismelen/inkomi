package main

import (
	"context"
	"log"

	"github.com/ismelen/inkomi/back/epub-worker/internal/config"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/cloud"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/epub"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/queue"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecaes/joinepubs"
)

func main() {
	queue, err := queue.NewNatsQueue(config.Env.NATSUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	cloudStorage, err := cloud.NewCloudflareR2Storage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	merger := epub.NewEpubMerger()
	if err := joinepubs.NewJoinEpubsUC(queue, cloudStorage, merger).Execute(); err != nil {
		log.Fatal(err)
	}
}

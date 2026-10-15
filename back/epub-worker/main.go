package main

import (
	"context"
	"log"

	"github.com/ismelen/inkomi/back/epub-worker/internal/config"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/cloud"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/epub"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/kepub"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/queue"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/store"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/kepubify"
	"github.com/redis/go-redis/v9"
)

func main() {
	config.Load()

	queue, err := queue.NewNatsQueue(config.Env.NATSUrl, config.Env.NATSAckWait)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	redisOpts, err := redis.ParseURL(config.Env.RedisURL)
	if err != nil {
		log.Fatal(err)
	}
	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()

	cloudStorage, err := cloud.NewCloudflareR2Storage(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	joinStore := store.NewRedisJoinStore(redisClient, config.Env.JoinTTL, config.Env.JoinLockTTL)
	joinEpubsUC := joinepubs.NewJoinEpubsUC(queue, joinStore, cloudStorage, epub.NewEpubMerger(), config.Env.JoinMaxBlockSize)
	if err := joinEpubsUC.Execute(); err != nil {
		log.Fatal(err)
	}

	kepubifyUC := kepubify.NewKepubifyUC(queue, cloudStorage, kepub.NewKepubifyKepubifier())
	if err := kepubifyUC.Execute(); err != nil {
		log.Fatal(err)
	}

	select {}
}

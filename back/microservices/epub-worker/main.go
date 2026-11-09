package main

import (
	"context"
	"log"

	"github.com/ismelen/inkomi/back/microservices/common/cloud"
	"github.com/ismelen/inkomi/back/microservices/common/jobrun"
	"github.com/ismelen/inkomi/back/microservices/common/jobs"
	commonqueue "github.com/ismelen/inkomi/back/microservices/common/queue"
	"github.com/ismelen/inkomi/back/microservices/common/resources"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/config"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/infra/epub"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/infra/kepub"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/infra/store"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/microservices/epub-worker/internal/usecases/kepubify"
	"github.com/redis/go-redis/v9"
)

func main() {
	config.Load()

	queue, err := commonqueue.NewNatsQueue(config.Env.NATSUrl, config.Env.NATSAckWait, config.Env.NATSMaxDeliver)
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

	cloudStorage, err := cloud.NewCloudflareR2Storage(
		context.Background(),
		config.Env.R2AccountId,
		config.Env.R2AccessKeyId,
		config.Env.R2SecretAccessKey,
		config.Env.R2BucketName,
	)
	if err != nil {
		log.Fatal(err)
	}

	joinStore := store.NewRedisJoinStore(redisClient, config.Env.JoinTTL, config.Env.JoinLockTTL)
	joinEpubsUC := joinepubs.NewJoinEpubsUC(queue, joinStore, cloudStorage, epub.NewEpubMerger(), config.Env.JoinMaxBlockSize)
	kepubifyUC := kepubify.NewKepubifyUC(queue, cloudStorage, kepub.NewKepubifyKepubifier())

	memoryGate := resources.NewMemoryGate(resources.NewMemory(), config.Env.MemoryReserve)
	for subject, handler := range map[jobs.Subject]jobrun.Handler{
		jobs.SubjectJoin:  joinEpubsUC.Execute,
		jobs.SubjectKepub: kepubifyUC.Execute,
	} {
		opts := commonqueue.SubscribeOptions{Concurrency: 1, Ready: memoryGate.Ready}
		if err := queue.SubscribeWith(string(subject), opts, jobrun.Handle(queue, handler)); err != nil {
			log.Fatal(err)
		}
	}

	select {}
}

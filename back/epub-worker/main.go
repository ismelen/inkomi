package main

import (
	"log"

	"github.com/ismelen/inkomi/back/epub-worker/internal/config"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/queue"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/store"
	"github.com/ismelen/inkomi/back/epub-worker/internal/usecases/joinepubs"
	"github.com/redis/go-redis/v9"
)

func main() {
	config.Load()

	queue, err := queue.NewNatsQueue(config.Env.NATSUrl)
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

	joinStore := store.NewRedisJoinStore(redisClient, config.Env.JoinTTL)
	if err := joinepubs.NewJoinEpubsUC(queue, joinStore).Execute(); err != nil {
		log.Fatal(err)
	}

	select {}
}

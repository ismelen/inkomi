package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const defaultJoinTTL = 24 * time.Hour

type AppConfig struct {
	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string
	NATSUrl           string
	RedisURL          string
	JoinTTL           time.Duration
}

var Env AppConfig

func Load() {
	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	Env = AppConfig{
		R2AccountId:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyId:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),
		NATSUrl:           os.Getenv("NATS_URL"),
		RedisURL:          os.Getenv("REDIS_URL"),
		JoinTTL:           durationOr(os.Getenv("JOIN_TTL"), defaultJoinTTL),
	}
}

func durationOr(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

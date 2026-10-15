package config

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultNATSAckWait    = 30 * time.Minute
	defaultNATSMaxDeliver = 3
)

type AppConfig struct {
	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string
	NATSUrl           string
	NATSAckWait       time.Duration
	NATSMaxDeliver    int
	// Workers is how many pages are processed at once.
	Workers int
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
		NATSAckWait:       durationOr(os.Getenv("NATS_ACK_WAIT"), defaultNATSAckWait),
		NATSMaxDeliver:    intOr(os.Getenv("NATS_MAX_DELIVER"), defaultNATSMaxDeliver),
		Workers:           intOr(os.Getenv("WORKERS"), max(1, runtime.NumCPU()*3/4)),
	}
}

func durationOr(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func intOr(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

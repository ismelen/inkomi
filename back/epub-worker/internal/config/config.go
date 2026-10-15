package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultJoinTTL        = 24 * time.Hour
	defaultJoinLockTTL    = 30 * time.Minute
	defaultNATSAckWait    = 30 * time.Minute
	defaultJoinMaxBlockMB = 200
)

type AppConfig struct {
	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string
	NATSUrl           string
	NATSAckWait       time.Duration
	RedisURL          string
	JoinTTL           time.Duration
	JoinLockTTL       time.Duration
	JoinMaxBlockSize  int64
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
		RedisURL:          os.Getenv("REDIS_URL"),
		JoinTTL:           durationOr(os.Getenv("JOIN_TTL"), defaultJoinTTL),
		JoinLockTTL:       durationOr(os.Getenv("JOIN_LOCK_TTL"), defaultJoinLockTTL),
		JoinMaxBlockSize:  intOr(os.Getenv("JOIN_MAX_BLOCK_MB"), defaultJoinMaxBlockMB) << 20,
	}
}

func durationOr(value string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func intOr(value string, fallback int64) int64 {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

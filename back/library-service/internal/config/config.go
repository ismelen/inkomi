package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort                  = "3001"
	defaultNATSAckWait           = 30 * time.Minute
	defaultNATSMaxDeliver        = 3
	defaultMirrorRefreshInterval = time.Hour
	defaultMirrorRefreshCooldown = time.Minute
)

type AppConfig struct {
	Port                  string
	JwtPublicKey          string
	R2AccountId           string
	R2AccessKeyId         string
	R2SecretAccessKey     string
	R2BucketName          string
	NATSUrl               string
	NATSAckWait           time.Duration
	NATSMaxDeliver        int
	SlumURL               string
	MirrorRefreshInterval time.Duration
	MirrorRefreshCooldown time.Duration
}

var Env AppConfig

func Load() {
	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	Env = AppConfig{
		Port:                  stringOr(os.Getenv("PORT"), defaultPort),
		JwtPublicKey:          os.Getenv("JWT_PUBLIC_KEY"),
		R2AccountId:           os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyId:         os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:     os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:          os.Getenv("R2_BUCKET_NAME"),
		NATSUrl:               os.Getenv("NATS_URL"),
		NATSAckWait:           durationOr(os.Getenv("NATS_ACK_WAIT"), defaultNATSAckWait),
		NATSMaxDeliver:        intOr(os.Getenv("NATS_MAX_DELIVER"), defaultNATSMaxDeliver),
		SlumURL:               os.Getenv("SLUM_URL"),
		MirrorRefreshInterval: durationOr(os.Getenv("MIRROR_REFRESH_INTERVAL"), defaultMirrorRefreshInterval),
		MirrorRefreshCooldown: durationOr(os.Getenv("MIRROR_REFRESH_COOLDOWN"), defaultMirrorRefreshCooldown),
	}
}

func stringOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
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

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultNATSAckWait             = 30 * time.Minute
	defaultNATSMaxDeliver          = 3
	defaultServiceName             = "push-service"
	defaultAuthServiceAudience     = "auth-service"
	defaultAuthServiceTimeout      = 10 * time.Second
	defaultServiceTokenTTL         = time.Minute
	defaultFirebaseCredentialsFile = "firebase.json"
	defaultDropboxChunkMB          = 64
)

type AppConfig struct {
	R2AccountId             string
	R2AccessKeyId           string
	R2SecretAccessKey       string
	R2BucketName            string
	NATSUrl                 string
	NATSAckWait             time.Duration
	NATSMaxDeliver          int
	ServiceName             string
	ServicePrivateKey       string
	ServiceTokenTTL         time.Duration
	AuthServiceURL          string
	AuthServiceAudience     string
	AuthServiceTimeout      time.Duration
	FirebaseCredentialsFile string
	DropboxChunkSize        int64
}

var Env AppConfig

func Load() {
	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	Env = AppConfig{
		R2AccountId:             os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyId:           os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey:       os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:            os.Getenv("R2_BUCKET_NAME"),
		NATSUrl:                 os.Getenv("NATS_URL"),
		NATSAckWait:             durationOr(os.Getenv("NATS_ACK_WAIT"), defaultNATSAckWait),
		NATSMaxDeliver:          int(intOr(os.Getenv("NATS_MAX_DELIVER"), defaultNATSMaxDeliver)),
		ServiceName:             stringOr(os.Getenv("SERVICE_NAME"), defaultServiceName),
		ServicePrivateKey:       os.Getenv("SERVICE_PRIVATE_KEY"),
		ServiceTokenTTL:         durationOr(os.Getenv("SERVICE_TOKEN_TTL"), defaultServiceTokenTTL),
		AuthServiceURL:          os.Getenv("AUTH_SERVICE_URL"),
		AuthServiceAudience:     stringOr(os.Getenv("AUTH_SERVICE_AUDIENCE"), defaultAuthServiceAudience),
		AuthServiceTimeout:      durationOr(os.Getenv("AUTH_SERVICE_TIMEOUT"), defaultAuthServiceTimeout),
		FirebaseCredentialsFile: stringOr(os.Getenv("FIREBASE_CREDENTIALS_FILE"), defaultFirebaseCredentialsFile),
		DropboxChunkSize:        intOr(os.Getenv("DROPBOX_CHUNK_MB"), defaultDropboxChunkMB) << 20,
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

func intOr(value string, fallback int64) int64 {
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

package config

import (
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort            = "3100"
	defaultInternalPort    = "3101"
	defaultDBPath          = "users.db"
	defaultServiceName     = "auth-service"
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
	defaultExternalTimeout = 10 * time.Second
)

type AppConfig struct {
	Port         string
	InternalPort string
	DBPath       string
	// ServiceName is the issuer of the access tokens and the audience of service tokens.
	ServiceName        string
	JwtPrivateKey      string
	JwtPublicKey       string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	GoogleClientIds    []string
	DropboxAppKey      string
	TokenEncryptionKey string
	// ServicePublicKeys is a JSON object of service name → public key PEM of the services
	// allowed to call the internal routes.
	ServicePublicKeys string
	ExternalTimeout   time.Duration
}

var Env AppConfig

func Load() {
	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	Env = AppConfig{
		Port:               stringOr(os.Getenv("PORT"), defaultPort),
		InternalPort:       stringOr(os.Getenv("INTERNAL_PORT"), defaultInternalPort),
		DBPath:             stringOr(os.Getenv("DB_PATH"), defaultDBPath),
		ServiceName:        stringOr(os.Getenv("SERVICE_NAME"), defaultServiceName),
		JwtPrivateKey:      os.Getenv("JWT_PRIVATE_KEY"),
		JwtPublicKey:       os.Getenv("JWT_PUBLIC_KEY"),
		AccessTokenTTL:     durationOr(os.Getenv("ACCESS_TOKEN_TTL"), defaultAccessTokenTTL),
		RefreshTokenTTL:    durationOr(os.Getenv("REFRESH_TOKEN_TTL"), defaultRefreshTokenTTL),
		GoogleClientIds:    list(os.Getenv("GOOGLE_CLIENT_IDS")),
		DropboxAppKey:      os.Getenv("DROPBOX_APP_KEY"),
		TokenEncryptionKey: os.Getenv("TOKEN_ENCRYPTION_KEY"),
		ServicePublicKeys:  stringOr(os.Getenv("SERVICE_PUBLIC_KEYS"), "{}"),
		ExternalTimeout:    durationOr(os.Getenv("EXTERNAL_TIMEOUT"), defaultExternalTimeout),
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

func list(value string) []string {
	var items []string
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

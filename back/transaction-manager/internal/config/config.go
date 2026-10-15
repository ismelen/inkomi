package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	JwtPublicKey      string
	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string
	NATSUrl           string
	DBPath            string
	SchemaPath        string
}

var Env AppConfig

func Load() {
	godotenv.Load()
	Env = AppConfig{
		JwtPublicKey:      os.Getenv("JWT_PUBLIC_KEY"),
		R2AccountId:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyId:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),
		NATSUrl:           os.Getenv("NATS_URL"),
		DBPath:            stringOr(os.Getenv("DB_PATH"), "transactions.db"),
		SchemaPath:        stringOr(os.Getenv("SCHEMA_PATH"), "./migrations/transactions.sql"),
	}
}

func stringOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

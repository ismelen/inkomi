package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort         = "3200"
	defaultGitHubRepo   = "ismelen/inkomi"
	defaultGitHubAPIURL = "https://api.github.com"
	defaultAPKAssetName = "inkomi.apk"
	defaultCacheTTL     = 5 * time.Minute
)

type AppConfig struct {
	Port string
	// GitHubRepo is owner/name of the repository whose releases carry the APK.
	GitHubRepo   string
	GitHubAPIURL string
	// GitHubToken is optional: without it the GitHub API allows 60 requests per hour.
	GitHubToken  string
	APKAssetName string
	CacheTTL     time.Duration
}

var Env AppConfig

func Load() {
	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	Env = AppConfig{
		Port:         stringOr(os.Getenv("PORT"), defaultPort),
		GitHubRepo:   stringOr(os.Getenv("GITHUB_REPO"), defaultGitHubRepo),
		GitHubAPIURL: stringOr(os.Getenv("GITHUB_API_URL"), defaultGitHubAPIURL),
		GitHubToken:  os.Getenv("GITHUB_TOKEN"),
		APKAssetName: stringOr(os.Getenv("APK_ASSET_NAME"), defaultAPKAssetName),
		CacheTTL:     durationOr(os.Getenv("RELEASE_CACHE_TTL"), defaultCacheTTL),
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

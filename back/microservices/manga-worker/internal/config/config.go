package config

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultNATSAckWait          = 30 * time.Minute
	defaultNATSMaxDeliver       = 3
	defaultMangaJobMemoryMB     = 512
	defaultMemoryReserveMB      = 400
	defaultPort                 = "3003"
	defaultPreviewSamplesId     = "samples"
	defaultPreviewColorSample   = "color.jpg"
	defaultPreviewBWSample      = "bw.jpg"
	defaultPreviewMaxSampleMB   = 15
	defaultPreviewRatePerMinute = 6
	defaultPreviewBurst         = 3
	defaultPreviewMaxConcurrent = 2
	defaultPreviewTimeout       = 20 * time.Second
)

type AppConfig struct {
	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string
	NATSUrl           string
	NATSAckWait       time.Duration
	NATSMaxDeliver    int
	// Workers is how many pages are processed at once, across every job and the preview.
	Workers int
	// MangaConcurrency is how many jobs run at once; 0 picks it from the CPUs and the free memory
	// at startup, assuming MangaJobMemory per job.
	MangaConcurrency int
	MangaJobMemory   uint64
	// MemoryReserve is the free memory below which no new job starts.
	MemoryReserve uint64
	Port          string

	// The preview samples are stored as PreviewSamplesId/<filename> in cloud storage.
	PreviewSamplesId     string
	PreviewColorSample   string
	PreviewBWSample      string
	PreviewMaxSampleSize int64
	PreviewRatePerMinute int
	PreviewBurst         int
	PreviewMaxConcurrent int
	PreviewTimeout       time.Duration
	// PreviewTrustProxy reads the client IP from X-Forwarded-For; only safe behind a proxy that sets it.
	PreviewTrustProxy bool
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
		Workers:           intOr(os.Getenv("WORKERS"), runtime.NumCPU()),
		MangaConcurrency:  intOr(os.Getenv("MANGA_CONCURRENCY"), 0),
		MangaJobMemory:    uint64(intOr(os.Getenv("MANGA_JOB_MEMORY_MB"), defaultMangaJobMemoryMB)) << 20,
		MemoryReserve:     uint64(intOr(os.Getenv("MEMORY_RESERVE_MB"), defaultMemoryReserveMB)) << 20,
		Port:              stringOr(os.Getenv("PORT"), defaultPort),

		PreviewSamplesId:     stringOr(os.Getenv("PREVIEW_SAMPLES_ID"), defaultPreviewSamplesId),
		PreviewColorSample:   stringOr(os.Getenv("PREVIEW_COLOR_SAMPLE"), defaultPreviewColorSample),
		PreviewBWSample:      stringOr(os.Getenv("PREVIEW_BW_SAMPLE"), defaultPreviewBWSample),
		PreviewMaxSampleSize: int64(intOr(os.Getenv("PREVIEW_MAX_SAMPLE_MB"), defaultPreviewMaxSampleMB)) << 20,
		PreviewRatePerMinute: intOr(os.Getenv("PREVIEW_RATE_PER_MINUTE"), defaultPreviewRatePerMinute),
		PreviewBurst:         intOr(os.Getenv("PREVIEW_BURST"), defaultPreviewBurst),
		PreviewMaxConcurrent: intOr(os.Getenv("PREVIEW_MAX_CONCURRENT"), defaultPreviewMaxConcurrent),
		PreviewTimeout:       durationOr(os.Getenv("PREVIEW_TIMEOUT"), defaultPreviewTimeout),
		PreviewTrustProxy:    os.Getenv("PREVIEW_TRUST_PROXY") == "true",
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

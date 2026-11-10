package config

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultPort            = "3000"
	defaultDataDir         = "data"
	defaultTokenIssuer     = "auth-service"
	defaultAccessTokenTTL  = 15 * time.Minute
	defaultRefreshTokenTTL = 30 * 24 * time.Hour
	defaultExternalTimeout = 10 * time.Second

	defaultNATSAckWait    = 30 * time.Minute
	defaultNATSMaxDeliver = 3

	defaultMangaJobMemoryMB = 512
	defaultMemoryReserveMB  = 400

	defaultStorage           = StorageLocal
	defaultFilesRetention    = 7 * 24 * time.Hour
	defaultDiskReserveMB     = 1024
	defaultStorageEpubRatio  = 1.2
	defaultLibraryEstimateMB = 100

	defaultMirrorRefreshInterval = time.Hour
	defaultMirrorRefreshCooldown = time.Minute

	defaultPreviewSamplesId     = "samples"
	defaultPreviewColorSample   = "color.jpg"
	defaultPreviewBWSample      = "bw.jpg"
	defaultPreviewMaxSampleMB   = 15
	defaultPreviewRatePerMinute = 6
	defaultPreviewBurst         = 3
	defaultPreviewMaxConcurrent = 2
	defaultPreviewTimeout       = 20 * time.Second

	defaultJoinTTL        = 24 * time.Hour
	defaultJoinLockTTL    = 30 * time.Minute
	defaultJoinMaxBlockMB = 200

	defaultFirebaseCredentialsFile = "firebase.json"
	defaultDropboxChunkMB          = 64

	defaultGitHubRepo   = "ismelen/inkomi"
	defaultGitHubAPIURL = "https://api.github.com"
	defaultAPKAssetName = "inkomi.apk"
	defaultCacheTTL     = 5 * time.Minute
)

const (
	StorageLocal = "local"
	StorageR2    = "r2"
)

type AppConfig struct {
	Port string
	// DataDir holds the SQLite databases and the JetStream store of the embedded NATS.
	DataDir string

	// TokenIssuer is the issuer of the access tokens; it stays auth-service so the tokens issued
	// by the microservices keep working.
	TokenIssuer        string
	JwtPrivateKey      string
	JwtPublicKey       string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	GoogleClientIds    []string
	DropboxAppKey      string
	TokenEncryptionKey string
	ExternalTimeout    time.Duration

	// Storage is where the files live: "local" (the server's disk, under DataDir/files) or "r2".
	Storage string
	// PublicURL is the public address of the monolith, the base of the signed /files urls.
	PublicURL string
	// StorageSigningKey signs the /files urls; empty derives it from TokenEncryptionKey.
	StorageSigningKey string
	// FilesRetention is how long a stored file nobody touched is kept (downloadable results).
	FilesRetention time.Duration
	// DiskReserve is the free disk kept out of every upload's reservation and below which no heavy
	// job starts.
	DiskReserve uint64
	// StorageEpubRatio is the epub size relative to its manga file, for the space reservations.
	StorageEpubRatio float64
	// LibraryEstimate is the size reserved for a library download, whose size is unknown.
	LibraryEstimate int64

	R2AccountId       string
	R2AccessKeyId     string
	R2SecretAccessKey string
	R2BucketName      string

	NATSAckWait    time.Duration
	NATSMaxDeliver int

	SlumURL               string
	MirrorRefreshInterval time.Duration
	MirrorRefreshCooldown time.Duration

	// Workers is how many manga pages are processed at once, across every job and the preview.
	Workers int
	// MangaConcurrency is how many manga jobs run at once; 0 picks it from the CPUs and the free
	// memory at startup, assuming MangaJobMemory per job.
	MangaConcurrency int
	MangaJobMemory   uint64
	// MemoryReserve is the free memory below which no new heavy job (manga, join, kepub) starts.
	MemoryReserve uint64
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

	JoinTTL          time.Duration
	JoinLockTTL      time.Duration
	JoinMaxBlockSize int64

	FirebaseCredentialsFile string
	DropboxChunkSize        int64

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
		Port:    stringOr(os.Getenv("PORT"), defaultPort),
		DataDir: stringOr(os.Getenv("DATA_DIR"), defaultDataDir),

		TokenIssuer:        stringOr(os.Getenv("TOKEN_ISSUER"), defaultTokenIssuer),
		JwtPrivateKey:      os.Getenv("JWT_PRIVATE_KEY"),
		JwtPublicKey:       os.Getenv("JWT_PUBLIC_KEY"),
		AccessTokenTTL:     durationOr(os.Getenv("ACCESS_TOKEN_TTL"), defaultAccessTokenTTL),
		RefreshTokenTTL:    durationOr(os.Getenv("REFRESH_TOKEN_TTL"), defaultRefreshTokenTTL),
		GoogleClientIds:    list(os.Getenv("GOOGLE_CLIENT_IDS")),
		DropboxAppKey:      os.Getenv("DROPBOX_APP_KEY"),
		TokenEncryptionKey: os.Getenv("TOKEN_ENCRYPTION_KEY"),
		ExternalTimeout:    durationOr(os.Getenv("EXTERNAL_TIMEOUT"), defaultExternalTimeout),

		Storage:           stringOr(os.Getenv("STORAGE"), defaultStorage),
		PublicURL:         os.Getenv("PUBLIC_URL"),
		StorageSigningKey: os.Getenv("STORAGE_SIGNING_KEY"),
		FilesRetention:    durationOr(os.Getenv("FILES_RETENTION"), defaultFilesRetention),
		DiskReserve:       uint64(intOr(os.Getenv("DISK_RESERVE_MB"), defaultDiskReserveMB)) << 20,
		StorageEpubRatio:  floatOr(os.Getenv("STORAGE_EPUB_RATIO"), defaultStorageEpubRatio),
		LibraryEstimate:   int64(intOr(os.Getenv("LIBRARY_ESTIMATE_MB"), defaultLibraryEstimateMB)) << 20,

		R2AccountId:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyId:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),

		NATSAckWait:    durationOr(os.Getenv("NATS_ACK_WAIT"), defaultNATSAckWait),
		NATSMaxDeliver: intOr(os.Getenv("NATS_MAX_DELIVER"), defaultNATSMaxDeliver),

		SlumURL:               os.Getenv("SLUM_URL"),
		MirrorRefreshInterval: durationOr(os.Getenv("MIRROR_REFRESH_INTERVAL"), defaultMirrorRefreshInterval),
		MirrorRefreshCooldown: durationOr(os.Getenv("MIRROR_REFRESH_COOLDOWN"), defaultMirrorRefreshCooldown),

		Workers:              intOr(os.Getenv("WORKERS"), runtime.NumCPU()),
		MangaConcurrency:     intOr(os.Getenv("MANGA_CONCURRENCY"), 0),
		MangaJobMemory:       uint64(intOr(os.Getenv("MANGA_JOB_MEMORY_MB"), defaultMangaJobMemoryMB)) << 20,
		MemoryReserve:        uint64(intOr(os.Getenv("MEMORY_RESERVE_MB"), defaultMemoryReserveMB)) << 20,
		PreviewSamplesId:     stringOr(os.Getenv("PREVIEW_SAMPLES_ID"), defaultPreviewSamplesId),
		PreviewColorSample:   stringOr(os.Getenv("PREVIEW_COLOR_SAMPLE"), defaultPreviewColorSample),
		PreviewBWSample:      stringOr(os.Getenv("PREVIEW_BW_SAMPLE"), defaultPreviewBWSample),
		PreviewMaxSampleSize: int64(intOr(os.Getenv("PREVIEW_MAX_SAMPLE_MB"), defaultPreviewMaxSampleMB)) << 20,
		PreviewRatePerMinute: intOr(os.Getenv("PREVIEW_RATE_PER_MINUTE"), defaultPreviewRatePerMinute),
		PreviewBurst:         intOr(os.Getenv("PREVIEW_BURST"), defaultPreviewBurst),
		PreviewMaxConcurrent: intOr(os.Getenv("PREVIEW_MAX_CONCURRENT"), defaultPreviewMaxConcurrent),
		PreviewTimeout:       durationOr(os.Getenv("PREVIEW_TIMEOUT"), defaultPreviewTimeout),
		PreviewTrustProxy:    os.Getenv("PREVIEW_TRUST_PROXY") == "true",

		JoinTTL:          durationOr(os.Getenv("JOIN_TTL"), defaultJoinTTL),
		JoinLockTTL:      durationOr(os.Getenv("JOIN_LOCK_TTL"), defaultJoinLockTTL),
		JoinMaxBlockSize: int64(intOr(os.Getenv("JOIN_MAX_BLOCK_MB"), defaultJoinMaxBlockMB)) << 20,

		FirebaseCredentialsFile: stringOr(os.Getenv("FIREBASE_CREDENTIALS_FILE"), defaultFirebaseCredentialsFile),
		DropboxChunkSize:        int64(intOr(os.Getenv("DROPBOX_CHUNK_MB"), defaultDropboxChunkMB)) << 20,

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

func intOr(value string, fallback int) int {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
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

func floatOr(value string, fallback float64) float64 {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil || f <= 0 {
		return fallback
	}
	return f
}

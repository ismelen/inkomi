package main

import (
	"context"
	"crypto/sha256"
	"log"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/cloud"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/storage"
	"github.com/ismelen/inkomi/back/monolith/samples"
)

// wireStorage picks where the files live: the server's disk (served on /files with signed urls,
// swept after FILES_RETENTION) or R2.
func wireStorage(d *deps, router chi.Router) {
	if config.Env.Storage == config.StorageR2 {
		r2, err := cloud.NewCloudflareR2Storage(
			d.ctx,
			config.Env.R2AccountId,
			config.Env.R2AccessKeyId,
			config.Env.R2SecretAccessKey,
			config.Env.R2BucketName,
		)
		if err != nil {
			log.Fatal(err)
		}
		d.cloud = r2
		log.Printf("storage: r2 bucket %s", config.Env.R2BucketName)
		return
	}

	if config.Env.PublicURL == "" {
		log.Fatal("PUBLIC_URL is required with STORAGE=local: the signed upload urls point to it")
	}
	local, err := storage.NewLocalStorage(
		filepath.Join(config.Env.DataDir, "files"),
		filepath.Join(config.Env.DataDir, "tmp"),
	)
	if err != nil {
		log.Fatal(err)
	}
	signer := storage.NewURLSigner(signingKey(), config.Env.PublicURL, time.Now)
	storage.NewFilesHandler(local, signer).Setup(router)
	storage.NewJanitor(local, config.Env.FilesRetention, time.Now).Start(d.ctx, time.Hour)
	seedSamples(d.ctx, local)

	d.cloud, d.local, d.signer = local, local, signer
	free, _ := storage.FreeBytes(local.Root())
	log.Printf("storage: local %s, %d MB free", local.Root(), free>>20)
}

// signingKey is STORAGE_SIGNING_KEY or, when unset, a key derived from TOKEN_ENCRYPTION_KEY, so
// the urls survive restarts without another secret.
func signingKey() []byte {
	if config.Env.StorageSigningKey != "" {
		return []byte(config.Env.StorageSigningKey)
	}
	sum := sha256.Sum256([]byte("inkomi-files:" + config.Env.TokenEncryptionKey))
	return sum[:]
}

// seedSamples stores the embedded preview samples when they're missing.
func seedSamples(ctx context.Context, local *storage.LocalStorage) {
	for embedded, name := range map[string]string{
		samples.Color: config.Env.PreviewColorSample,
		samples.BW:    config.Env.PreviewBWSample,
	} {
		if local.Exists(config.Env.PreviewSamplesId, name) {
			continue
		}
		file, err := samples.Files.Open(embedded)
		if err != nil {
			log.Printf("storage: sample %s: %v", embedded, err)
			continue
		}
		if err := local.Upload(ctx, config.Env.PreviewSamplesId, name, file); err != nil {
			log.Printf("storage: seed sample %s: %v", name, err)
		}
		// an embedded file has nothing to release
		_ = file.Close()
	}
}

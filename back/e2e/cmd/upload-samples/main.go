// Command upload-samples stores the manga-worker preview samples in cloud storage as
// <PREVIEW_SAMPLES_ID>/<file>, uploading only the ones whose content changed. It prints
// "samples-changed" when it uploaded any, so the pipeline knows to restart manga-worker.
package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"github.com/ismelen/inkomi/back/common/cloud"
)

func main() {
	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	ctx := context.Background()

	dir := getenv("SAMPLES_DIR", "/app/samples")
	id := getenv("PREVIEW_SAMPLES_ID", "samples")

	storage, err := cloud.NewCloudflareR2Storage(
		ctx,
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_SECRET_ACCESS_KEY"),
		os.Getenv("R2_BUCKET_NAME"),
	)
	if err != nil {
		log.Fatal(err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	changed := false
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		local, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		if same(ctx, storage, id, entry.Name(), local) {
			log.Printf("%s/%s unchanged", id, entry.Name())
			continue
		}
		if _, err := storage.UploadFile(ctx, id, entry.Name(), path); err != nil {
			log.Fatalf("upload %s: %v", entry.Name(), err)
		}
		log.Printf("%s/%s uploaded (%d bytes)", id, entry.Name(), len(local))
		changed = true
	}

	if changed {
		log.Print("samples-changed")
	}
}

// same reports whether id/filename already holds local. A missing or unreadable object counts
// as different, so it's uploaded.
func same(ctx context.Context, storage cloud.ICloudStorage, id, filename string, local []byte) bool {
	reader, err := storage.Download(ctx, id, filename)
	if err != nil {
		return false
	}
	defer reader.Close()

	remote, err := io.ReadAll(reader)
	return err == nil && bytes.Equal(remote, local)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

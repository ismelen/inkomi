package cloudfile

import (
	"context"
	"fmt"
	"os"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
)

// Upload stores the local file at path as id/filename and returns its size.
func Upload(ctx context.Context, storage ports.CloudStorage, id, filename, path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("stat %s: %w", path, err)
	}

	if err := storage.Upload(ctx, id, filename, file); err != nil {
		return 0, fmt.Errorf("upload %s/%s: %w", id, filename, err)
	}
	return info.Size(), nil
}

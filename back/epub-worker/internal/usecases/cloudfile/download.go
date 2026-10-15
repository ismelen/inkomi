package cloudfile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
)

// Download saves the cloud file id/filename to path.
func Download(ctx context.Context, storage ports.CloudStorage, id, filename, path string) error {
	reader, err := storage.Download(ctx, id, filename)
	if err != nil {
		return fmt.Errorf("download %s/%s: %w", id, filename, err)
	}
	defer reader.Close()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if _, err := io.Copy(file, reader); err != nil {
		return errors.Join(fmt.Errorf("write %s: %w", path, err), file.Close())
	}
	return file.Close()
}

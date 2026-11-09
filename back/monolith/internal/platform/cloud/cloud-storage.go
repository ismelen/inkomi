package cloud

import (
	"context"
	"io"
)

type ICloudStorage interface {
	Upload(ctx context.Context, id string, filename string, body io.Reader) error
	Download(ctx context.Context, id string, filename string) (io.ReadCloser, error)
	// UploadFile stores the local file at path as id/filename and returns its size.
	UploadFile(ctx context.Context, id string, filename string, path string) (int64, error)
	// DownloadFile saves id/filename to the local file at path.
	DownloadFile(ctx context.Context, id string, filename string, path string) error
	Size(ctx context.Context, id string, filename string) (int64, error)
	Delete(ctx context.Context, id string, filename string) error
}

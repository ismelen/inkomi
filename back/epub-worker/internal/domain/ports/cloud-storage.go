package ports

import (
	"context"
	"io"
)

type CloudStorage interface {
	Upload(ctx context.Context, id string, filename string, body io.Reader) error
	Download(ctx context.Context, id string, filename string) (io.ReadCloser, error)
	Delete(ctx context.Context, id string, filename string) error
}


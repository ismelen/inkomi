package libgen

import (
	"context"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
)

// IMirror is one libgen site; each flavour (classic, plus) parses its own search pages.
type IMirror interface {
	Search(ctx context.Context, query string) ([]models.Book, error)
	Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error)
	URL() string
}

type IMirrorList interface {
	Mirrors(ctx context.Context) []IMirror
}

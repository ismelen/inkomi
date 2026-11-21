package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
)

type ILibrary interface {
	// Search returns a page of results; page starts at 1.
	Search(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error)
	// Download returns models.ErrBookNotFound when the library has no file for md5.
	Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error)
}

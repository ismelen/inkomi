package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
)

type ILibrary interface {
	Search(ctx context.Context, query string) ([]models.Book, error)
	// Download returns models.ErrBookNotFound when the library has no file for md5.
	Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error)
}

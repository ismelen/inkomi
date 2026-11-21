package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type ICatalogLookup interface {
	Books(ctx context.Context, query string, limit int) ([]models.Suggestion, error)
	Authors(ctx context.Context, query string, limit int) ([]models.Suggestion, error)
	// Work returns models.ErrWorkNotFound when the catalog has no such work.
	Work(ctx context.Context, workKey string) (models.WorkDetails, error)
	// AuthorWorks ignores a language it can't filter by.
	AuthorWorks(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error)
}

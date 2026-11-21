package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type ICatalogLookup interface {
	Books(ctx context.Context, query string, limit int) ([]models.Suggestion, error)
	Authors(ctx context.Context, query string, limit int) ([]models.Suggestion, error)
	// AuthorWorks ignores a language it can't filter by.
	AuthorWorks(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error)
}

package authorworks

import (
	"context"
	"fmt"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

const pageSize = 20

// AuthorWorksUC lists the works of an author, a page at a time.
type AuthorWorksUC struct {
	catalog ports.ICatalogLookup
}

func NewAuthorWorksUC(catalog ports.ICatalogLookup) *AuthorWorksUC {
	return &AuthorWorksUC{catalog: catalog}
}

// Execute sorts by popularity when query has no sort.
func (a *AuthorWorksUC) Execute(ctx context.Context, query models.AuthorWorksQuery) (models.WorksPage, error) {
	query.AuthorKey = strings.TrimSpace(query.AuthorKey)
	query.Author = strings.TrimSpace(query.Author)
	query.Language = strings.TrimSpace(query.Language)
	if query.AuthorKey == "" && query.Author == "" {
		return models.WorksPage{}, models.ErrEmptyQuery
	}
	if query.Sort == "" {
		query.Sort = models.WorksSortPopular
	}
	if !query.Sort.IsValid() {
		return models.WorksPage{}, models.ErrInvalidSort
	}
	if query.Page < 1 {
		return models.WorksPage{}, models.ErrInvalidPage
	}

	page, err := a.catalog.AuthorWorks(ctx, query, pageSize)
	if err != nil {
		return models.WorksPage{}, fmt.Errorf("%w: %w", models.ErrLookupFailure, err)
	}
	return page, nil
}

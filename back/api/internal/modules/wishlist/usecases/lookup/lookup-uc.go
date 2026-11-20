package lookup

import (
	"context"
	"fmt"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

const (
	minQueryLength = 2
	limit          = 10
)

// LookupUC suggests catalog books or authors while the user types a wish.
type LookupUC struct {
	catalog ports.ICatalogLookup
}

func NewLookupUC(catalog ports.ICatalogLookup) *LookupUC {
	return &LookupUC{catalog: catalog}
}

func (l *LookupUC) Execute(ctx context.Context, kind models.ItemKind, query string) ([]models.Suggestion, error) {
	if !kind.IsValid() {
		return nil, models.ErrInvalidKind
	}
	query = strings.TrimSpace(query)
	if len([]rune(query)) < minQueryLength {
		return nil, models.ErrEmptyQuery
	}

	var suggestions []models.Suggestion
	var err error
	if kind == models.ItemKindAuthor {
		suggestions, err = l.catalog.Authors(ctx, query, limit)
	} else {
		suggestions, err = l.catalog.Books(ctx, query, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrLookupFailure, err)
	}
	return suggestions, nil
}

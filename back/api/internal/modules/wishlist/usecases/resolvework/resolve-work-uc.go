package resolvework

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

// ResolveWorkUC gives a wished book its catalog work and stores the work's details, so the
// book has a cover and a synopsis: a book typed by hand gets the work key of the best catalog
// match.
type ResolveWorkUC struct {
	items   ports.IItemRepository
	works   ports.IWorkRepository
	catalog ports.ICatalogLookup
}

func NewResolveWorkUC(items ports.IItemRepository, works ports.IWorkRepository, catalog ports.ICatalogLookup) *ResolveWorkUC {
	return &ResolveWorkUC{items: items, works: works, catalog: catalog}
}

// Execute updates item (WorkKey, CoverURL). coverHint is the cover the user saw when picking
// the book (an edition in their language); it wins over the work's one and is ignored unless
// it is a catalog cover. A book with no catalog match is left as it is.
func (r *ResolveWorkUC) Execute(ctx context.Context, item *models.Item, coverHint string) error {
	if item.Kind != models.ItemKindBook {
		return nil
	}
	if !models.IsCatalogCoverURL(coverHint) {
		coverHint = ""
	}

	if item.WorkKey == "" {
		query := strings.TrimSpace(item.Title + " " + item.Author)
		matches, err := r.catalog.Books(ctx, query, 1)
		if err != nil {
			return fmt.Errorf("%w: %w", models.ErrLookupFailure, err)
		}
		if len(matches) == 0 || matches[0].WorkKey == "" {
			return nil
		}
		if coverHint == "" && models.IsCatalogCoverURL(matches[0].CoverURL) {
			coverHint = matches[0].CoverURL
		}
		err = r.items.SetWorkKey(ctx, item, matches[0].WorkKey)
		if errors.Is(err, models.ErrDuplicate) {
			return nil
		}
		if err != nil {
			return err
		}
	}

	work, err := r.details(ctx, item)
	if err != nil {
		return err
	}
	// a stored cover is kept: the first user to wish the work chose it
	if coverHint != "" && (!work.stored || work.CoverURL == "") && work.CoverURL != coverHint {
		work.CoverURL = coverHint
		work.stored = false
	}
	if !work.stored {
		if err := r.works.Save(ctx, work.WorkDetails); err != nil {
			return err
		}
	}
	item.CoverURL = work.CoverURL
	return nil
}

type workDetails struct {
	models.WorkDetails
	stored bool
}

// details returns the stored details of the work of item, or the catalog's ones; a work the
// catalog doesn't know gets details with just its title, so it isn't asked for again.
func (r *ResolveWorkUC) details(ctx context.Context, item *models.Item) (workDetails, error) {
	stored, err := r.works.Get(ctx, item.WorkKey)
	if err == nil {
		return workDetails{WorkDetails: *stored, stored: true}, nil
	}
	if !errors.Is(err, models.ErrWorkNotFound) {
		return workDetails{}, err
	}

	work, err := r.catalog.Work(ctx, item.WorkKey)
	switch {
	case errors.Is(err, models.ErrWorkNotFound):
		work = models.WorkDetails{Title: item.Title}
	case err != nil:
		return workDetails{}, fmt.Errorf("%w: %w", models.ErrLookupFailure, err)
	}
	work.WorkKey = item.WorkKey
	return workDetails{WorkDetails: work}, nil
}

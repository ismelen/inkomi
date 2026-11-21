package getwork

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

// GetWorkUC returns the details of a work (synopsis, cover...), from the stored ones or else
// from the catalog, which it then stores.
type GetWorkUC struct {
	works   ports.IWorkRepository
	catalog ports.ICatalogLookup
}

func NewGetWorkUC(works ports.IWorkRepository, catalog ports.ICatalogLookup) *GetWorkUC {
	return &GetWorkUC{works: works, catalog: catalog}
}

func (g *GetWorkUC) Execute(ctx context.Context, workKey string) (*models.WorkDetails, error) {
	if !models.IsWorkKey(workKey) {
		return nil, models.ErrInvalidWorkKey
	}
	stored, err := g.works.Get(ctx, workKey)
	if err == nil {
		return stored, nil
	}
	if !errors.Is(err, models.ErrWorkNotFound) {
		return nil, err
	}

	work, err := g.catalog.Work(ctx, workKey)
	if errors.Is(err, models.ErrWorkNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrLookupFailure, err)
	}
	work.WorkKey = workKey
	if err := g.works.Save(ctx, work); err != nil {
		// the details are still good to answer with; the next request fetches them again
		log.Printf("wishlist: save work %s: %v", workKey, err)
	}
	return &work, nil
}

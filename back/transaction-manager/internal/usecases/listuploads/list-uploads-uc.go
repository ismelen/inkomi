package listuploads

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type ListUploadsUC struct {
	sourceRepo ports.ISourceRepository
}

func NewListUploadsUC(sourceRepo ports.ISourceRepository) *ListUploadsUC {
	return &ListUploadsUC{sourceRepo: sourceRepo}
}

func (u *ListUploadsUC) Execute(ctx context.Context, userId int, filter models.SourceListFilter) (*models.SourcePage, error) {
	// Read the clock first so changes made while listing are returned again on the next sync.
	syncedAt, err := u.sourceRepo.Now(ctx)
	if err != nil {
		return nil, err
	}

	query := filter
	query.Limit++
	trees, err := u.sourceRepo.ListTreesByUserId(ctx, userId, query)
	if err != nil {
		return nil, err
	}

	page := &models.SourcePage{Trees: trees, SyncedAt: syncedAt}
	if len(trees) > filter.Limit {
		page.Trees = trees[:filter.Limit]
		page.NextCursor = models.NewSourceCursor(&page.Trees[filter.Limit-1].Source)
	}

	return page, nil
}

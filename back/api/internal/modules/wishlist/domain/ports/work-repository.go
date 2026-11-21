package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type IWorkRepository interface {
	// Get returns models.ErrWorkNotFound when the work isn't stored.
	Get(ctx context.Context, workKey string) (*models.WorkDetails, error)
	// Save inserts or replaces the work.
	Save(ctx context.Context, work models.WorkDetails) error
}

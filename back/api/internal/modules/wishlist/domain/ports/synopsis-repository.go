package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type ISynopsisRepository interface {
	// Get returns models.ErrSynopsisNotFound when it isn't stored.
	Get(ctx context.Context, workKey, lang string) (*models.Synopsis, error)
	// Save inserts or replaces the synopsis, stamped with the current time.
	Save(ctx context.Context, synopsis models.Synopsis) error
}

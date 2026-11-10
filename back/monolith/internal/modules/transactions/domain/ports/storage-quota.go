package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
)

type IStorageQuota interface {
	// Exclusive runs fn while no other reservation is checked, so a check and the transaction
	// that commits it can't interleave with another upload.
	Exclusive(fn func() error) error
	// Reserve records the reservations, in ctx's transaction, when they fit in the free disk
	// besides the space still pending for the active ones; models.ErrInsufficientStorage otherwise.
	Reserve(ctx context.Context, reservations []models.StorageReservation) error
	// ReleaseFinished frees the reservation of sourceId's tree once the whole tree finished.
	ReleaseFinished(ctx context.Context, sourceId string, userId int) error
}

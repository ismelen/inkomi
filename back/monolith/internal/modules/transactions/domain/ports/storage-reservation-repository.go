package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
)

type IStorageReservationRepository interface {
	Create(ctx context.Context, reservation *models.StorageReservation) error
	ListActive(ctx context.Context) ([]models.StorageReservation, error)
	Release(ctx context.Context, rootId string) error
	// RootOf returns the root of the tree sourceId belongs to (itself when it has no parent).
	RootOf(ctx context.Context, sourceId string, userId int) (string, error)
	// TreeIds returns the ids of the root and every source under it, results included.
	TreeIds(ctx context.Context, rootId string) ([]string, error)
	// IsTreeFinished reports whether every file of the tree reached a final status; a joined
	// folder also needs its results.
	IsTreeFinished(ctx context.Context, rootId string) (bool, error)
	// IsTreeUnconfirmed reports whether no file of the tree was ever confirmed.
	IsTreeUnconfirmed(ctx context.Context, rootId string) (bool, error)
}

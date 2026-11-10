package storagequota

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
)

var _ ports.IStorageQuota = NoQuota{}

// NoQuota admits every upload: R2 doesn't run out of space.
type NoQuota struct{}

func (NoQuota) Exclusive(fn func() error) error { return fn() }

func (NoQuota) Reserve(context.Context, []models.StorageReservation) error { return nil }

func (NoQuota) ReleaseFinished(context.Context, string, int) error { return nil }

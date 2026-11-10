package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
)

type MockStorageQuota struct {
	ReserveFn         func(ctx context.Context, reservations []models.StorageReservation) error
	ReleaseFinishedFn func(ctx context.Context, sourceId string, userId int) error
}

func (m *MockStorageQuota) Exclusive(fn func() error) error { return fn() }

func (m *MockStorageQuota) Reserve(ctx context.Context, reservations []models.StorageReservation) error {
	if m.ReserveFn != nil {
		return m.ReserveFn(ctx, reservations)
	}
	return nil
}

func (m *MockStorageQuota) ReleaseFinished(ctx context.Context, sourceId string, userId int) error {
	if m.ReleaseFinishedFn != nil {
		return m.ReleaseFinishedFn(ctx, sourceId, userId)
	}
	return nil
}

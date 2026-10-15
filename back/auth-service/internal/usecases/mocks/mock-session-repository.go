package mocks

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type MockSessionRepository struct {
	SaveFn                func(ctx context.Context, session models.Session) error
	RotateFn              func(ctx context.Context, oldHash, newHash string, now, expiresAt time.Time) (int, error)
	DeleteByRefreshHashFn func(ctx context.Context, hash string) error
}

func (m *MockSessionRepository) Save(ctx context.Context, session models.Session) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, session)
	}
	return nil
}

func (m *MockSessionRepository) Rotate(ctx context.Context, oldHash, newHash string, now, expiresAt time.Time) (int, error) {
	if m.RotateFn != nil {
		return m.RotateFn(ctx, oldHash, newHash, now, expiresAt)
	}
	return 0, models.ErrInvalidRefreshToken
}

func (m *MockSessionRepository) DeleteByRefreshHash(ctx context.Context, hash string) error {
	if m.DeleteByRefreshHashFn != nil {
		return m.DeleteByRefreshHashFn(ctx, hash)
	}
	return nil
}

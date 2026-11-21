package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type MockSynopsisRepository struct {
	GetFn  func(ctx context.Context, workKey, lang string) (*models.Synopsis, error)
	SaveFn func(ctx context.Context, synopsis models.Synopsis) error
}

func (m *MockSynopsisRepository) Get(ctx context.Context, workKey, lang string) (*models.Synopsis, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, workKey, lang)
	}
	return nil, models.ErrSynopsisNotFound
}

func (m *MockSynopsisRepository) Save(ctx context.Context, synopsis models.Synopsis) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, synopsis)
	}
	return nil
}

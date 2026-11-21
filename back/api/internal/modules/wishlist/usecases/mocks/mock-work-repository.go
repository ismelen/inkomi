package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type MockWorkRepository struct {
	GetFn  func(ctx context.Context, workKey string) (*models.WorkDetails, error)
	SaveFn func(ctx context.Context, work models.WorkDetails) error
}

func (m *MockWorkRepository) Get(ctx context.Context, workKey string) (*models.WorkDetails, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, workKey)
	}
	return nil, models.ErrWorkNotFound
}

func (m *MockWorkRepository) Save(ctx context.Context, work models.WorkDetails) error {
	if m.SaveFn != nil {
		return m.SaveFn(ctx, work)
	}
	return nil
}

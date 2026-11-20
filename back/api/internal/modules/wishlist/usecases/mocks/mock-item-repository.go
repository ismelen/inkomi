package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
)

type MockItemRepository struct {
	AddFn           func(ctx context.Context, item *models.Item) error
	DeleteFn        func(ctx context.Context, id int64, userId int) error
	ListByUserIdFn  func(ctx context.Context, userId int) ([]models.Item, error)
	CountByUserIdFn func(ctx context.Context, userId int) (int, error)
	ListAllFn       func(ctx context.Context) ([]models.Item, error)
}

func (m *MockItemRepository) Add(ctx context.Context, item *models.Item) error {
	if m.AddFn != nil {
		return m.AddFn(ctx, item)
	}
	item.Id = 1
	return nil
}

func (m *MockItemRepository) Delete(ctx context.Context, id int64, userId int) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id, userId)
	}
	return nil
}

func (m *MockItemRepository) ListByUserId(ctx context.Context, userId int) ([]models.Item, error) {
	if m.ListByUserIdFn != nil {
		return m.ListByUserIdFn(ctx, userId)
	}
	return nil, nil
}

func (m *MockItemRepository) CountByUserId(ctx context.Context, userId int) (int, error) {
	if m.CountByUserIdFn != nil {
		return m.CountByUserIdFn(ctx, userId)
	}
	return 0, nil
}

func (m *MockItemRepository) ListAll(ctx context.Context) ([]models.Item, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx)
	}
	return nil, nil
}

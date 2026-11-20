package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type MockWishlist struct {
	ListAllFn func(ctx context.Context) ([]models.WishedItem, error)
}

func (m *MockWishlist) ListAll(ctx context.Context) ([]models.WishedItem, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx)
	}
	return nil, nil
}

package usecases_test

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type MockConfigRepository struct {
	CreateFn func(ctx context.Context, config *models.Config) (string, error)
}

func (m *MockConfigRepository) Create(ctx context.Context, config *models.Config) (string, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, config)
	}
	return "", nil
}

func (m *MockConfigRepository) GetByHash(ctx context.Context, hash string) (*models.Config, error) {
	return nil, nil
}


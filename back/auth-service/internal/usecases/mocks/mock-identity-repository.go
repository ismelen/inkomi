package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type MockIdentityRepository struct {
	GetFn           func(ctx context.Context, provider models.Provider, subject string) (*models.Identity, error)
	CreateFn        func(ctx context.Context, identity models.Identity) error
	ListProvidersFn func(ctx context.Context, userId int) ([]models.Provider, error)
}

func (m *MockIdentityRepository) Get(ctx context.Context, provider models.Provider, subject string) (*models.Identity, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, provider, subject)
	}
	return nil, models.ErrIdentityNotFound
}

func (m *MockIdentityRepository) Create(ctx context.Context, identity models.Identity) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, identity)
	}
	return nil
}

func (m *MockIdentityRepository) ListProviders(ctx context.Context, userId int) ([]models.Provider, error) {
	if m.ListProvidersFn != nil {
		return m.ListProvidersFn(ctx, userId)
	}
	return nil, nil
}

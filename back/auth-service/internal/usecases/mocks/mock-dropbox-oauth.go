package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type MockDropboxOAuth struct {
	RefreshFn func(ctx context.Context, refreshToken string) (models.AccessToken, error)
	AccountFn func(ctx context.Context, accessToken string) (*models.ExternalAccount, error)
}

func (m *MockDropboxOAuth) Refresh(ctx context.Context, refreshToken string) (models.AccessToken, error) {
	if m.RefreshFn != nil {
		return m.RefreshFn(ctx, refreshToken)
	}
	return models.AccessToken{}, models.ErrInvalidExternalToken
}

func (m *MockDropboxOAuth) Account(ctx context.Context, accessToken string) (*models.ExternalAccount, error) {
	if m.AccountFn != nil {
		return m.AccountFn(ctx, accessToken)
	}
	return nil, models.ErrInvalidExternalToken
}

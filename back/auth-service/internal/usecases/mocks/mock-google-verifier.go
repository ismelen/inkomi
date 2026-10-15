package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type MockGoogleVerifier struct {
	VerifyFn func(ctx context.Context, idToken string) (*models.ExternalAccount, error)
}

func (m *MockGoogleVerifier) Verify(ctx context.Context, idToken string) (*models.ExternalAccount, error) {
	if m.VerifyFn != nil {
		return m.VerifyFn(ctx, idToken)
	}
	return nil, models.ErrInvalidExternalToken
}

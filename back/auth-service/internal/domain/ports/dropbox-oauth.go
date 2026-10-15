package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type IDropboxOAuth interface {
	// Refresh returns models.ErrInvalidExternalToken when refreshToken is revoked or isn't
	// from this app.
	Refresh(ctx context.Context, refreshToken string) (models.AccessToken, error)
	Account(ctx context.Context, accessToken string) (*models.ExternalAccount, error)
}

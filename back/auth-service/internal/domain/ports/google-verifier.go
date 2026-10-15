package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type IGoogleVerifier interface {
	// Verify returns models.ErrInvalidExternalToken when idToken isn't valid for this app.
	Verify(ctx context.Context, idToken string) (*models.ExternalAccount, error)
}

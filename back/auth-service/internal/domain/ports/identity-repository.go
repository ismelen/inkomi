package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type IIdentityRepository interface {
	Get(ctx context.Context, provider models.Provider, subject string) (*models.Identity, error)
	Create(ctx context.Context, identity models.Identity) error
	ListProviders(ctx context.Context, userId int) ([]models.Provider, error)
}

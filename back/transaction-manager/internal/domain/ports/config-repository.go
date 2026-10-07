package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type ConfigRepository interface {
	Create(ctx context.Context, config *models.Config) (string, error)
	GetByHash(ctx context.Context, hash string) (*models.Config, error)
}

package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// ConfigRepository is the port implemented by infra/repositories to persist manga configs.
// GetByHash returns nil, nil when the config doesn't exist.
type ConfigRepository interface {
	Create(ctx context.Context, config *models.Config) (string, error)
	GetByHash(ctx context.Context, hash string) (*models.Config, error)
}

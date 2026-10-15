package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// IConfigRepository is the port implemented by infra/repositories to persist manga configs.
// GetByHash returns nil, nil when the config doesn't exist.
type IConfigRepository interface {
	Create(ctx context.Context, config *models.Config) (string, error)
	GetByHash(ctx context.Context, hash string) (*models.Config, error)
}

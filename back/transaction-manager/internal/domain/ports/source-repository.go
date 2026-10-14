package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// SourceRepository is the port implemented by infra/repositories to persist sources.
// Getters return models.ErrSourceNotFound when the source doesn't exist.
type SourceRepository interface {
	Create(ctx context.Context, source *models.Source) (string, error)
	GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error)
	GetDetailByIDAndUserID(ctx context.Context, id string, userID int) (*models.SourceDetail, error)
	UpdateStatus(ctx context.Context, id string, userID int, status models.SourceStatus) error
	Delete(ctx context.Context, id string, userID int) error
}

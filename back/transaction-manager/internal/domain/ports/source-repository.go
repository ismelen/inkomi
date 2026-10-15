package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// SourceRepository is the port implemented by infra/repositories to persist sources.
// Getters return models.ErrSourceNotFound when the source doesn't exist.
type SourceRepository interface {
	Create(ctx context.Context, source *models.Source) (string, error)
	GetByIdAndUserId(ctx context.Context, id string, userId int) (*models.Source, error)
	GetDetailByIdAndUserId(ctx context.Context, id string, userId int) (*models.SourceDetail, error)
	UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error
	// UpdateErrorMsg stores errorMsg and marks the source as failed.
	UpdateErrorMsg(ctx context.Context, id string, userId int, errorMsg string) error
	Delete(ctx context.Context, id string, userId int) error
}

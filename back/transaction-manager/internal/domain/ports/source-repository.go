package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type SourceRepository interface {
	Create(ctx context.Context, source *models.Source) (string, error)
	CreateAll(ctx context.Context, sources []*models.Source) ([]string, error)
	GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error)
	Update(ctx context.Context, source *models.Source) error
	Delete(ctx context.Context, id string, userID int) error
}

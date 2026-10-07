package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type SourceRepository interface {
	Create(ctx context.Context, source *models.Source) (string, error)
	GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error)
	GetByIdAndUserIdCompact(ctx context.Context, id string, userID int) (*models.CompactSoruce, error)
	UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus, errMsg *string) error
	Delete(ctx context.Context, id string, userID int) error
}

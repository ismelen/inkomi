package ports

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

// ISourceRepository is the port implemented by infra/repositories to persist sources.
// Getters return models.ErrSourceNotFound when the source doesn't exist.
type ISourceRepository interface {
	Create(ctx context.Context, source *models.Source) (string, error)
	GetByIdAndUserId(ctx context.Context, id string, userId int) (*models.Source, error)
	GetDetailByIdAndUserId(ctx context.Context, id string, userId int) (*models.SourceDetail, error)
	// ListTreesByUserId returns up to filter.Limit root sources, newest first, with their children.
	// Trees where no node left pending_upload are skipped, and with UpdatedSince only trees with a node
	// updated at or after it are returned.
	ListTreesByUserId(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error)
	// Now returns the database clock, the one updatedAt is written with.
	Now(ctx context.Context) (time.Time, error)
	// UpdateStatus also sets completedAt when the status is terminal.
	UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error
	// UpdateFile stores the current filename and size of the source; empty values keep the stored ones.
	UpdateFile(ctx context.Context, id string, userId int, filename string, size int64) error
	// UpdateErrorMsg stores errorMsg and marks the source as failed.
	UpdateErrorMsg(ctx context.Context, id string, userId int, errorMsg string) error
	// UpsertResult creates the join result with its own id or updates its status, size and timestamps.
	UpsertResult(ctx context.Context, result *models.Source) error
	ListResultsByParentIdAndUserId(ctx context.Context, parentId string, userId int) ([]models.Source, error)
	Delete(ctx context.Context, id string, userId int) error
}

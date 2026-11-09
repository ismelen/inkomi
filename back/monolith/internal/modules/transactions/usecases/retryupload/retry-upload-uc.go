package retryupload

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
)

type RetryUploadUC struct {
	sourceRepo   ports.ISourceRepository
	cloudStorage ports.ICloudStorage
}

func NewRetryUploadUC(sourceRepo ports.ISourceRepository, cloudStorage ports.ICloudStorage) *RetryUploadUC {
	return &RetryUploadUC{sourceRepo: sourceRepo, cloudStorage: cloudStorage}
}

// Execute returns a new presigned PUT url for a file whose upload was never confirmed, so the
// client can upload it again and confirm it with PATCH /uploads/done/{id}. The source keeps its
// id, parent and position, so a folder being joined still gets all its children.
func (r *RetryUploadUC) Execute(ctx context.Context, userId int, sourceId string) (string, error) {
	src, err := r.sourceRepo.GetByIdAndUserId(ctx, sourceId, userId)
	if err != nil {
		return "", err
	}
	if !src.IsFile || src.IsLibrary() || src.Status != models.SourceStatusPendingUpload {
		return "", models.ErrSourceNotRetryable
	}
	return r.cloudStorage.GetUrl(src.Id, src.Filename)
}

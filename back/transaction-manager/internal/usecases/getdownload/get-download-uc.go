package getdownload

import (
	"context"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
)

type GetDownloadUC struct {
	sourceRepo   ports.ISourceRepository
	cloudStorage ports.ICloudStorage
}

func NewGetDownloadUC(sourceRepo ports.ISourceRepository, cloudStorage ports.ICloudStorage) *GetDownloadUC {
	return &GetDownloadUC{
		sourceRepo:   sourceRepo,
		cloudStorage: cloudStorage,
	}
}

func (u *GetDownloadUC) Execute(ctx context.Context, userId int, sourceId string) (*models.SourceDownload, error) {
	src, err := u.sourceRepo.GetByIdAndUserId(ctx, sourceId, userId)
	if err != nil {
		return nil, err
	}

	if !src.IsDownloadable() {
		return nil, models.ErrSourceNotReady
	}

	if !u.cloudStorage.Check(src.Id, src.Filename) {
		return nil, models.ErrFileNotAvailable
	}

	url, expiresAt, err := u.cloudStorage.GetDownloadUrl(src.Id, src.Filename)
	if err != nil {
		return nil, err
	}

	return &models.SourceDownload{
		Url:       url,
		Filename:  src.Filename,
		Size:      src.Size,
		ExpiresAt: expiresAt,
	}, nil
}

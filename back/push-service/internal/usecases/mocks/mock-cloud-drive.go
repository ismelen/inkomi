package mocks

import (
	"context"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
)

type MockCloudDrive struct {
	UploadFileFn func(ctx context.Context, account *models.DropboxAccount, path string, filename string) error
}

func (m *MockCloudDrive) UploadFile(ctx context.Context, account *models.DropboxAccount, path string, filename string) error {
	if m.UploadFileFn != nil {
		return m.UploadFileFn(ctx, account, path, filename)
	}
	return nil
}

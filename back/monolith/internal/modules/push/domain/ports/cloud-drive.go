package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/domain/models"
)

type ICloudDrive interface {
	// UploadFile stores the local file at path as filename in the account folder, replacing it
	// if it exists. It returns models.ErrUploadRejected when a retry can't succeed.
	UploadFile(ctx context.Context, account *models.DropboxAccount, path string, filename string) error
}

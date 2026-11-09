package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/models"
)

type IDropboxFiles interface {
	// ListFolders returns the subfolders of path ("" is the root) sorted by name, or
	// models.ErrDropboxFolderMissing when path doesn't exist.
	ListFolders(ctx context.Context, accessToken, path string) ([]models.DropboxFolder, error)
	// CreateFolder returns models.ErrDropboxFolderExists when something already has that path.
	CreateFolder(ctx context.Context, accessToken, path string) (*models.DropboxFolder, error)
}

package updatedropboxfolder

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type UpdateDropboxFolderUC struct {
	links ports.IDropboxLinkRepository
}

func NewUpdateDropboxFolderUC(links ports.IDropboxLinkRepository) *UpdateDropboxFolderUC {
	return &UpdateDropboxFolderUC{links: links}
}

// Execute sets the folder where the files of the user are uploaded and returns it normalized.
func (u *UpdateDropboxFolderUC) Execute(ctx context.Context, userId int, folder string) (string, error) {
	folder, err := models.NormalizeDropboxFolder(folder)
	if err != nil {
		return "", err
	}
	if err := u.links.UpdateFolder(ctx, userId, folder); err != nil {
		return "", err
	}
	return folder, nil
}

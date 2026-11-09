package createdropboxfolder

import (
	"context"

	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/usecases/getdropboxaccess"
)

type CreateDropboxFolderUC struct {
	access *getdropboxaccess.GetDropboxAccessUC
	files  ports.IDropboxFiles
}

func NewCreateDropboxFolderUC(access *getdropboxaccess.GetDropboxAccessUC, files ports.IDropboxFiles) *CreateDropboxFolderUC {
	return &CreateDropboxFolderUC{access: access, files: files}
}

// Execute creates the folder at path in the user's Dropbox; the root can't be created.
func (c *CreateDropboxFolderUC) Execute(ctx context.Context, userId int, path string) (*models.DropboxFolder, error) {
	path, err := models.NormalizeDropboxFolder(path)
	if err != nil {
		return nil, err
	}
	if path == "" {
		return nil, models.ErrInvalidDropboxFolder
	}
	link, err := c.access.Execute(ctx, userId)
	if err != nil {
		return nil, err
	}
	return c.files.CreateFolder(ctx, link.Access.Token, path)
}

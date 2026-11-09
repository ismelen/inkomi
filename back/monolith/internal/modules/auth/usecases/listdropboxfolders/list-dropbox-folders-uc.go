package listdropboxfolders

import (
	"context"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getdropboxaccess"
)

type ListDropboxFoldersUC struct {
	access *getdropboxaccess.GetDropboxAccessUC
	files  ports.IDropboxFiles
}

func NewListDropboxFoldersUC(access *getdropboxaccess.GetDropboxAccessUC, files ports.IDropboxFiles) *ListDropboxFoldersUC {
	return &ListDropboxFoldersUC{access: access, files: files}
}

// Execute returns the normalized folder ("" is the root) and its subfolders in the user's Dropbox.
func (l *ListDropboxFoldersUC) Execute(ctx context.Context, userId int, folder string) (string, []models.DropboxFolder, error) {
	folder, err := models.NormalizeDropboxFolder(folder)
	if err != nil {
		return "", nil, err
	}
	link, err := l.access.Execute(ctx, userId)
	if err != nil {
		return "", nil, err
	}
	folders, err := l.files.ListFolders(ctx, link.Access.Token, folder)
	if err != nil {
		return "", nil, err
	}
	return folder, folders, nil
}

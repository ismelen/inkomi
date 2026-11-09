package createdropboxfolder_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/createdropboxfolder"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getdropboxaccess"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func access() *getdropboxaccess.GetDropboxAccessUC {
	links := &mocks.MockDropboxLinkRepository{
		GetFn: func(ctx context.Context, userId int) (*models.DropboxLink, error) {
			return &models.DropboxLink{UserId: userId, Access: models.AccessToken{Token: "tok", ExpiresAt: time.Now().Add(time.Hour)}}, nil
		},
	}
	return getdropboxaccess.NewGetDropboxAccessUC(links, &mocks.MockDropboxOAuth{})
}

func TestCreateDropboxFolderUC_Execute_Path_CreatesNormalizedPath(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotPath string
	files := &mocks.MockDropboxFiles{CreateFolderFn: func(ctx context.Context, token, path string) (*models.DropboxFolder, error) {
		gotPath = path
		return &models.DropboxFolder{Name: "Kobo", Path: path}, nil
	}}
	uc := createdropboxfolder.NewCreateDropboxFolderUC(access(), files)

	// Act
	folder, err := uc.Execute(context.Background(), 7, "Books/Kobo/")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "/Books/Kobo", gotPath)
	assert.Equal(t, "/Books/Kobo", folder.Path)
}

func TestCreateDropboxFolderUC_Execute_Root_ReturnsInvalidFolder(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := createdropboxfolder.NewCreateDropboxFolderUC(access(), &mocks.MockDropboxFiles{})

	// Act
	_, err := uc.Execute(context.Background(), 7, "/")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidDropboxFolder)
}

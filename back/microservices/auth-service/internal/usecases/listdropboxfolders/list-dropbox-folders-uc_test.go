package listdropboxfolders_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/usecases/getdropboxaccess"
	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/usecases/listdropboxfolders"
	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/usecases/mocks"
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

func TestListDropboxFoldersUC_Execute_Path_ListsNormalizedPathWithToken(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotToken, gotPath string
	files := &mocks.MockDropboxFiles{ListFoldersFn: func(ctx context.Context, token, path string) ([]models.DropboxFolder, error) {
		gotToken, gotPath = token, path
		return []models.DropboxFolder{{Name: "Kobo", Path: "/Books/Kobo"}}, nil
	}}
	uc := listdropboxfolders.NewListDropboxFoldersUC(access(), files)

	// Act
	folder, folders, err := uc.Execute(context.Background(), 7, " Books/ ")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "/Books", folder)
	assert.Equal(t, "/Books", gotPath)
	assert.Equal(t, "tok", gotToken)
	assert.Equal(t, []models.DropboxFolder{{Name: "Kobo", Path: "/Books/Kobo"}}, folders)
}

func TestListDropboxFoldersUC_Execute_InvalidPath_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := listdropboxfolders.NewListDropboxFoldersUC(access(), &mocks.MockDropboxFiles{})

	// Act
	_, _, err := uc.Execute(context.Background(), 7, "/a/../b")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidDropboxFolder)
}

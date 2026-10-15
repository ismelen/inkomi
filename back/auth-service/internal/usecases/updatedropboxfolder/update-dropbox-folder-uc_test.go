package updatedropboxfolder_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/updatedropboxfolder"
	"github.com/stretchr/testify/assert"
)

func TestUpdateDropboxFolderUC_Execute_Folder_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		folder string
		update func(ctx context.Context, userId int, folder string) error
		want   string
		err    error
	}{
		{"Normalized", "Inkomi/Books/", nil, "/Inkomi/Books", nil},
		{"Invalid", "/a/../b", nil, "", models.ErrInvalidDropboxFolder},
		{"Not linked", "/Inkomi", func(ctx context.Context, userId int, folder string) error {
			return models.ErrDropboxNotLinked
		}, "", models.ErrDropboxNotLinked},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			links := &mocks.MockDropboxLinkRepository{UpdateFolderFn: tc.update}

			// Act
			folder, err := updatedropboxfolder.NewUpdateDropboxFolderUC(links).Execute(context.Background(), 1, tc.folder)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.want, folder)
		})
	}
}

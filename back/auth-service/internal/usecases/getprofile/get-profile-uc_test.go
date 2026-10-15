package getprofile_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getprofile"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProfileUC_Execute_DropboxLink_Profile(t *testing.T) {
	t.Parallel()

	link := &models.DropboxLink{UserId: 3, Folder: "/Inkomi"}
	cases := []struct {
		name  string
		links *mocks.MockDropboxLinkRepository
		want  *models.DropboxLink
	}{
		{"Linked", &mocks.MockDropboxLinkRepository{
			GetFn: func(ctx context.Context, userId int) (*models.DropboxLink, error) { return link, nil },
		}, link},
		{"Not linked", &mocks.MockDropboxLinkRepository{}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			users := &mocks.MockUserRepository{
				GetByIdFn: func(ctx context.Context, id int) (*models.User, error) { return &models.User{Id: id}, nil },
			}
			identities := &mocks.MockIdentityRepository{
				ListProvidersFn: func(ctx context.Context, userId int) ([]models.Provider, error) {
					return []models.Provider{models.ProviderGoogle}, nil
				},
			}

			// Act
			profile, err := getprofile.NewGetProfileUC(users, identities, tc.links).Execute(context.Background(), 3)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, &models.Profile{User: &models.User{Id: 3}, Providers: []models.Provider{models.ProviderGoogle}, Dropbox: tc.want}, profile)
		})
	}
}

func TestGetProfileUC_Execute_UnknownUser_ReturnsErrUserNotFound(t *testing.T) {
	t.Parallel()

	// Act
	profile, err := getprofile.NewGetProfileUC(&mocks.MockUserRepository{}, &mocks.MockIdentityRepository{}, &mocks.MockDropboxLinkRepository{}).
		Execute(context.Background(), 3)

	// Assert
	assert.ErrorIs(t, err, models.ErrUserNotFound)
	assert.Nil(t, profile)
}

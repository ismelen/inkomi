package linkdropbox_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/linkdropbox"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validOAuth() *mocks.MockDropboxOAuth {
	return &mocks.MockDropboxOAuth{
		RefreshFn: func(ctx context.Context, refreshToken string) (models.AccessToken, error) {
			return models.AccessToken{Token: "access"}, nil
		},
		AccountFn: func(ctx context.Context, accessToken string) (*models.ExternalAccount, error) {
			return &models.ExternalAccount{Provider: models.ProviderDropbox, Subject: "dbid:1", Email: "a@b.com"}, nil
		},
	}
}

func identityOf(userId int) *mocks.MockIdentityRepository {
	return &mocks.MockIdentityRepository{
		GetFn: func(ctx context.Context, provider models.Provider, subject string) (*models.Identity, error) {
			return &models.Identity{Provider: provider, Subject: subject, UserId: userId}, nil
		},
	}
}

func TestLinkDropboxUC_Execute_Identity_SavesLink(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		identities *mocks.MockIdentityRepository
		created    int
	}{
		{"New account creates identity", &mocks.MockIdentityRepository{}, 1},
		{"Account of same user", identityOf(3), 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			created := 0
			tc.identities.CreateFn = func(ctx context.Context, identity models.Identity) error {
				assert.Equal(t, models.Identity{Provider: models.ProviderDropbox, Subject: "dbid:1", UserId: 3}, identity)
				created++
				return nil
			}
			var saved *models.DropboxLink
			links := &mocks.MockDropboxLinkRepository{
				SaveFn: func(ctx context.Context, link *models.DropboxLink) error {
					saved = link
					return nil
				},
			}

			// Act
			err := linkdropbox.NewLinkDropboxUC(validOAuth(), tc.identities, links).Execute(context.Background(), 3, "dbx-refresh")

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.created, created)
			require.NotNil(t, saved)
			assert.Equal(t, 3, saved.UserId)
			assert.Equal(t, "dbx-refresh", saved.RefreshToken)
			assert.Equal(t, "access", saved.Access.Token)
		})
	}
}

func TestLinkDropboxUC_Execute_Failures_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		oauth      *mocks.MockDropboxOAuth
		identities *mocks.MockIdentityRepository
		err        error
	}{
		{"Account of another user", validOAuth(), identityOf(9), models.ErrDropboxAccountTaken},
		{"Revoked token", &mocks.MockDropboxOAuth{}, &mocks.MockIdentityRepository{}, models.ErrInvalidExternalToken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			saved := false
			links := &mocks.MockDropboxLinkRepository{
				SaveFn: func(ctx context.Context, link *models.DropboxLink) error {
					saved = true
					return nil
				},
			}

			// Act
			err := linkdropbox.NewLinkDropboxUC(tc.oauth, tc.identities, links).Execute(context.Background(), 3, "dbx-refresh")

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, saved)
		})
	}
}

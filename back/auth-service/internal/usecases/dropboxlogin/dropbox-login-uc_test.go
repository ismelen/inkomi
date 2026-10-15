package dropboxlogin_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/dropboxlogin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var dropboxAccount = models.ExternalAccount{Provider: models.ProviderDropbox, Subject: "dbid:1", Email: "a@b.com", EmailVerified: true}

func newUC(oauth *mocks.MockDropboxOAuth, links *mocks.MockDropboxLinkRepository) *dropboxlogin.DropboxLoginUC {
	signIn := socialsignin.NewSocialSignInUC(&mocks.MockTxManager{}, &mocks.MockUserRepository{}, &mocks.MockIdentityRepository{})
	startSession := startsession.NewStartSessionUC(&mocks.MockSessionRepository{}, &mocks.MockTokenIssuer{}, time.Hour)
	return dropboxlogin.NewDropboxLoginUC(oauth, links, signIn, startSession)
}

func validOAuth() *mocks.MockDropboxOAuth {
	return &mocks.MockDropboxOAuth{
		RefreshFn: func(ctx context.Context, refreshToken string) (models.AccessToken, error) {
			return models.AccessToken{Token: "access-for-" + refreshToken}, nil
		},
		AccountFn: func(ctx context.Context, accessToken string) (*models.ExternalAccount, error) {
			account := dropboxAccount
			return &account, nil
		},
	}
}

func TestDropboxLoginUC_Execute_ValidRefreshToken_LinksDropboxAndReturnsTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved *models.DropboxLink
	links := &mocks.MockDropboxLinkRepository{
		SaveFn: func(ctx context.Context, link *models.DropboxLink) error {
			saved = link
			return nil
		},
	}

	// Act
	pair, err := newUC(validOAuth(), links).Execute(context.Background(), "dbx-refresh", "d")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "access:1", pair.AccessToken)
	assert.Equal(t, &models.DropboxLink{
		UserId:       1,
		AccountId:    "dbid:1",
		Email:        "a@b.com",
		RefreshToken: "dbx-refresh",
		Access:       models.AccessToken{Token: "access-for-dbx-refresh"},
	}, saved)
}

func TestDropboxLoginUC_Execute_InvalidInput_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		oauth    *mocks.MockDropboxOAuth
		deviceId string
		err      error
	}{
		{"Revoked token", &mocks.MockDropboxOAuth{}, "d", models.ErrInvalidExternalToken},
		{"No device", validOAuth(), "", models.ErrInvalidDeviceId},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			pair, err := newUC(tc.oauth, &mocks.MockDropboxLinkRepository{}).Execute(context.Background(), "dbx-refresh", tc.deviceId)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Nil(t, pair)
		})
	}
}

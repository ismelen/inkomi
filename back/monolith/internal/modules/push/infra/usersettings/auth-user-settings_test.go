package usersettings_test

import (
	"context"
	"errors"
	"testing"

	authmodels "github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/infra/usersettings"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type getDropboxAccessFn func(ctx context.Context, userId int) (*authmodels.DropboxLink, error)

func (f getDropboxAccessFn) Execute(ctx context.Context, userId int) (*authmodels.DropboxLink, error) {
	return f(ctx, userId)
}

type getFCMTokenFn func(ctx context.Context, userId int) (string, error)

func (f getFCMTokenFn) Execute(ctx context.Context, userId int) (string, error) {
	return f(ctx, userId)
}

func TestAuthUserSettings_GetDropboxAccount_LinkedUser_ReturnsNormalizedAccount(t *testing.T) {
	t.Parallel()

	// Arrange
	settings := usersettings.NewAuthUserSettings(
		getDropboxAccessFn(func(ctx context.Context, userId int) (*authmodels.DropboxLink, error) {
			return &authmodels.DropboxLink{Access: authmodels.AccessToken{Token: "tok"}, Folder: "books/"}, nil
		}),
		nil,
	)

	// Act
	account, err := settings.GetDropboxAccount(context.Background(), 7)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, &models.DropboxAccount{Token: "tok", Folder: "/books"}, account)
}

func TestAuthUserSettings_GetDropboxAccount_Errors_AreTranslated(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"Not linked", authmodels.ErrDropboxNotLinked, models.ErrDropboxNotLinked},
		{"Unknown user", authmodels.ErrUserNotFound, models.ErrDropboxNotLinked},
		{"Other error", boom, boom},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			settings := usersettings.NewAuthUserSettings(
				getDropboxAccessFn(func(ctx context.Context, userId int) (*authmodels.DropboxLink, error) { return nil, tc.err }),
				nil,
			)

			// Act
			_, err := settings.GetDropboxAccount(context.Background(), 7)

			// Assert
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestAuthUserSettings_GetFCMToken_Cases(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	cases := []struct {
		name    string
		token   string
		err     error
		want    string
		wantErr error
	}{
		{"Registered device", "fcm", nil, "fcm", nil},
		{"Empty token", "", nil, "", models.ErrFCMTokenNotFound},
		{"Not found", "", authmodels.ErrFCMTokenNotFound, "", models.ErrFCMTokenNotFound},
		{"Unknown user", "", authmodels.ErrUserNotFound, "", models.ErrFCMTokenNotFound},
		{"Other error", "", boom, "", boom},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			settings := usersettings.NewAuthUserSettings(
				nil,
				getFCMTokenFn(func(ctx context.Context, userId int) (string, error) { return tc.token, tc.err }),
			)

			// Act
			token, err := settings.GetFCMToken(context.Background(), 7)

			// Assert
			assert.Equal(t, tc.want, token)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

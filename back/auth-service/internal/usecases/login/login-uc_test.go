package login_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/login"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUC(users *mocks.MockUserRepository, hasher *mocks.MockPasswordHasher) *login.LoginUC {
	startSession := startsession.NewStartSessionUC(&mocks.MockSessionRepository{}, &mocks.MockTokenIssuer{}, time.Hour)
	return login.NewLoginUC(users, hasher, startSession)
}

func usersWith(user *models.User) *mocks.MockUserRepository {
	return &mocks.MockUserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			if user == nil || email != user.Email {
				return nil, models.ErrUserNotFound
			}
			return user, nil
		},
	}
}

func TestLoginUC_Execute_RightPassword_ReturnsTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	user := &models.User{Id: 4, Email: "a@b.com", PasswordHash: "hash:secret-password"}

	// Act
	pair, err := newUC(usersWith(user), &mocks.MockPasswordHasher{}).Execute(context.Background(), "A@B.com", "secret-password", "d")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "access:4", pair.AccessToken)
}

func TestLoginUC_Execute_WrongCredentials_ReturnsErrInvalidCredentials(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		user     *models.User
		email    string
		password string
		hashes   int
	}{
		{"Wrong password", &models.User{Email: "a@b.com", PasswordHash: "hash:other-password"}, "a@b.com", "secret-password", 0},
		{"Unknown email", nil, "a@b.com", "secret-password", 1},
		{"Social account without password", &models.User{Email: "a@b.com"}, "a@b.com", "secret-password", 1},
		{"Invalid email", nil, "nope", "secret-password", 0},
		{"Too long password", nil, "a@b.com", string(make([]byte, 200)), 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			hashes := 0
			hasher := &mocks.MockPasswordHasher{
				HashFn: func(password string) (string, error) {
					hashes++
					return "hash:" + password, nil
				},
			}

			// Act
			pair, err := newUC(usersWith(tc.user), hasher).Execute(context.Background(), tc.email, tc.password, "d")

			// Assert
			assert.ErrorIs(t, err, models.ErrInvalidCredentials)
			assert.Nil(t, pair)
			assert.Equal(t, tc.hashes, hashes)
		})
	}
}

func TestLoginUC_Execute_NoDevice_ReturnsErrInvalidDeviceId(t *testing.T) {
	t.Parallel()

	// Act
	_, err := newUC(usersWith(nil), &mocks.MockPasswordHasher{}).Execute(context.Background(), "a@b.com", "secret-password", "")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidDeviceId)
}

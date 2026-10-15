package register_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/register"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUC(users *mocks.MockUserRepository) *register.RegisterUC {
	startSession := startsession.NewStartSessionUC(&mocks.MockSessionRepository{}, &mocks.MockTokenIssuer{}, time.Hour)
	return register.NewRegisterUC(users, &mocks.MockPasswordHasher{}, startSession)
}

func TestRegisterUC_Execute_ValidInput_CreatesUserAndSession(t *testing.T) {
	t.Parallel()

	// Arrange
	var created *models.User
	users := &mocks.MockUserRepository{
		CreateFn: func(ctx context.Context, user *models.User) error {
			user.Id = 3
			created = user
			return nil
		},
	}

	// Act
	pair, err := newUC(users).Execute(context.Background(), " User@Example.com", "secret-password", "device-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "access:3", pair.AccessToken)
	assert.Equal(t, &models.User{Id: 3, Email: "user@example.com", PasswordHash: "hash:secret-password", Role: models.RoleUser}, created)
}

func TestRegisterUC_Execute_InvalidInput_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		email    string
		password string
		deviceId string
		create   func(ctx context.Context, user *models.User) error
		err      error
	}{
		{"Invalid email", "nope", "secret-password", "d", nil, models.ErrInvalidEmail},
		{"Weak password", "a@b.com", "short", "d", nil, models.ErrWeakPassword},
		{"No device", "a@b.com", "secret-password", "", nil, models.ErrInvalidDeviceId},
		{"Email taken", "a@b.com", "secret-password", "d", func(ctx context.Context, user *models.User) error {
			return models.ErrEmailTaken
		}, models.ErrEmailTaken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			users := &mocks.MockUserRepository{CreateFn: tc.create}

			// Act
			pair, err := newUC(users).Execute(context.Background(), tc.email, tc.password, tc.deviceId)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Nil(t, pair)
		})
	}
}

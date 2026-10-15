package refreshsession_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/refreshsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var existingUser = &mocks.MockUserRepository{
	GetByIdFn: func(ctx context.Context, id int) (*models.User, error) {
		return &models.User{Id: id}, nil
	},
}

func TestRefreshSessionUC_Execute_ValidToken_RotatesAndReturnsTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	var oldHash, newHash string
	sessions := &mocks.MockSessionRepository{
		RotateFn: func(ctx context.Context, old, new string, now, expiresAt time.Time) (int, error) {
			oldHash, newHash = old, new
			return 8, nil
		},
	}
	uc := refreshsession.NewRefreshSessionUC(sessions, existingUser, &mocks.MockTokenIssuer{}, time.Hour)

	// Act
	pair, err := uc.Execute(context.Background(), "refresh-old")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "hash:refresh-old", oldHash)
	assert.Equal(t, "hash:refresh-1", newHash)
	assert.Equal(t, "refresh-1", pair.RefreshToken)
	assert.Equal(t, "access:8", pair.AccessToken)
	assert.WithinDuration(t, time.Now().Add(time.Hour), pair.RefreshExpiresAt, time.Minute)
}

func TestRefreshSessionUC_Execute_InvalidToken_ReturnsErrInvalidRefreshToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		token    string
		sessions *mocks.MockSessionRepository
		users    *mocks.MockUserRepository
	}{
		{"Empty", "", &mocks.MockSessionRepository{}, existingUser},
		{"Unknown, used or expired", "refresh-old", &mocks.MockSessionRepository{}, existingUser},
		{"Deleted user", "refresh-old", &mocks.MockSessionRepository{
			RotateFn: func(ctx context.Context, old, new string, now, expiresAt time.Time) (int, error) { return 8, nil },
		}, &mocks.MockUserRepository{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := refreshsession.NewRefreshSessionUC(tc.sessions, tc.users, &mocks.MockTokenIssuer{}, time.Hour)

			// Act
			pair, err := uc.Execute(context.Background(), tc.token)

			// Assert
			assert.ErrorIs(t, err, models.ErrInvalidRefreshToken)
			assert.Nil(t, pair)
		})
	}
}

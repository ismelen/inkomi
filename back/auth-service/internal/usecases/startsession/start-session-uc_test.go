package startsession_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartSessionUC_Execute_ValidDevice_SavesSessionAndReturnsTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved models.Session
	sessions := &mocks.MockSessionRepository{
		SaveFn: func(ctx context.Context, session models.Session) error {
			saved = session
			return nil
		},
	}
	uc := startsession.NewStartSessionUC(sessions, &mocks.MockTokenIssuer{}, time.Hour)

	// Act
	pair, err := uc.Execute(context.Background(), &models.User{Id: 7}, "device-1")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "access:7", pair.AccessToken)
	assert.Equal(t, "refresh-1", pair.RefreshToken)
	assert.Equal(t, models.Session{UserId: 7, DeviceId: "device-1", RefreshHash: "hash:refresh-1", ExpiresAt: pair.RefreshExpiresAt}, saved)
	assert.WithinDuration(t, time.Now().Add(time.Hour), pair.RefreshExpiresAt, time.Minute)
}

func TestStartSessionUC_Execute_Failures_ReturnsError(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")
	cases := []struct {
		name     string
		deviceId string
		sessions *mocks.MockSessionRepository
		issuer   *mocks.MockTokenIssuer
		err      error
	}{
		{"Empty device", "", &mocks.MockSessionRepository{}, &mocks.MockTokenIssuer{}, models.ErrInvalidDeviceId},
		{"Save fails", "d", &mocks.MockSessionRepository{
			SaveFn: func(ctx context.Context, session models.Session) error { return boom },
		}, &mocks.MockTokenIssuer{}, boom},
		{"Issue fails", "d", &mocks.MockSessionRepository{}, &mocks.MockTokenIssuer{
			IssueAccessFn: func(user *models.User) (models.AccessToken, error) { return models.AccessToken{}, boom },
		}, boom},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := startsession.NewStartSessionUC(tc.sessions, tc.issuer, time.Hour)

			// Act
			pair, err := uc.Execute(context.Background(), &models.User{Id: 7}, tc.deviceId)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Nil(t, pair)
		})
	}
}

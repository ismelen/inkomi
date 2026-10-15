package googlelogin_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/googlelogin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newUC(verifier *mocks.MockGoogleVerifier) *googlelogin.GoogleLoginUC {
	signIn := socialsignin.NewSocialSignInUC(&mocks.MockTxManager{}, &mocks.MockUserRepository{}, &mocks.MockIdentityRepository{})
	startSession := startsession.NewStartSessionUC(&mocks.MockSessionRepository{}, &mocks.MockTokenIssuer{}, time.Hour)
	return googlelogin.NewGoogleLoginUC(verifier, signIn, startSession)
}

func TestGoogleLoginUC_Execute_ValidToken_ReturnsTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	verifier := &mocks.MockGoogleVerifier{
		VerifyFn: func(ctx context.Context, idToken string) (*models.ExternalAccount, error) {
			require.Equal(t, "id-token", idToken)
			return &models.ExternalAccount{Provider: models.ProviderGoogle, Subject: "g-1", Email: "a@b.com", EmailVerified: true}, nil
		},
	}

	// Act
	pair, err := newUC(verifier).Execute(context.Background(), "id-token", "d")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "access:1", pair.AccessToken)
}

func TestGoogleLoginUC_Execute_InvalidInput_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		deviceId string
		err      error
	}{
		{"Invalid token", "d", models.ErrInvalidExternalToken},
		{"No device", "", models.ErrInvalidDeviceId},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			pair, err := newUC(&mocks.MockGoogleVerifier{}).Execute(context.Background(), "id-token", tc.deviceId)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Nil(t, pair)
		})
	}
}

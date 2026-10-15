package google

import (
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/idtoken"
)

// Validating a real token needs Google's signing keys, so only the claim checks are tested.
func TestAccountFromPayload_Payload_Result(t *testing.T) {
	t.Parallel()

	claims := map[string]any{"email": "a@b.com", "email_verified": true}
	cases := []struct {
		name    string
		payload idtoken.Payload
		account *models.ExternalAccount
	}{
		{"Valid", idtoken.Payload{Issuer: "https://accounts.google.com", Audience: "android", Subject: "g-1", Claims: claims},
			&models.ExternalAccount{Provider: models.ProviderGoogle, Subject: "g-1", Email: "a@b.com", EmailVerified: true}},
		{"Unverified email", idtoken.Payload{Issuer: "accounts.google.com", Audience: "ios", Subject: "g-1", Claims: map[string]any{"email": "a@b.com"}},
			&models.ExternalAccount{Provider: models.ProviderGoogle, Subject: "g-1", Email: "a@b.com"}},
		{"Other app", idtoken.Payload{Issuer: "accounts.google.com", Audience: "other", Subject: "g-1", Claims: claims}, nil},
		{"Other issuer", idtoken.Payload{Issuer: "evil.com", Audience: "android", Subject: "g-1", Claims: claims}, nil},
		{"No subject", idtoken.Payload{Issuer: "accounts.google.com", Audience: "android", Claims: claims}, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			account, err := accountFromPayload(&tc.payload, []string{"android", "ios"})

			// Assert
			assert.Equal(t, tc.account, account)
			if tc.account == nil {
				assert.ErrorIs(t, err, models.ErrInvalidExternalToken)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

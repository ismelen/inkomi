package httpauth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	return key, string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func signToken(t *testing.T, key *rsa.PrivateKey, expiresAt time.Time) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, httpauth.UserClaims{
		Id:               7,
		Role:             "user",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expiresAt)},
	})
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func TestNewAuthMiddleware_InvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	middleware, err := httpauth.NewAuthMiddleware("not a pem")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, middleware)
}

func TestAuthMiddleware_Request_Authorization_Status(t *testing.T) {
	t.Parallel()

	key, publicPEM := newKeyPair(t)
	otherKey, _ := newKeyPair(t)

	cases := []struct {
		name     string
		header   string
		protocol string
		status   int
	}{
		{"No header", "", "", http.StatusUnauthorized},
		{"Not bearer", "Basic abc", "", http.StatusUnauthorized},
		{"Wrong signature", "Bearer " + signToken(t, otherKey, time.Now().Add(time.Hour)), "", http.StatusUnauthorized},
		{"Expired", "Bearer " + signToken(t, key, time.Now().Add(-time.Hour)), "", http.StatusUnauthorized},
		{"Valid", "Bearer " + signToken(t, key, time.Now().Add(time.Hour)), "", http.StatusOK},
		{"Protocol without token", "", "bearer", http.StatusUnauthorized},
		{"Protocol not bearer", "", "chat, " + signToken(t, key, time.Now().Add(time.Hour)), http.StatusUnauthorized},
		{"Protocol expired", "", "bearer, " + signToken(t, key, time.Now().Add(-time.Hour)), http.StatusUnauthorized},
		{"Protocol valid", "", "bearer, " + signToken(t, key, time.Now().Add(time.Hour)), http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			middleware, err := httpauth.NewAuthMiddleware(publicPEM)
			require.NoError(t, err)

			var userId int
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, ok := httpauth.GetUserClaims(r.Context())
				require.True(t, ok)
				userId = claims.Id
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			if tc.protocol != "" {
				req.Header.Set("Sec-WebSocket-Protocol", tc.protocol)
			}
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tc.status, rec.Code)
			if tc.status == http.StatusOK {
				assert.Equal(t, 7, userId)
			}
		})
	}
}

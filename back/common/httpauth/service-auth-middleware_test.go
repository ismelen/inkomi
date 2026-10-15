package httpauth_test

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSigner(t *testing.T, service string, ttl time.Duration) (*httpauth.ServiceTokenSigner, string) {
	t.Helper()

	key, publicPEM := newKeyPair(t)
	privatePEM := string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}))
	signer, err := httpauth.NewServiceTokenSigner(service, strings.ReplaceAll(privatePEM, "\n", "\n"), ttl)
	require.NoError(t, err)
	return signer, publicPEM
}

func sign(t *testing.T, signer *httpauth.ServiceTokenSigner, audience string) string {
	t.Helper()

	token, err := signer.Sign(audience)
	require.NoError(t, err)
	return "Bearer " + token
}

func TestNewServiceTokenSigner_InvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	signer, err := httpauth.NewServiceTokenSigner("push-service", "not a pem", time.Minute)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, signer)
}

func TestNewServiceAuthMiddleware_InvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	middleware, err := httpauth.NewServiceAuthMiddleware("auth-service", map[string]string{"push-service": "not a pem"})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, middleware)
}

func TestServiceAuthMiddleware_Request_Authorization_Status(t *testing.T) {
	t.Parallel()

	pushSigner, pushPEM := newSigner(t, "push-service", time.Minute)
	_, otherPEM := newSigner(t, "other-service", time.Minute)
	unknownSigner, _ := newSigner(t, "unknown-service", time.Minute)
	expiredSigner, expiredPEM := newSigner(t, "expired-service", -time.Minute)
	impostorSigner, _ := newSigner(t, "push-service", time.Minute)

	publicKeys := map[string]string{
		"push-service":    pushPEM,
		"other-service":   otherPEM,
		"expired-service": expiredPEM,
	}

	cases := []struct {
		name    string
		header  string
		status  int
		service string
	}{
		{"No header", "", http.StatusUnauthorized, ""},
		{"Not bearer", "Basic abc", http.StatusUnauthorized, ""},
		{"Unknown service", sign(t, unknownSigner, "auth-service"), http.StatusUnauthorized, ""},
		{"Signed with another service key", sign(t, impostorSigner, "auth-service"), http.StatusUnauthorized, ""},
		{"Wrong audience", sign(t, pushSigner, "library-service"), http.StatusUnauthorized, ""},
		{"Expired", sign(t, expiredSigner, "auth-service"), http.StatusUnauthorized, ""},
		{"Valid", sign(t, pushSigner, "auth-service"), http.StatusOK, "push-service"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			middleware, err := httpauth.NewServiceAuthMiddleware("auth-service", publicKeys)
			require.NoError(t, err)

			var service string
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, ok := httpauth.GetServiceClaims(r.Context())
				require.True(t, ok)
				service = claims.Subject
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tc.status, rec.Code)
			assert.Equal(t, tc.service, service)
		})
	}
}

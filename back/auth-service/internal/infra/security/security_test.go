package security_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/security"
	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArgon2PasswordHasher_Verify_Password_Result(t *testing.T) {
	t.Parallel()

	// Arrange
	hasher := security.NewArgon2PasswordHasher()
	hash, err := hasher.Hash("secret-password")
	require.NoError(t, err)

	cases := []struct {
		name     string
		password string
		hash     string
		ok       bool
		err      error
	}{
		{"Right password", "secret-password", hash, true, nil},
		{"Wrong password", "other-password", hash, false, nil},
		{"Malformed hash", "secret-password", "$bcrypt$x", false, security.ErrInvalidHash},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			ok, err := hasher.Verify(tc.password, tc.hash)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.ok, ok)
		})
	}
}

func TestArgon2PasswordHasher_Hash_SamePassword_DifferentSalt(t *testing.T) {
	t.Parallel()

	// Arrange
	hasher := security.NewArgon2PasswordHasher()

	// Act
	first, err1 := hasher.Hash("secret-password")
	second, err2 := hasher.Hash("secret-password")

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, first, second)
	assert.True(t, strings.HasPrefix(first, "$argon2id$v=19$m=19456,t=2,p=1$"))
}

func newPrivateKeyPEM(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	privatePEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return key, strings.ReplaceAll(string(privatePEM), "\n", "\\n")
}

func TestJWTTokenIssuer_IssueAccess_User_ClaimsReadableByHttpauth(t *testing.T) {
	t.Parallel()

	// Arrange
	key, privatePEM := newPrivateKeyPEM(t)
	issuer, err := security.NewJWTTokenIssuer("auth-service", privatePEM, 15*time.Minute)
	require.NoError(t, err)

	// Act
	access, err := issuer.IssueAccess(&models.User{Id: 42, Role: models.RoleUser})

	// Assert
	require.NoError(t, err)
	claims := &httpauth.UserClaims{}
	_, err = jwt.ParseWithClaims(access.Token, claims, func(token *jwt.Token) (any, error) { return &key.PublicKey, nil })
	require.NoError(t, err)
	assert.Equal(t, 42, claims.Id)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "42", claims.Subject)
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), access.ExpiresAt, time.Minute)
	assert.Equal(t, access.ExpiresAt.Unix(), claims.ExpiresAt.Unix())
}

func TestJWTTokenIssuer_NewRefresh_Token_HashMatches(t *testing.T) {
	t.Parallel()

	// Arrange
	_, privatePEM := newPrivateKeyPEM(t)
	issuer, err := security.NewJWTTokenIssuer("auth-service", privatePEM, time.Minute)
	require.NoError(t, err)

	// Act
	token, hash, err := issuer.NewRefresh()
	other, _, _ := issuer.NewRefresh()

	// Assert
	require.NoError(t, err)
	assert.Len(t, token, 43)
	assert.Equal(t, issuer.HashRefresh(token), hash)
	assert.NotEqual(t, token, hash)
	assert.NotEqual(t, token, other)
}

func TestNewJWTTokenIssuer_InvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()

	// Act
	issuer, err := security.NewJWTTokenIssuer("auth-service", "not a pem", time.Minute)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, issuer)
}

func newKey(t *testing.T) string {
	t.Helper()

	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(key)
}

func TestAESTokenCipher_Decrypt_Ciphertext_Result(t *testing.T) {
	t.Parallel()

	// Arrange
	cipher, err := security.NewAESTokenCipher(newKey(t))
	require.NoError(t, err)
	other, err := security.NewAESTokenCipher(newKey(t))
	require.NoError(t, err)

	encrypted, err := cipher.Encrypt("dropbox-refresh-token")
	require.NoError(t, err)
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	require.NoError(t, err)
	raw[len(raw)-1] ^= 1
	tampered := base64.StdEncoding.EncodeToString(raw)

	// Act
	plain, err := cipher.Decrypt(encrypted)
	_, otherErr := other.Decrypt(encrypted)
	_, tamperedErr := cipher.Decrypt(tampered)
	_, garbageErr := cipher.Decrypt("garbage")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "dropbox-refresh-token", plain)
	assert.NotContains(t, encrypted, "dropbox")
	assert.ErrorIs(t, otherErr, security.ErrInvalidCiphertext)
	assert.ErrorIs(t, tamperedErr, security.ErrInvalidCiphertext)
	assert.ErrorIs(t, garbageErr, security.ErrInvalidCiphertext)
}

func TestNewAESTokenCipher_InvalidKey_ReturnsError(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		key  string
	}{
		{"Not base64", "%%%"},
		{"Short", base64.StdEncoding.EncodeToString([]byte("short"))},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			cipher, err := security.NewAESTokenCipher(tc.key)

			// Assert
			assert.Error(t, err)
			assert.Nil(t, cipher)
		})
	}
}

package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/common/httpauth"
)

var _ ports.ITokenIssuer = (*JWTTokenIssuer)(nil)

const refreshTokenBytes = 32

// JWTTokenIssuer signs RS256 access tokens with the claims that httpauth.NewAuthMiddleware
// reads, and issues opaque random refresh tokens.
type JWTTokenIssuer struct {
	issuer     string
	privateKey *rsa.PrivateKey
	accessTTL  time.Duration
}

// NewJWTTokenIssuer signs with privateKeyPEM. Literal "\n" in the PEM are accepted.
func NewJWTTokenIssuer(issuer, privateKeyPEM string, accessTTL time.Duration) (*JWTTokenIssuer, error) {
	pemStr := strings.ReplaceAll(privateKeyPEM, "\\n", "\n")
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemStr))
	if err != nil {
		return nil, fmt.Errorf("parse jwt private key: %w", err)
	}
	return &JWTTokenIssuer{issuer: issuer, privateKey: privateKey, accessTTL: accessTTL}, nil
}

func (j *JWTTokenIssuer) IssueAccess(user *models.User) (models.AccessToken, error) {
	now := time.Now()
	expiresAt := now.Add(j.accessTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, httpauth.UserClaims{
		Id:   user.Id,
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   strconv.Itoa(user.Id),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	})

	signed, err := token.SignedString(j.privateKey)
	if err != nil {
		return models.AccessToken{}, fmt.Errorf("sign access token: %w", err)
	}
	return models.AccessToken{Token: signed, ExpiresAt: expiresAt}, nil
}

func (j *JWTTokenIssuer) NewRefresh() (string, string, error) {
	raw := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return token, j.HashRefresh(token), nil
}

// HashRefresh uses a plain SHA-256: refresh tokens are random, so they don't need a slow hash.
func (j *JWTTokenIssuer) HashRefresh(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

package httpauth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthFormat = errors.New("invalid authorization format")
	ErrInvalidToken      = errors.New("invalid or expired token")
)

// NewAuthMiddleware rejects requests without a valid RS* bearer token signed with publicKeyPEM
// and stores its UserClaims in the request context. Literal "\n" in the PEM are accepted.
func NewAuthMiddleware(publicKeyPEM string) (func(http.Handler) http.Handler, error) {
	pemStr := strings.ReplaceAll(publicKeyPEM, "\\n", "\n")
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pemStr))
	if err != nil {
		return nil, fmt.Errorf("parse jwt public key: %w", err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := parseClaims(r, publicKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userClaimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}

func parseClaims(r *http.Request, publicKey *rsa.PublicKey) (*UserClaims, error) {
	tokenStr, err := bearerToken(r)
	if err != nil {
		return nil, err
	}

	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		return publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func bearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingAuthHeader
	}

	scheme, token, ok := strings.Cut(authHeader, " ")
	if !ok || !strings.EqualFold(scheme, "bearer") || token == "" {
		return "", ErrInvalidAuthFormat
	}
	return token, nil
}

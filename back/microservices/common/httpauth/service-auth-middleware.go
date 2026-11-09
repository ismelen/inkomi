package httpauth

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// NewServiceAuthMiddleware only lets through requests with a bearer token for audience signed
// by one of the services in publicKeys (service name → RS* public key PEM) and stores its
// ServiceClaims in the request context. Literal "\n" in the PEMs are accepted.
func NewServiceAuthMiddleware(audience string, publicKeys map[string]string) (func(http.Handler) http.Handler, error) {
	keys := make(map[string]*rsa.PublicKey, len(publicKeys))
	for service, publicKeyPEM := range publicKeys {
		pemStr := strings.ReplaceAll(publicKeyPEM, "\n", "\n")
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(pemStr))
		if err != nil {
			return nil, fmt.Errorf("parse %s public key: %w", service, err)
		}
		keys[service] = key
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := parseServiceClaims(r, audience, keys)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), serviceClaimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}, nil
}

func parseServiceClaims(r *http.Request, audience string, keys map[string]*rsa.PublicKey) (*ServiceClaims, error) {
	tokenStr, err := bearerToken(r)
	if err != nil {
		return nil, err
	}

	claims := &ServiceClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}
		subject, err := token.Claims.GetSubject()
		if err != nil {
			return nil, ErrInvalidToken
		}
		key, ok := keys[subject]
		if !ok {
			return nil, ErrInvalidToken
		}
		return key, nil
	}, jwt.WithAudience(audience), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

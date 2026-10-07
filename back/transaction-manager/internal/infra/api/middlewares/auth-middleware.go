package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

const UserClaimsKey string = "userClaims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := getToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		claims := &UserClaims{}

		tokenData, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("invalid or expired token")
			}

			pubKey, err := getPublicKey()
			if err != nil {
				return nil, fmt.Errorf("failed to validate token: %w", err)
			}
			return pubKey, nil
		})

		if err != nil || !tokenData.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("Missing authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("Invalid autorization format")
	}

	return parts[1], nil
}

var (
	rsaPublicKey   any
	initPubKeyOnce sync.Once
)

func getPublicKey() (any, error) {
	var err error
	initPubKeyOnce.Do(func() {
		pubKeyPEM := os.Getenv("JWT_PUBLIC_KEY")
		pemStr := strings.ReplaceAll(pubKeyPEM, "\\n", "\n")
		rsaPublicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(pemStr))
	})
	return rsaPublicKey, err
}

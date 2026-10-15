package httpauth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type userClaimsKey struct{}

type UserClaims struct {
	Id   int    `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func GetUserClaims(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey{}).(*UserClaims)
	return claims, ok
}

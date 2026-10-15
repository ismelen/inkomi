package httpauth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type serviceClaimsKey struct{}

// ServiceClaims identify the service that signed a request in Subject.
type ServiceClaims struct {
	jwt.RegisteredClaims
}

func GetServiceClaims(ctx context.Context) (*ServiceClaims, bool) {
	claims, ok := ctx.Value(serviceClaimsKey{}).(*ServiceClaims)
	return claims, ok
}

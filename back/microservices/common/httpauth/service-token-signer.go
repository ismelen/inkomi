package httpauth

import (
	"crypto/rsa"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ServiceTokenSigner issues short-lived RS256 tokens that identify a service in
// service-to-service requests.
type ServiceTokenSigner struct {
	service    string
	privateKey *rsa.PrivateKey
	ttl        time.Duration
}

// NewServiceTokenSigner signs as service with privateKeyPEM. Literal "\n" in the PEM are accepted.
func NewServiceTokenSigner(service, privateKeyPEM string, ttl time.Duration) (*ServiceTokenSigner, error) {
	pemStr := strings.ReplaceAll(privateKeyPEM, "\n", "\n")
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pemStr))
	if err != nil {
		return nil, fmt.Errorf("parse service private key: %w", err)
	}
	return &ServiceTokenSigner{service: service, privateKey: privateKey, ttl: ttl}, nil
}

// Sign returns a token for a request to the audience service.
func (s *ServiceTokenSigner) Sign(audience string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, ServiceClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.service,
			Subject:   s.service,
			Audience:  jwt.ClaimStrings{audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	})
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("sign service token: %w", err)
	}
	return signed, nil
}

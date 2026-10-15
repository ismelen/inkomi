package google

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
)

var _ ports.IGoogleVerifier = (*IdTokenGoogleVerifier)(nil)

// IdTokenGoogleVerifier accepts Google ID tokens issued to any of clientIds (one per platform).
type IdTokenGoogleVerifier struct {
	validator *idtoken.Validator
	clientIds []string
}

func NewIdTokenGoogleVerifier(ctx context.Context, client *http.Client, clientIds []string) (*IdTokenGoogleVerifier, error) {
	if len(clientIds) == 0 {
		return nil, fmt.Errorf("no google client ids configured")
	}
	validator, err := idtoken.NewValidator(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("create google id token validator: %w", err)
	}
	return &IdTokenGoogleVerifier{validator: validator, clientIds: clientIds}, nil
}

func (g *IdTokenGoogleVerifier) Verify(ctx context.Context, idToken string) (*models.ExternalAccount, error) {
	// The audience is checked below against every client id, so it isn't passed here.
	payload, err := g.validator.Validate(ctx, idToken, "")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrInvalidExternalToken, err)
	}
	return accountFromPayload(payload, g.clientIds)
}

func accountFromPayload(payload *idtoken.Payload, clientIds []string) (*models.ExternalAccount, error) {
	if !slices.Contains(clientIds, payload.Audience) {
		return nil, fmt.Errorf("%w: unexpected audience %q", models.ErrInvalidExternalToken, payload.Audience)
	}
	if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
		return nil, fmt.Errorf("%w: unexpected issuer %q", models.ErrInvalidExternalToken, payload.Issuer)
	}
	if payload.Subject == "" {
		return nil, fmt.Errorf("%w: missing subject", models.ErrInvalidExternalToken)
	}

	email, _ := payload.Claims["email"].(string)
	verified, _ := payload.Claims["email_verified"].(bool)
	return &models.ExternalAccount{
		Provider:      models.ProviderGoogle,
		Subject:       payload.Subject,
		Email:         email,
		EmailVerified: verified,
	}, nil
}

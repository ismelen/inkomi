package refreshsession

import (
	"context"
	"errors"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type RefreshSessionUC struct {
	sessions   ports.ISessionRepository
	users      ports.IUserRepository
	issuer     ports.ITokenIssuer
	refreshTTL time.Duration
}

func NewRefreshSessionUC(
	sessions ports.ISessionRepository,
	users ports.IUserRepository,
	issuer ports.ITokenIssuer,
	refreshTTL time.Duration,
) *RefreshSessionUC {
	return &RefreshSessionUC{sessions: sessions, users: users, issuer: issuer, refreshTTL: refreshTTL}
}

// Execute exchanges refreshToken for a new token pair. The refresh token is rotated, so each
// one can be used only once.
func (r *RefreshSessionUC) Execute(ctx context.Context, refreshToken string) (*models.TokenPair, error) {
	if refreshToken == "" {
		return nil, models.ErrInvalidRefreshToken
	}

	newToken, newHash, err := r.issuer.NewRefresh()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(r.refreshTTL)
	userId, err := r.sessions.Rotate(ctx, r.issuer.HashRefresh(refreshToken), newHash, now, expiresAt)
	if err != nil {
		return nil, err
	}

	user, err := r.users.GetById(ctx, userId)
	if errors.Is(err, models.ErrUserNotFound) {
		return nil, models.ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, err
	}

	access, err := r.issuer.IssueAccess(user)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:      access.Token,
		AccessExpiresAt:  access.ExpiresAt,
		RefreshToken:     newToken,
		RefreshExpiresAt: expiresAt,
	}, nil
}

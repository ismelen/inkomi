package startsession

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type StartSessionUC struct {
	sessions   ports.ISessionRepository
	issuer     ports.ITokenIssuer
	refreshTTL time.Duration
}

func NewStartSessionUC(sessions ports.ISessionRepository, issuer ports.ITokenIssuer, refreshTTL time.Duration) *StartSessionUC {
	return &StartSessionUC{sessions: sessions, issuer: issuer, refreshTTL: refreshTTL}
}

// Execute opens a session of user on deviceId, replacing the previous one of that device.
func (s *StartSessionUC) Execute(ctx context.Context, user *models.User, deviceId string) (*models.TokenPair, error) {
	if err := models.ValidateDeviceId(deviceId); err != nil {
		return nil, err
	}

	refreshToken, refreshHash, err := s.issuer.NewRefresh()
	if err != nil {
		return nil, err
	}

	session := models.Session{
		UserId:      user.Id,
		DeviceId:    deviceId,
		RefreshHash: refreshHash,
		ExpiresAt:   time.Now().Add(s.refreshTTL),
	}
	if err := s.sessions.Save(ctx, session); err != nil {
		return nil, err
	}

	access, err := s.issuer.IssueAccess(user)
	if err != nil {
		return nil, err
	}

	return &models.TokenPair{
		AccessToken:      access.Token,
		AccessExpiresAt:  access.ExpiresAt,
		RefreshToken:     refreshToken,
		RefreshExpiresAt: session.ExpiresAt,
	}, nil
}

package logout

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type LogoutUC struct {
	sessions ports.ISessionRepository
	issuer   ports.ITokenIssuer
}

func NewLogoutUC(sessions ports.ISessionRepository, issuer ports.ITokenIssuer) *LogoutUC {
	return &LogoutUC{sessions: sessions, issuer: issuer}
}

// Execute closes the session of refreshToken; an unknown token is ignored.
func (l *LogoutUC) Execute(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return l.sessions.DeleteByRefreshHash(ctx, l.issuer.HashRefresh(refreshToken))
}

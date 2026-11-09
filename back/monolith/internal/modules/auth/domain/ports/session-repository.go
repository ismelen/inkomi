package ports

import (
	"context"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
)

type ISessionRepository interface {
	// Save replaces the session of the same user and device.
	Save(ctx context.Context, session models.Session) error
	// Rotate swaps the refresh hash of the session that still holds oldHash and hasn't expired
	// at now, and returns its user; otherwise it returns models.ErrInvalidRefreshToken.
	Rotate(ctx context.Context, oldHash, newHash string, now, expiresAt time.Time) (int, error)
	DeleteByRefreshHash(ctx context.Context, hash string) error
}

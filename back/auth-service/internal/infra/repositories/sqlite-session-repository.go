package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/datasources"
)

var _ ports.ISessionRepository = (*SQLiteSessionRepository)(nil)

type SQLiteSessionRepository struct {
	db *sql.DB
}

func NewSQLiteSessionRepository(db *sql.DB) *SQLiteSessionRepository {
	return &SQLiteSessionRepository{db: db}
}

func (r *SQLiteSessionRepository) Save(ctx context.Context, session models.Session) error {
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, `
		INSERT INTO sessions (userId, deviceId, refreshHash, expiresAt) VALUES (?, ?, ?, ?)
		ON CONFLICT (userId, deviceId) DO UPDATE SET
			refreshHash = excluded.refreshHash,
			expiresAt = excluded.expiresAt`,
		session.UserId, session.DeviceId, session.RefreshHash, session.ExpiresAt.Unix(),
	)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	return nil
}

func (r *SQLiteSessionRepository) Rotate(ctx context.Context, oldHash, newHash string, now, expiresAt time.Time) (int, error) {
	var userId int
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, `
		UPDATE sessions SET refreshHash = ?, expiresAt = ?
		WHERE refreshHash = ? AND expiresAt > ?
		RETURNING userId`,
		newHash, expiresAt.Unix(), oldHash, now.Unix(),
	).Scan(&userId)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, models.ErrInvalidRefreshToken
	}
	if err != nil {
		return 0, fmt.Errorf("rotate session: %w", err)
	}
	return userId, nil
}

func (r *SQLiteSessionRepository) DeleteByRefreshHash(ctx context.Context, hash string) error {
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, `DELETE FROM sessions WHERE refreshHash = ?`, hash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

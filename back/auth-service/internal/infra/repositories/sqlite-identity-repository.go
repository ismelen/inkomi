package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/datasources"
)

var _ ports.IIdentityRepository = (*SQLiteIdentityRepository)(nil)

type SQLiteIdentityRepository struct {
	db *sql.DB
}

func NewSQLiteIdentityRepository(db *sql.DB) *SQLiteIdentityRepository {
	return &SQLiteIdentityRepository{db: db}
}

func (r *SQLiteIdentityRepository) Get(ctx context.Context, provider models.Provider, subject string) (*models.Identity, error) {
	identity := models.Identity{Provider: provider, Subject: subject}
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx,
		`SELECT userId FROM identities WHERE provider = ? AND subject = ?`, provider, subject,
	).Scan(&identity.UserId)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrIdentityNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select identity: %w", err)
	}
	return &identity, nil
}

func (r *SQLiteIdentityRepository) Create(ctx context.Context, identity models.Identity) error {
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`INSERT INTO identities (provider, subject, userId) VALUES (?, ?, ?)`,
		identity.Provider, identity.Subject, identity.UserId,
	)
	if err != nil {
		return fmt.Errorf("insert identity: %w", err)
	}
	return nil
}

func (r *SQLiteIdentityRepository) ListProviders(ctx context.Context, userId int) ([]models.Provider, error) {
	rows, err := datasources.GetDB(ctx, r.db).QueryContext(ctx,
		`SELECT DISTINCT provider FROM identities WHERE userId = ? ORDER BY provider`, userId,
	)
	if err != nil {
		return nil, fmt.Errorf("select providers: %w", err)
	}
	defer rows.Close()

	providers := []models.Provider{}
	for rows.Next() {
		var provider models.Provider
		if err := rows.Scan(&provider); err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}
		providers = append(providers, provider)
	}
	return providers, rows.Err()
}

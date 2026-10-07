package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/datasources"
)

type SQLiteConfigRepository struct {
	db *sql.DB
}

func NewSQLiteConfigRepository(db *sql.DB) ports.ConfigRepository {
	return &SQLiteConfigRepository{db: db}
}

func (r *SQLiteConfigRepository) Create(ctx context.Context, config *models.Config) (string, error) {
	query := `
		INSERT INTO configs (hash, data)
		VALUES (?, ?)
		ON CONFLICT(hash) DO UPDATE SET
			lastUsed = CURRENT_TIMESTAMP
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		config.Hash,
		config.Data,
	)
	return config.Hash, err
}

func (r *SQLiteConfigRepository) GetByHash(ctx context.Context, hash string) (*models.Config, error) {
	query := `
		SELECT hash, data, createdAt, lastUsed
		FROM configs
		WHERE hash = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, hash)

	var config models.Config
	err := row.Scan(
		&config.Hash,
		&config.Data,
		&config.CreatedAt,
		&config.LastUsed,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &config, nil
}

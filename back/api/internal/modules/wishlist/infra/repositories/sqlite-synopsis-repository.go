package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

type SQLiteSynopsisRepository struct {
	db *sql.DB
}

var _ ports.ISynopsisRepository = (*SQLiteSynopsisRepository)(nil)

func NewSQLiteSynopsisRepository(db *sql.DB) *SQLiteSynopsisRepository {
	return &SQLiteSynopsisRepository{db: db}
}

func (r *SQLiteSynopsisRepository) Get(ctx context.Context, workKey, lang string) (*models.Synopsis, error) {
	synopsis := models.Synopsis{WorkKey: workKey, Language: lang}
	err := r.db.QueryRowContext(ctx, `
		SELECT description, fetchedAt FROM wishlist_synopsis WHERE workKey = ? AND lang = ?
	`, workKey, lang).Scan(&synopsis.Text, &synopsis.FetchedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrSynopsisNotFound
	}
	if err != nil {
		return nil, err
	}
	return &synopsis, nil
}

func (r *SQLiteSynopsisRepository) Save(ctx context.Context, synopsis models.Synopsis) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wishlist_synopsis (workKey, lang, description, fetchedAt)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (workKey, lang) DO UPDATE SET
			description = excluded.description, fetchedAt = excluded.fetchedAt
	`, synopsis.WorkKey, synopsis.Language, synopsis.Text)
	return err
}

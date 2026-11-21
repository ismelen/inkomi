package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

const subjectSeparator = "\n"

type SQLiteWorkRepository struct {
	db *sql.DB
}

var _ ports.IWorkRepository = (*SQLiteWorkRepository)(nil)

func NewSQLiteWorkRepository(db *sql.DB) *SQLiteWorkRepository {
	return &SQLiteWorkRepository{db: db}
}

func (r *SQLiteWorkRepository) Get(ctx context.Context, workKey string) (*models.WorkDetails, error) {
	work := models.WorkDetails{WorkKey: workKey}
	var subjects string
	err := r.db.QueryRowContext(ctx, `
		SELECT title, description, coverUrl, subjects, firstPublished
		FROM wishlist_work WHERE workKey = ?
	`, workKey).Scan(&work.Title, &work.Description, &work.CoverURL, &subjects, &work.FirstPublished)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, models.ErrWorkNotFound
	}
	if err != nil {
		return nil, err
	}
	if subjects != "" {
		work.Subjects = strings.Split(subjects, subjectSeparator)
	}
	return &work, nil
}

func (r *SQLiteWorkRepository) Save(ctx context.Context, work models.WorkDetails) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO wishlist_work (workKey, title, description, coverUrl, subjects, firstPublished, fetchedAt)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT (workKey) DO UPDATE SET
			title = excluded.title, description = excluded.description, coverUrl = excluded.coverUrl,
			subjects = excluded.subjects, firstPublished = excluded.firstPublished, fetchedAt = excluded.fetchedAt
	`, work.WorkKey, work.Title, work.Description, work.CoverURL,
		strings.Join(work.Subjects, subjectSeparator), work.FirstPublished)
	return err
}

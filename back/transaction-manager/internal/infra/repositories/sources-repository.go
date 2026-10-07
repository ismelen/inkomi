package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/datasources"
)

type SQLiteSourceRepository struct {
	db *sql.DB
}

func NewSQLiteSourceRepository(db *sql.DB) ports.SourceRepository {
	return &SQLiteSourceRepository{db: db}
}

func (r *SQLiteSourceRepository) Create(ctx context.Context, source *models.Source) (string, error) {
	source.Id = uuid.New().String()

	query := `
		INSERT INTO sources (
			id, userId, size, filename, title,  shouldJoin, 
			readingDirection, folderId, configHash
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, 
			?, ?, ?, ?, 
			?, ?, ?
		)
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		source.Id,
		source.UserId,
		source.Size,
		source.Filename,
		source.Title,
		source.ShouldJoin,
		source.ReadingDirection,
		source.FolderId,
		source.ConfigHash,
	)
	if err != nil {
		return "", err
	}

	return source.Id, nil
}

func (r *SQLiteSourceRepository) CreateAll(ctx context.Context, sources []*models.Source) ([]string, error) {
	ids := make([]string, 0, len(sources))
	for _, source := range sources {
		id, err := r.Create(ctx, source)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *SQLiteSourceRepository) GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error) {
	query := `
		SELECT 
			id, userId, size, filename, title, status, error, 
			createdAt, updatedAt, completedAt, shouldJoin, 
			readingDirection, folderId, configHash
		FROM sources
		WHERE id = ? AND userId = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userID)

	var source models.Source
	err := row.Scan(
		&source.Id,
		&source.UserId,
		&source.Size,
		&source.Filename,
		&source.Title,
		&source.Status,
		&source.Error,
		&source.CreatedAt,
		&source.UpdatedAt,
		&source.CompletedAt,
		&source.ShouldJoin,
		&source.ReadingDirection,
		&source.FolderId,
		&source.ConfigHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or domain specific error
		}
		return nil, err
	}

	return &source, nil
}

func (r *SQLiteSourceRepository) Update(ctx context.Context, source *models.Source) error {
	query := `
		UPDATE sources SET
			size = ?,
			filename = ?,
			title = ?,
			status = ?,
			error = ?,
			updatedAt = CURRENT_TIMESTAMP,
			completedAt = ?,
			shouldJoin = ?,
			readingDirection = ?,
			folderId = ?,
			configHash = ?
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		source.Size,
		source.Filename,
		source.Title,
		source.Status,
		source.Error,
		source.CompletedAt,
		source.ShouldJoin,
		source.ReadingDirection,
		source.FolderId,
		source.ConfigHash,
		source.Id,
		source.UserId,
	)
	return err
}

func (r *SQLiteSourceRepository) Delete(ctx context.Context, id string, userID int) error {
	query := `DELETE FROM sources WHERE id = ? AND userId = ?`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, id, userID)
	return err
}

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
			id, userId, size, filename, title, kepubify, type, shouldJoin, 
			readingDirection, folderId, configHash
		) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, 
			?, ?, ?
		)
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		source.Id,
		source.UserId,
		source.Size,
		source.Filename,
		source.Title,
		source.Kepubify,
		source.Type,
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

func (r *SQLiteSourceRepository) GetByIdAndUserIdCompact(ctx context.Context, id string, userID int) (*models.CompactSoruce, error) {
	query := `
		SELECT 
			s.id, s.userId, s.filename, s.title, s.shouldJoin, 
			s.readingDirection, s.folderId, s.kepubify, c.data
		FROM sources as s
		INNER JOIN configs as c ON c.hash = s.configHash
		WHERE id = ? AND userId = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userID)

	var source models.CompactSoruce
	err := row.Scan(
		&source.Id,
		&source.UserId,
		&source.Filename,
		&source.Title,
		&source.ShouldJoin,
		&source.ReadingDirection,
		&source.FolderId,
		&source.Kepubify,
		&source.Config.Data,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or domain specific error
		}
		return nil, err
	}

	return &source, nil
}

func (r *SQLiteSourceRepository) GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error) {
	query := `
		SELECT 
			s.id, s.userId, s.size, s.filename, s.title, s.status, s.error, 
			s.createdAt, s.updatedAt, s.completedAt, s.shouldJoin, 
			s.readingDirection, s.folderId, s.configHash
		FROM sources as s
		INNER JOIN configs as c ON c.hash = s.configHash
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
		&source.MangaConfig.Hash,
		&source.MangaConfig.Data,
		&source.MangaConfig.CreatedAt,
		&source.MangaConfig.LastUsed,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Or domain specific error
		}
		return nil, err
	}

	return &source, nil
}

func (r *SQLiteSourceRepository) UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus, errMsg *string) error {
	query := `
		UPDATE sources SET
			status = ?,
			error = ?,
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		status,
		errMsg,
		id,
		userId,
	)
	return err
}

func (r *SQLiteSourceRepository) Delete(ctx context.Context, id string, userID int) error {
	query := `DELETE FROM sources WHERE id = ? AND userId = ?`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, id, userID)
	return err
}

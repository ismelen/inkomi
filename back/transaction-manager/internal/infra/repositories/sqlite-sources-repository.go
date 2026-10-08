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

func (r *SQLiteSourceRepository) SetError(ctx context.Context, id string, userId int, errMsg string) error {
	query := `
		UPDATE sources SET
			error = ?,
			status = ?,
			upldatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := r.db.Query(query, errMsg, models.SourceStatusFailed, id, userId)
	return err
}

func (r *SQLiteSourceRepository) CheckFolderIsWaitingJoin(ctx context.Context, folderId string, userId int) (*models.FolderWithItems, bool, error) {
	// itemsQuery := `
	// 	SELECT id, filename, size, status
	// 	FROM sources
	// 	WHERE folderId = ? AND userId = ?
	// `
	// rows, err := datasources.GetDB(ctx, r.db).QueryContext(ctx, itemsQuery, folderId, userId)
	// if err != nil {
	// 	return nil, false, err
	// }
	// defer rows.Close()

	// var items []models.SourceItem
	// hasItems := false

	// for rows.Next() {
	// 	hasItems = true
	// 	var item models.SourceItem
	// 	var status string
	// 	if err := rows.Scan(&item.Id, &item.Filename, &item.Size, &status); err != nil {
	// 		return nil, false, err
	// 	}
	// 	if status != string(models.SourceStatusWaitingJoin) {
	// 		return nil, false, nil
	// 	}
	// 	items = append(items, item)
	// }

	// if err := rows.Err(); err != nil {
	// 	return nil, false, err
	// }

	// if !hasItems {
	// 	return nil, false, nil
	// }

	// folderQuery := `
	// 	SELECT
	// 		s.id, s.userId, s.filename, s.title, s.shouldJoin,
	// 		s.readingDirection, s.folderId, s.kepubify, c.data
	// 	FROM sources as s
	// 	LEFT JOIN configs as c ON c.hash = s.configHash
	// 	WHERE s.id = ? AND s.userId = ?
	// `
	// row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, folderQuery, folderId, userId)

	// var folder models.CompactSoruce
	// var configData sql.NullString
	// err = row.Scan(
	// 	&folder.Id,
	// 	&folder.UserId,
	// 	&folder.Filename,
	// 	&folder.Title,
	// 	&folder.ShouldJoin,
	// 	&folder.ReadingDirection,
	// 	&folder.FolderId,
	// 	&folder.Kepubify,
	// 	&configData,
	// )
	// if err != nil {
	// 	if errors.Is(err, sql.ErrNoRows) {
	// 		return nil, false, nil
	// 	}
	// 	return nil, false, err
	// }

	// if configData.Valid {
	// 	folder.Config = &models.Config{Data: configData.String}
	// }

	// return &models.FolderWithItems{
	// 	Folder: folder,
	// 	Items:  items,
	// }, true, nil
}

func (r *SQLiteSourceRepository) CheckSiblingsInWaitingJoin(ctx context.Context, id string, userId string) (*models.CompactSoruce, bool, error) {
	query := `
		SELECT 
			p.id, p.userId, p.filename, p.title, p.shouldJoin, 
			p.readingDirection, p.folderId, p.kepubify, c.data
		FROM sources as s
		INNER JOIN sources as p ON s.folderId = p.id AND s.userId = p.userId
		LEFT JOIN configs as c ON c.hash = p.configHash
		WHERE s.id = ? AND s.userId = ?
		  AND NOT EXISTS (
			  SELECT 1 
			  FROM sources as sib 
			  WHERE sib.folderId = s.folderId 
			    AND sib.userId = s.userId 
			    AND sib.id != s.id 
			    AND sib.status != 'waiting_join'
		  )
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userId)

	var source models.CompactSoruce
	var configData sql.NullString
	err := row.Scan(
		&source.Id,
		&source.UserId,
		&source.Filename,
		&source.Title,
		&source.ShouldJoin,
		&source.ReadingDirection,
		&source.FolderId,
		&source.Kepubify,
		&configData,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if configData.Valid {
		source.Config = &models.Config{Data: configData.String}
	}

	return &source, true, nil
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

func (r *SQLiteSourceRepository) UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error {
	query := `
		UPDATE sources SET
			status = ?,
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		status,
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

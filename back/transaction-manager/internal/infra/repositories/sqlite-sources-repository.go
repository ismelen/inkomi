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

const sourceColumns = `
	s.id, s.userId, s.title, s.filename, s.kepubify, s.toCloud, s.configKey,
	s.isRtl, s."join", s.status, s.size, s.isFile,
	s.createdAt, s.updatedAt, s.completedAt, s.parentId`

type SQLiteSourceRepository struct {
	db *sql.DB
}

func NewSQLiteSourceRepository(db *sql.DB) ports.SourceRepository {
	return &SQLiteSourceRepository{db: db}
}

func (r *SQLiteSourceRepository) Create(ctx context.Context, source *models.Source) (string, error) {
	source.ID = uuid.New().String()

	query := `
		INSERT INTO source (
			id, userId, title, filename, kepubify, toCloud, configKey,
			isRtl, "join", size, isFile, parentId
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
		)
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		source.ID,
		source.UserID,
		source.Title,
		source.Filename,
		source.Kepubify,
		source.ToCloud,
		source.ConfigKey,
		source.IsRTL,
		source.Join,
		source.Size,
		source.IsFile,
		source.ParentID,
	)
	if err != nil {
		return "", err
	}

	return source.ID, nil
}

func (r *SQLiteSourceRepository) GetByIDAndUserID(ctx context.Context, id string, userID int) (*models.Source, error) {
	query := `SELECT ` + sourceColumns + `
		FROM source AS s
		WHERE s.id = ? AND s.userId = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userID)

	var source models.Source
	if err := row.Scan(sourceFields(&source)...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSourceNotFound
		}
		return nil, err
	}

	return &source, nil
}

func (r *SQLiteSourceRepository) GetDetailByIDAndUserID(ctx context.Context, id string, userID int) (*models.SourceDetail, error) {
	query := `SELECT ` + sourceColumns + `,
			c.hash, c.data, c.createdAt, c.lastUsed,
			p.id, p.title, p.isRtl, p."join"
		FROM source AS s
		LEFT JOIN config AS c ON c.hash = s.configKey
		LEFT JOIN source AS p ON p.id = s.parentId AND p.userId = s.userId
		WHERE s.id = ? AND s.userId = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userID)

	var detail models.SourceDetail
	var cfgHash, cfgData, parentID, parentTitle sql.NullString
	var cfgCreatedAt, cfgLastUsed sql.NullTime
	var parentIsRTL, parentJoin sql.NullBool

	fields := append(sourceFields(&detail.Source),
		&cfgHash, &cfgData, &cfgCreatedAt, &cfgLastUsed,
		&parentID, &parentTitle, &parentIsRTL, &parentJoin,
	)
	if err := row.Scan(fields...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSourceNotFound
		}
		return nil, err
	}

	if cfgHash.Valid {
		detail.Config = &models.Config{
			Hash:      cfgHash.String,
			Data:      cfgData.String,
			CreatedAt: cfgCreatedAt.Time,
			LastUsed:  cfgLastUsed.Time,
		}
	}

	if parentID.Valid {
		detail.Parent = &models.SourceParent{
			ID:    parentID.String,
			Title: parentTitle.String,
			IsRTL: parentIsRTL.Bool,
			Join:  parentJoin.Bool,
		}
	}

	return &detail, nil
}

func (r *SQLiteSourceRepository) UpdateStatus(ctx context.Context, id string, userID int, status models.SourceStatus) error {
	query := `
		UPDATE source SET
			status = ?,
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, status, id, userID)
	return err
}

func (r *SQLiteSourceRepository) Delete(ctx context.Context, id string, userID int) error {
	query := `DELETE FROM source WHERE id = ? AND userId = ?`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, id, userID)
	return err
}

// sourceFields returns the scan destinations matching sourceColumns.
func sourceFields(s *models.Source) []any {
	return []any{
		&s.ID, &s.UserID, &s.Title, &s.Filename, &s.Kepubify, &s.ToCloud, &s.ConfigKey,
		&s.IsRTL, &s.Join, &s.Status, &s.Size, &s.IsFile,
		&s.CreatedAt, &s.UpdatedAt, &s.CompletedAt, &s.ParentID,
	}
}

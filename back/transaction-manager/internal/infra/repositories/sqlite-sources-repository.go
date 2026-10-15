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
	s.isRtl, s."join", s.status, s.errorMsg, s.size, s.isFile,
	s.createdAt, s.updatedAt, s.completedAt, s.parentId, s.position, s.isResult`

type SQLiteSourceRepository struct {
	db *sql.DB
}

func NewSQLiteSourceRepository(db *sql.DB) ports.SourceRepository {
	return &SQLiteSourceRepository{db: db}
}

func (r *SQLiteSourceRepository) Create(ctx context.Context, source *models.Source) (string, error) {
	source.Id = uuid.New().String()

	query := `
		INSERT INTO source (
			id, userId, title, filename, kepubify, toCloud, configKey,
			isRtl, "join", size, isFile, parentId, position
		) VALUES (
			?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?
		)
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		source.Id,
		source.UserId,
		source.Title,
		source.Filename,
		source.Kepubify,
		source.ToCloud,
		source.ConfigKey,
		source.IsRTL,
		source.Join,
		source.Size,
		source.IsFile,
		source.ParentId,
		source.Position,
	)
	if err != nil {
		return "", err
	}

	return source.Id, nil
}

func (r *SQLiteSourceRepository) GetByIdAndUserId(ctx context.Context, id string, userId int) (*models.Source, error) {
	query := `SELECT ` + sourceColumns + `
		FROM source AS s
		WHERE s.id = ? AND s.userId = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userId)

	var source models.Source
	if err := row.Scan(sourceFields(&source)...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSourceNotFound
		}
		return nil, err
	}

	return &source, nil
}

func (r *SQLiteSourceRepository) GetDetailByIdAndUserId(ctx context.Context, id string, userId int) (*models.SourceDetail, error) {
	query := `SELECT ` + sourceColumns + `,
			c.hash, c.data, c.createdAt, c.lastUsed,
			p.id, p.title, p.isRtl, p."join",
			(SELECT COUNT(*) FROM source AS ch WHERE ch.parentId = p.id AND ch.userId = p.userId AND ch.isResult = FALSE)
		FROM source AS s
		LEFT JOIN config AS c ON c.hash = s.configKey
		LEFT JOIN source AS p ON p.id = s.parentId AND p.userId = s.userId
		WHERE s.id = ? AND s.userId = ?
	`
	row := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, query, id, userId)

	var detail models.SourceDetail
	var cfgHash, cfgData, parentId, parentTitle sql.NullString
	var cfgCreatedAt, cfgLastUsed sql.NullTime
	var parentIsRTL, parentJoin sql.NullBool
	var parentChildCount sql.NullInt64

	fields := append(sourceFields(&detail.Source),
		&cfgHash, &cfgData, &cfgCreatedAt, &cfgLastUsed,
		&parentId, &parentTitle, &parentIsRTL, &parentJoin, &parentChildCount,
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

	if parentId.Valid {
		detail.Parent = &models.SourceParent{
			Id:         parentId.String,
			Title:      parentTitle.String,
			IsRTL:      parentIsRTL.Bool,
			Join:       parentJoin.Bool,
			ChildCount: int(parentChildCount.Int64),
		}
	}

	return &detail, nil
}

func (r *SQLiteSourceRepository) UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error {
	query := `
		UPDATE source SET
			status = ?,
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, status, id, userId)
	return err
}

func (r *SQLiteSourceRepository) UpdateErrorMsg(ctx context.Context, id string, userId int, errorMsg string) error {
	query := `
		UPDATE source SET
			status = ?,
			errorMsg = ?,
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, models.SourceStatusFailed, errorMsg, id, userId)
	return err
}

func (r *SQLiteSourceRepository) UpsertResult(ctx context.Context, result *models.Source) error {
	query := `
		INSERT INTO source (
			id, userId, title, filename, kepubify, toCloud,
			status, errorMsg, size, isFile, parentId, position, isResult, completedAt
		) VALUES (
			?, ?, ?, ?, ?, ?,
			?, ?, ?, TRUE, ?, ?, TRUE,
			CASE WHEN ? IN ('done', 'sent', 'failed') THEN CURRENT_TIMESTAMP END
		)
		ON CONFLICT (id, userId) DO UPDATE SET
			filename = excluded.filename,
			status = excluded.status,
			errorMsg = excluded.errorMsg,
			size = COALESCE(excluded.size, source.size),
			updatedAt = CURRENT_TIMESTAMP,
			completedAt = COALESCE(excluded.completedAt, source.completedAt)
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query,
		result.Id,
		result.UserId,
		result.Title,
		result.Filename,
		result.Kepubify,
		result.ToCloud,
		result.Status,
		result.ErrorMsg,
		result.Size,
		result.ParentId,
		result.Position,
		result.Status,
	)
	return err
}

func (r *SQLiteSourceRepository) ListResultsByParentIdAndUserId(ctx context.Context, parentId string, userId int) ([]models.Source, error) {
	query := `SELECT ` + sourceColumns + `
		FROM source AS s
		WHERE s.parentId = ? AND s.userId = ? AND s.isResult = TRUE
		ORDER BY s.position
	`
	rows, err := datasources.GetDB(ctx, r.db).QueryContext(ctx, query, parentId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.Source
	for rows.Next() {
		var s models.Source
		if err := rows.Scan(sourceFields(&s)...); err != nil {
			return nil, err
		}
		results = append(results, s)
	}

	return results, rows.Err()
}

func (r *SQLiteSourceRepository) Delete(ctx context.Context, id string, userId int) error {
	query := `DELETE FROM source WHERE id = ? AND userId = ?`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, id, userId)
	return err
}

// sourceFields returns the scan destinations matching sourceColumns.
func sourceFields(s *models.Source) []any {
	return []any{
		&s.Id, &s.UserId, &s.Title, &s.Filename, &s.Kepubify, &s.ToCloud, &s.ConfigKey,
		&s.IsRTL, &s.Join, &s.Status, &s.ErrorMsg, &s.Size, &s.IsFile,
		&s.CreatedAt, &s.UpdatedAt, &s.CompletedAt, &s.ParentId, &s.Position, &s.IsResult,
	}
}

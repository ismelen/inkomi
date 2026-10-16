package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/ports"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/infra/datasources"
)

const sourceColumns = `
	s.id, s.userId, s.title, s.filename, s.kepubify, s.toCloud, s.configKey,
	s.isRtl, s."join", s.status, s.errorMsg, s.size, s.isFile,
	s.createdAt, s.updatedAt, s.completedAt, s.parentId, s.position, s.isResult`

// sqliteTimeLayout is the format of CURRENT_TIMESTAMP; time arguments compared with stored columns must use it.
const sqliteTimeLayout = "2006-01-02 15:04:05"

// completedAtOnTerminal expects the new status as its argument.
const completedAtOnTerminal = `CASE WHEN ? IN ('done', 'sent', 'failed') THEN CURRENT_TIMESTAMP ELSE completedAt END`

type SQLiteSourceRepository struct {
	db *sql.DB
}

func NewSQLiteSourceRepository(db *sql.DB) ports.ISourceRepository {
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

func (r *SQLiteSourceRepository) ListTreesByUserId(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error) {
	query := `SELECT ` + sourceColumns + `
		FROM source AS s
		WHERE s.userId = ? AND s.parentId IS NULL
			AND (s.status != 'pending_upload' OR EXISTS (
				SELECT 1 FROM source AS c
				WHERE c.parentId = s.id AND c.userId = s.userId AND c.status != 'pending_upload'
			))`
	args := []any{userId}

	if filter.Cursor != nil {
		createdAt := filter.Cursor.CreatedAt.UTC().Format(sqliteTimeLayout)
		query += ` AND (s.createdAt < ? OR (s.createdAt = ? AND s.id < ?))`
		args = append(args, createdAt, createdAt, filter.Cursor.Id)
	}

	if filter.UpdatedSince != nil {
		updatedSince := filter.UpdatedSince.UTC().Format(sqliteTimeLayout)
		query += ` AND (s.updatedAt >= ? OR EXISTS (
				SELECT 1 FROM source AS c
				WHERE c.parentId = s.id AND c.userId = s.userId AND c.updatedAt >= ?
			))`
		args = append(args, updatedSince, updatedSince)
	}

	query += ` ORDER BY s.createdAt DESC, s.id DESC LIMIT ?`
	args = append(args, filter.Limit)

	roots, err := r.listSources(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return []models.SourceTree{}, nil
	}

	children, err := r.listChildren(ctx, userId, roots)
	if err != nil {
		return nil, err
	}

	trees := make([]models.SourceTree, len(roots))
	for i := range roots {
		trees[i] = models.SourceTree{Source: roots[i], Children: children[roots[i].Id]}
	}

	return trees, nil
}

// listChildren returns the items and join results of roots grouped by parent id.
func (r *SQLiteSourceRepository) listChildren(ctx context.Context, userId int, roots []models.Source) (map[string][]models.Source, error) {
	placeholders := make([]string, len(roots))
	args := make([]any, 0, len(roots)+1)
	args = append(args, userId)
	for i := range roots {
		placeholders[i] = "?"
		args = append(args, roots[i].Id)
	}

	query := `SELECT ` + sourceColumns + `
		FROM source AS s
		WHERE s.userId = ? AND s.parentId IN (` + strings.Join(placeholders, ", ") + `)
		ORDER BY s.isResult, s.position, s.createdAt
	`
	sources, err := r.listSources(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	children := make(map[string][]models.Source, len(roots))
	for _, s := range sources {
		children[*s.ParentId] = append(children[*s.ParentId], s)
	}

	return children, nil
}

func (r *SQLiteSourceRepository) Now(ctx context.Context) (time.Time, error) {
	var now string
	if err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, `SELECT CURRENT_TIMESTAMP`).Scan(&now); err != nil {
		return time.Time{}, err
	}

	t, err := time.Parse(sqliteTimeLayout, now)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse database time: %w", err)
	}

	return t.UTC(), nil
}

func (r *SQLiteSourceRepository) UpdateStatus(ctx context.Context, id string, userId int, status models.SourceStatus) error {
	query := `
		UPDATE source SET
			status = ?,
			updatedAt = CURRENT_TIMESTAMP,
			completedAt = ` + completedAtOnTerminal + `
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, status, status, id, userId)
	return err
}

func (r *SQLiteSourceRepository) UpdateFile(ctx context.Context, id string, userId int, filename string, size int64) error {
	query := `
		UPDATE source SET
			filename = COALESCE(NULLIF(?, ''), filename),
			size = COALESCE(NULLIF(?, 0), size),
			updatedAt = CURRENT_TIMESTAMP
		WHERE id = ? AND userId = ?
	`
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx, query, filename, size, id, userId)
	return err
}

func (r *SQLiteSourceRepository) UpdateErrorMsg(ctx context.Context, id string, userId int, errorMsg string) error {
	query := `
		UPDATE source SET
			status = ?,
			errorMsg = ?,
			updatedAt = CURRENT_TIMESTAMP,
			completedAt = CURRENT_TIMESTAMP
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
			filename = COALESCE(NULLIF(excluded.filename, ''), source.filename),
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
	return r.listSources(ctx, query, parentId, userId)
}

func (r *SQLiteSourceRepository) listSources(ctx context.Context, query string, args ...any) ([]models.Source, error) {
	rows, err := datasources.GetDB(ctx, r.db).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var s models.Source
		if err := rows.Scan(sourceFields(&s)...); err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}

	return sources, rows.Err()
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

package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/datasources"
)

var _ ports.IStorageReservationRepository = (*SQLiteStorageReservationRepository)(nil)

type SQLiteStorageReservationRepository struct {
	db *sql.DB
}

func NewSQLiteStorageReservationRepository(db *sql.DB) *SQLiteStorageReservationRepository {
	return &SQLiteStorageReservationRepository{db: db}
}

func (r *SQLiteStorageReservationRepository) Create(ctx context.Context, res *models.StorageReservation) error {
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`INSERT INTO storage_reservation (rootId, userId, bytes) VALUES (?, ?, ?)`,
		res.RootId, res.UserId, res.Bytes,
	)
	if err != nil {
		return fmt.Errorf("create storage reservation %s: %w", res.RootId, err)
	}
	return nil
}

func (r *SQLiteStorageReservationRepository) ListActive(ctx context.Context) ([]models.StorageReservation, error) {
	rows, err := datasources.GetDB(ctx, r.db).QueryContext(ctx,
		`SELECT rootId, userId, bytes, createdAt FROM storage_reservation WHERE releasedAt IS NULL`,
	)
	if err != nil {
		return nil, fmt.Errorf("list storage reservations: %w", err)
	}
	defer rows.Close()

	var out []models.StorageReservation
	for rows.Next() {
		var res models.StorageReservation
		if err := rows.Scan(&res.RootId, &res.UserId, &res.Bytes, &res.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan storage reservation: %w", err)
		}
		out = append(out, res)
	}
	return out, rows.Err()
}

func (r *SQLiteStorageReservationRepository) Release(ctx context.Context, rootId string) error {
	_, err := datasources.GetDB(ctx, r.db).ExecContext(ctx,
		`UPDATE storage_reservation SET releasedAt = CURRENT_TIMESTAMP WHERE rootId = ? AND releasedAt IS NULL`,
		rootId,
	)
	if err != nil {
		return fmt.Errorf("release storage reservation %s: %w", rootId, err)
	}
	return nil
}

func (r *SQLiteStorageReservationRepository) RootOf(ctx context.Context, sourceId string, userId int) (string, error) {
	var root string
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx,
		`SELECT COALESCE(parentId, id) FROM source WHERE id = ? AND userId = ?`, sourceId, userId,
	).Scan(&root)
	if err == sql.ErrNoRows {
		return "", models.ErrSourceNotFound
	}
	if err != nil {
		return "", fmt.Errorf("root of %s: %w", sourceId, err)
	}
	return root, nil
}

func (r *SQLiteStorageReservationRepository) TreeIds(ctx context.Context, rootId string) ([]string, error) {
	rows, err := datasources.GetDB(ctx, r.db).QueryContext(ctx,
		`SELECT id FROM source WHERE id = ? OR parentId = ?`, rootId, rootId,
	)
	if err != nil {
		return nil, fmt.Errorf("tree of %s: %w", rootId, err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan tree id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *SQLiteStorageReservationRepository) IsTreeFinished(ctx context.Context, rootId string) (bool, error) {
	var pending, results int
	var join bool
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx, `
		SELECT
			COALESCE((SELECT "join" FROM source WHERE id = ?), FALSE),
			(SELECT COUNT(*) FROM source WHERE (id = ? OR parentId = ?) AND isFile = TRUE
				AND status NOT IN ('done', 'sent', 'failed', 'joined', 'cancelled')),
			(SELECT COUNT(*) FROM source WHERE parentId = ? AND isResult = TRUE)`,
		rootId, rootId, rootId, rootId,
	).Scan(&join, &pending, &results)
	if err != nil {
		return false, fmt.Errorf("is tree %s finished: %w", rootId, err)
	}
	return pending == 0 && (!join || results > 0), nil
}

func (r *SQLiteStorageReservationRepository) IsTreeUnconfirmed(ctx context.Context, rootId string) (bool, error) {
	var confirmed int
	err := datasources.GetDB(ctx, r.db).QueryRowContext(ctx,
		`SELECT COUNT(*) FROM source WHERE (id = ? OR parentId = ?) AND isFile = TRUE AND status != 'pending_upload'`,
		rootId, rootId,
	).Scan(&confirmed)
	if err != nil {
		return false, fmt.Errorf("is tree %s unconfirmed: %w", rootId, err)
	}
	return confirmed == 0, nil
}

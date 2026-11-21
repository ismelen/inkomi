package repositories

import (
	"context"
	"database/sql"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

const itemColumns = `i.id, i.userId, i.kind, i.title, i.author, i.workKey, i.authorKey, COALESCE(w.coverUrl, ''), i.createdAt`

const itemFrom = ` FROM wishlist_item i LEFT JOIN wishlist_work w ON i.workKey != '' AND w.workKey = i.workKey`

type SQLiteItemRepository struct {
	db *sql.DB
}

var _ ports.IItemRepository = (*SQLiteItemRepository)(nil)

func NewSQLiteItemRepository(db *sql.DB) *SQLiteItemRepository {
	return &SQLiteItemRepository{db: db}
}

func (r *SQLiteItemRepository) Add(ctx context.Context, item *models.Item) error {
	query := `
		INSERT INTO wishlist_item (userId, kind, title, author, workKey, authorKey, dedupKey)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, createdAt
	`
	err := r.db.QueryRowContext(ctx, query,
		item.UserId, item.Kind, item.Title, item.Author, item.WorkKey, item.AuthorKey, item.DedupKey(),
	).Scan(&item.Id, &item.CreatedAt)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return models.ErrDuplicate
	}
	return err
}

func (r *SQLiteItemRepository) Delete(ctx context.Context, id int64, userId int) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM wishlist_item WHERE id = ? AND userId = ?`, id, userId)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return models.ErrItemNotFound
	}
	return nil
}

func (r *SQLiteItemRepository) ListByUserId(ctx context.Context, userId int) ([]models.Item, error) {
	return r.list(ctx, `SELECT `+itemColumns+itemFrom+` WHERE i.userId = ? ORDER BY i.createdAt DESC, i.id DESC`, userId)
}

func (r *SQLiteItemRepository) SetWorkKey(ctx context.Context, item *models.Item, workKey string) error {
	updated := *item
	updated.WorkKey = workKey
	_, err := r.db.ExecContext(ctx, `UPDATE wishlist_item SET workKey = ?, dedupKey = ? WHERE id = ?`,
		workKey, updated.DedupKey(), item.Id)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return models.ErrDuplicate
	}
	if err != nil {
		return err
	}
	item.WorkKey = workKey
	return nil
}

func (r *SQLiteItemRepository) ListBooksWithoutWork(ctx context.Context) ([]models.Item, error) {
	return r.list(ctx, `SELECT `+itemColumns+itemFrom+` WHERE i.kind = 'book' AND w.workKey IS NULL ORDER BY i.id`)
}

func (r *SQLiteItemRepository) CountByUserId(ctx context.Context, userId int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM wishlist_item WHERE userId = ?`, userId).Scan(&count)
	return count, err
}

func (r *SQLiteItemRepository) list(ctx context.Context, query string, args ...any) ([]models.Item, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []models.Item{}
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.Id, &item.UserId, &item.Kind, &item.Title, &item.Author, &item.WorkKey, &item.AuthorKey, &item.CoverURL, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

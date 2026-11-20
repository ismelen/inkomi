package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const signalColumns = `id, userId, ref, source, kind, title, author, status, attempts, matchedKey`

type SQLiteSignalRepository struct {
	db *sql.DB
}

var _ ports.ISignalRepository = (*SQLiteSignalRepository)(nil)

func NewSQLiteSignalRepository(db *sql.DB) *SQLiteSignalRepository {
	return &SQLiteSignalRepository{db: db}
}

func (r *SQLiteSignalRepository) AddSignal(ctx context.Context, signal *models.Signal) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO user_signal (userId, ref, source, kind, title, author, matchedKey)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (userId, ref) DO NOTHING`,
		signal.UserId, signal.Ref, signal.Source, signal.Kind, signal.Title, signal.Author, signal.MatchedKey)
	if err != nil {
		return false, err
	}
	inserted, err := result.RowsAffected()
	if err != nil || inserted == 0 {
		return false, err
	}
	return true, r.MarkStale(ctx, signal.UserId)
}

func (r *SQLiteSignalRepository) RemoveWishlistExcept(ctx context.Context, refs map[string]bool) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, userId, ref FROM user_signal WHERE source = 'wishlist'`)
	if err != nil {
		return nil, err
	}
	type stale struct {
		id     int64
		userId int
	}
	var gone []stale
	for rows.Next() {
		var s stale
		var ref string
		if err := rows.Scan(&s.id, &s.userId, &ref); err != nil {
			return nil, errors.Join(err, rows.Close())
		}
		if !refs[ref] {
			gone = append(gone, s)
		}
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}

	users := map[int]bool{}
	for _, s := range gone {
		if _, err := r.db.ExecContext(ctx, `DELETE FROM user_signal WHERE id = ?`, s.id); err != nil {
			return nil, err
		}
		users[s.userId] = true
	}
	ids := make([]int, 0, len(users))
	for id := range users {
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *SQLiteSignalRepository) ListPending(ctx context.Context, limit, maxAttempts int) ([]models.Signal, error) {
	return r.list(ctx, `SELECT `+signalColumns+` FROM user_signal
		WHERE status = 'pending' AND attempts < ? ORDER BY attempts, id LIMIT ?`, maxAttempts, limit)
}

func (r *SQLiteSignalRepository) Resolve(ctx context.Context, id int64, status models.SignalStatus, matchedKey string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_signal SET status = ?, matchedKey = CASE WHEN ? <> '' THEN ? ELSE matchedKey END, attempts = attempts + 1
		WHERE id = ?`, status, matchedKey, matchedKey, id)
	return err
}

func (r *SQLiteSignalRepository) ListMatched(ctx context.Context, userId int) ([]models.Signal, error) {
	return r.list(ctx, `SELECT `+signalColumns+` FROM user_signal WHERE userId = ? AND status = 'matched' ORDER BY id`, userId)
}

func (r *SQLiteSignalRepository) ListUsers(ctx context.Context) ([]int, error) {
	return r.ints(ctx, `SELECT DISTINCT userId FROM user_signal WHERE status = 'matched' ORDER BY userId`)
}

func (r *SQLiteSignalRepository) MarkStale(ctx context.Context, userIds ...int) error {
	for _, id := range userIds {
		if _, err := r.db.ExecContext(ctx, `INSERT INTO stale_user (userId) VALUES (?) ON CONFLICT (userId) DO NOTHING`, id); err != nil {
			return err
		}
	}
	return nil
}

func (r *SQLiteSignalRepository) TakeStale(ctx context.Context) ([]int, error) {
	ids, err := r.ints(ctx, `DELETE FROM stale_user RETURNING userId`)
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *SQLiteSignalRepository) Dismiss(ctx context.Context, userId int, workKey string) error {
	if _, err := r.db.ExecContext(ctx, `INSERT INTO dismissed_work (userId, workKey) VALUES (?, ?) ON CONFLICT DO NOTHING`, userId, workKey); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM recommendation WHERE userId = ? AND workKey = ?`, userId, workKey)
	return err
}

func (r *SQLiteSignalRepository) ListDismissed(ctx context.Context, userId int) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT workKey FROM dismissed_work WHERE userId = ?`, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dismissed := map[string]bool{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		dismissed[key] = true
	}
	return dismissed, rows.Err()
}

func (r *SQLiteSignalRepository) GetState(ctx context.Context, name string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM sync_state WHERE name = ?`, name).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return value, err
}

func (r *SQLiteSignalRepository) SetState(ctx context.Context, name, value string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sync_state (name, value) VALUES (?, ?)
		ON CONFLICT (name) DO UPDATE SET value = excluded.value`, name, value)
	return err
}

func (r *SQLiteSignalRepository) list(ctx context.Context, query string, args ...any) ([]models.Signal, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var signals []models.Signal
	for rows.Next() {
		var s models.Signal
		if err := rows.Scan(&s.Id, &s.UserId, &s.Ref, &s.Source, &s.Kind, &s.Title, &s.Author, &s.Status, &s.Attempts, &s.MatchedKey); err != nil {
			return nil, err
		}
		signals = append(signals, s)
	}
	return signals, rows.Err()
}

func (r *SQLiteSignalRepository) ints(ctx context.Context, query string) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

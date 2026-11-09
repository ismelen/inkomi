package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
)

var _ ports.IJoinStore = (*SQLiteJoinStore)(nil)

// SQLiteJoinStore keeps the children of each join group until all of them arrived. Groups expire
// ttl after their last child and claims lockTTL after they were taken.
type SQLiteJoinStore struct {
	db      *sql.DB
	ttl     time.Duration
	lockTTL time.Duration
	now     func() time.Time
	// Add must count, claim and read the group atomically; SQLite allows a single writer anyway.
	mu sync.Mutex
}

func NewSQLiteJoinStore(db *sql.DB, ttl, lockTTL time.Duration, now func() time.Time) *SQLiteJoinStore {
	return &SQLiteJoinStore{
		db:      db,
		ttl:     ttl,
		lockTTL: lockTTL,
		now:     now,
	}
}

func (s *SQLiteJoinStore) Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("join store: marshal msg: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var msgs []jobs.Msg
	err = s.inTx(ctx, func(tx *sql.Tx) error {
		now := s.now()
		nowMs := now.UnixMilli()
		if _, err := tx.ExecContext(ctx, `DELETE FROM join_children WHERE expires_at <= ?`, nowMs); err != nil {
			return fmt.Errorf("purge expired children: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM join_claims WHERE expires_at <= ?`, nowMs); err != nil {
			return fmt.Errorf("purge expired claims: %w", err)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO join_children (parent_id, child_id, payload, expires_at) VALUES (?, ?, ?, ?)
			 ON CONFLICT (parent_id, child_id) DO UPDATE SET payload = excluded.payload`,
			parentId, msg.Id, string(payload), nowMs,
		); err != nil {
			return fmt.Errorf("store child: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE join_children SET expires_at = ? WHERE parent_id = ?`, now.Add(s.ttl).UnixMilli(), parentId,
		); err != nil {
			return fmt.Errorf("refresh group ttl: %w", err)
		}

		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM join_children WHERE parent_id = ?`, parentId).Scan(&count); err != nil {
			return fmt.Errorf("count children: %w", err)
		}
		if count < total {
			return nil
		}

		claim, err := tx.ExecContext(ctx,
			`INSERT INTO join_claims (parent_id, expires_at) VALUES (?, ?) ON CONFLICT (parent_id) DO NOTHING`,
			parentId, now.Add(s.lockTTL).UnixMilli(),
		)
		if err != nil {
			return fmt.Errorf("claim group: %w", err)
		}
		if claimed, err := claim.RowsAffected(); err != nil || claimed == 0 {
			return err
		}

		msgs, err = s.children(ctx, tx, parentId)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("join store: add %s: %w", parentId, err)
	}
	return msgs, nil
}

func (s *SQLiteJoinStore) Clear(ctx context.Context, parentId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.inTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM join_children WHERE parent_id = ?`, parentId); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, `DELETE FROM join_claims WHERE parent_id = ?`, parentId)
		return err
	})
	if err != nil {
		return fmt.Errorf("join store: clear %s: %w", parentId, err)
	}
	return nil
}

func (s *SQLiteJoinStore) Release(ctx context.Context, parentId string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := s.db.ExecContext(ctx, `DELETE FROM join_claims WHERE parent_id = ?`, parentId); err != nil {
		return fmt.Errorf("join store: release %s: %w", parentId, err)
	}
	return nil
}

func (s *SQLiteJoinStore) children(ctx context.Context, tx *sql.Tx, parentId string) ([]jobs.Msg, error) {
	rows, err := tx.QueryContext(ctx, `SELECT payload FROM join_children WHERE parent_id = ?`, parentId)
	if err != nil {
		return nil, fmt.Errorf("read children: %w", err)
	}
	defer rows.Close()

	var msgs []jobs.Msg
	for rows.Next() {
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("scan child: %w", err)
		}
		var msg jobs.Msg
		if err := json.Unmarshal([]byte(payload), &msg); err != nil {
			return nil, fmt.Errorf("unmarshal child: %w", err)
		}
		msgs = append(msgs, msg)
	}
	return msgs, rows.Err()
}

func (s *SQLiteJoinStore) inTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	if err := fn(tx); err != nil {
		return errors.Join(err, tx.Rollback())
	}
	return tx.Commit()
}

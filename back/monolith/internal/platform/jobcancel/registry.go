package jobcancel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
)

var _ jobrun.ICancellations = (*Registry)(nil)

// retention is how long a cancellation is remembered: longer than any job waits in the queue.
const retention = 7 * 24 * time.Hour

// Registry remembers the cancelled job ids (in SQLite, so it survives restarts) and interrupts
// the running jobs that watch them.
type Registry struct {
	db  *sql.DB
	now func() time.Time

	mu        sync.Mutex
	cancelled map[string]bool
	watchers  map[string]map[int]func()
	nextWatch int
}

// NewRegistry loads the cancellations stored in db (table cancelled_job of migrations.Joins).
func NewRegistry(db *sql.DB, now func() time.Time) (*Registry, error) {
	r := &Registry{db: db, now: now, cancelled: map[string]bool{}, watchers: map[string]map[int]func(){}}
	ids, err := r.query(context.Background(), `SELECT id FROM cancelled_job`)
	if err != nil {
		return nil, fmt.Errorf("load cancelled jobs: %w", err)
	}
	for _, id := range ids {
		r.cancelled[id] = true
	}
	return r, nil
}

// Cancel records ids and interrupts the jobs running on any of them.
func (r *Registry) Cancel(ctx context.Context, ids []string) error {
	at := r.now().UnixMilli()
	for _, id := range ids {
		if _, err := r.db.ExecContext(ctx,
			`INSERT INTO cancelled_job (id, at) VALUES (?, ?) ON CONFLICT (id) DO NOTHING`, id, at,
		); err != nil {
			return fmt.Errorf("record cancelled job %s: %w", id, err)
		}
	}

	r.mu.Lock()
	var interrupt []func()
	for _, id := range ids {
		r.cancelled[id] = true
		for _, cancel := range r.watchers[id] {
			interrupt = append(interrupt, cancel)
		}
	}
	r.mu.Unlock()

	// outside the lock: a cancel func may call back into the registry
	for _, cancel := range interrupt {
		cancel()
	}
	return nil
}

func (r *Registry) IsCancelled(ids ...string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range ids {
		if id != "" && r.cancelled[id] {
			return true
		}
	}
	return false
}

// Watch calls cancel when any of ids is cancelled, until stop is called.
func (r *Registry) Watch(ids []string, cancel func()) (stop func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := r.nextWatch
	r.nextWatch++
	for _, id := range ids {
		if id == "" {
			continue
		}
		if r.watchers[id] == nil {
			r.watchers[id] = map[int]func(){}
		}
		r.watchers[id][key] = cancel
	}
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		for _, id := range ids {
			delete(r.watchers[id], key)
			if len(r.watchers[id]) == 0 {
				delete(r.watchers, id)
			}
		}
	}
}

// Start forgets, every interval until ctx is done, the cancellations older than the retention.
func (r *Registry) Start(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := r.Sweep(ctx); err != nil {
				log.Printf("jobcancel: sweep: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (r *Registry) Sweep(ctx context.Context) error {
	cutoff := r.now().Add(-retention).UnixMilli()
	old, err := r.query(ctx, `SELECT id FROM cancelled_job WHERE at < ?`, cutoff)
	if err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, `DELETE FROM cancelled_job WHERE at < ?`, cutoff); err != nil {
		return fmt.Errorf("forget old cancellations: %w", err)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range old {
		delete(r.cancelled, id)
	}
	return nil
}

func (r *Registry) query(ctx context.Context, query string, args ...any) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Join(err, rows.Close())
		}
		ids = append(ids, id)
	}
	return ids, errors.Join(rows.Err(), rows.Close())
}

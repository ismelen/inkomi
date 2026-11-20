package storage

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"
)

// partialUploadTTL is how long a partial upload may sit in tmp before it is considered dead.
const partialUploadTTL = 24 * time.Hour

// Janitor deletes the stored files nobody touched for longer than retention: results already
// downloaded or sent, and uploads that were never confirmed.
type Janitor struct {
	store     *LocalStorage
	retention time.Duration
	now       func() time.Time
}

func NewJanitor(store *LocalStorage, retention time.Duration, now func() time.Time) *Janitor {
	return &Janitor{store: store, retention: retention, now: now}
}

// Start sweeps right away and then every interval until ctx is done.
func (j *Janitor) Start(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			j.Sweep()
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// Sweep removes every id directory whose newest file is older than the retention, and dead
// partial uploads. It returns how many directories it removed.
func (j *Janitor) Sweep() int {
	cutoff := j.now().Add(-j.retention)
	removed := 0

	dirs, err := os.ReadDir(j.store.root)
	if err != nil {
		log.Printf("storage janitor: %v", err)
		return 0
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		dir := filepath.Join(j.store.root, d.Name())
		if newest(dir).After(cutoff) {
			continue
		}
		if err := os.RemoveAll(dir); err != nil {
			log.Printf("storage janitor: remove %s: %v", dir, err)
			continue
		}
		removed++
	}

	if tmps, err := os.ReadDir(j.store.tmp); err == nil {
		for _, t := range tmps {
			if info, err := t.Info(); err == nil && info.ModTime().Before(j.now().Add(-partialUploadTTL)) {
				// best effort: it's tried again on the next sweep
				_ = os.RemoveAll(filepath.Join(j.store.tmp, t.Name()))
			}
		}
	}
	if removed > 0 {
		log.Printf("storage janitor: removed %d expired file groups", removed)
	}
	return removed
}

// newest returns the latest modification time of dir and its files.
func newest(dir string) time.Time {
	latest := time.Time{}
	if info, err := os.Stat(dir); err == nil {
		latest = info.ModTime()
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return latest
	}
	for _, e := range entries {
		if info, err := e.Info(); err == nil && info.ModTime().After(latest) {
			latest = info.ModTime()
		}
	}
	return latest
}

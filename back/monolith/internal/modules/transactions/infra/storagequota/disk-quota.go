package storagequota

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
)

var _ ports.IStorageQuota = (*DiskQuota)(nil)

// unconfirmedTTL is how long a reservation waits for its first confirmed file: the upload urls
// expire after 15 minutes.
const unconfirmedTTL = 30 * time.Minute

// maxReservationAge frees reservations of trees that never finished (e.g. a lost job).
const maxReservationAge = 72 * time.Hour

// IUsage measures what the files of some sources take on disk.
type IUsage interface {
	Usage(ids []string) int64
}

// DiskQuota admits an upload only when its estimated peak fits in the free disk, besides the
// reserve and the space the active uploads still have to fill.
type DiskQuota struct {
	repo    ports.IStorageReservationRepository
	usage   IUsage
	free    func() (uint64, bool)
	reserve int64
	now     func() time.Time
	mu      sync.Mutex
}

func NewDiskQuota(
	repo ports.IStorageReservationRepository,
	usage IUsage,
	free func() (uint64, bool),
	reserve int64,
	now func() time.Time,
) *DiskQuota {
	return &DiskQuota{repo: repo, usage: usage, free: free, reserve: reserve, now: now}
}

func (q *DiskQuota) Exclusive(fn func() error) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	return fn()
}

// Reserve must run inside Exclusive.
func (q *DiskQuota) Reserve(ctx context.Context, reservations []models.StorageReservation) error {
	var need int64
	for _, r := range reservations {
		need += r.Bytes
	}
	if need == 0 {
		return nil
	}

	if free, ok := q.free(); ok {
		pending, err := q.pending(ctx)
		if err != nil {
			return err
		}
		if int64(free)-pending-q.reserve < need {
			log.Printf("storage: rejected %d bytes (free %d, pending %d, reserve %d)", need, free, pending, q.reserve)
			return models.ErrInsufficientStorage
		}
	}

	for i := range reservations {
		if err := q.repo.Create(ctx, &reservations[i]); err != nil {
			return err
		}
	}
	return nil
}

// pending is the space the active reservations still have to take: what they reserved minus
// what their files already use, which statfs counts as used.
func (q *DiskQuota) pending(ctx context.Context) (int64, error) {
	active, err := q.repo.ListActive(ctx)
	if err != nil {
		return 0, err
	}
	var pending int64
	for _, r := range active {
		ids, err := q.repo.TreeIds(ctx, r.RootId)
		if err != nil {
			return 0, err
		}
		pending += max(0, r.Bytes-q.usage.Usage(ids))
	}
	return pending, nil
}

func (q *DiskQuota) ReleaseFinished(ctx context.Context, sourceId string, userId int) error {
	root, err := q.repo.RootOf(ctx, sourceId, userId)
	if errors.Is(err, models.ErrSourceNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	finished, err := q.repo.IsTreeFinished(ctx, root)
	if err != nil || !finished {
		return err
	}
	return q.repo.Release(ctx, root)
}

// Start releases, every interval until ctx is done, the reservations whose uploads were never
// confirmed or that are too old to still be running.
func (q *DiskQuota) Start(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := q.Sweep(ctx); err != nil {
				log.Printf("storage: sweep reservations: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (q *DiskQuota) Sweep(ctx context.Context) error {
	active, err := q.repo.ListActive(ctx)
	if err != nil {
		return err
	}
	for _, r := range active {
		age := q.now().Sub(r.CreatedAt)
		release := age > maxReservationAge
		if !release && age > unconfirmedTTL {
			unconfirmed, err := q.repo.IsTreeUnconfirmed(ctx, r.RootId)
			if err != nil {
				return err
			}
			release = unconfirmed
		}
		if !release {
			continue
		}
		if err := q.repo.Release(ctx, r.RootId); err != nil {
			return fmt.Errorf("release %s: %w", r.RootId, err)
		}
		log.Printf("storage: released the reservation of %s (%s old)", r.RootId, age.Round(time.Minute))
	}
	return nil
}

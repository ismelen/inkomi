package libgen

import (
	"context"
	"log"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"golang.org/x/sync/singleflight"
)

const refreshTimeout = 30 * time.Second

type activeMirror struct {
	IMirror
}

// LibgenLibrary searches and downloads on the fastest known mirror. It renews the mirror
// periodically (Start) and, when a request fails, replaces it and retries once; failure
// driven renewals are spaced by cooldown so a dead network doesn't flood the mirror list.
type LibgenLibrary struct {
	mirrors  IMirrorList
	cooldown time.Duration
	current  atomic.Pointer[activeMirror]
	group    singleflight.Group

	mu          sync.Mutex
	lastRefresh time.Time
}

func NewLibgenLibrary(mirrors IMirrorList, cooldown time.Duration) *LibgenLibrary {
	return &LibgenLibrary{
		mirrors:  mirrors,
		cooldown: cooldown,
	}
}

// Start picks a mirror right away and then renews it every interval until ctx is done.
func (l *LibgenLibrary) Start(ctx context.Context, interval time.Duration) {
	go func() {
		l.refresh(ctx, "")

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.refresh(ctx, "")
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (l *LibgenLibrary) Search(ctx context.Context, query string) ([]models.Book, error) {
	return withFailover(ctx, l, func(m IMirror) ([]models.Book, error) {
		return m.Search(ctx, query)
	})
}

func (l *LibgenLibrary) Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
	return withFailover(ctx, l, func(m IMirror) (*models.BookDownload, error) {
		return m.Download(ctx, md5)
	})
}

func withFailover[T any](ctx context.Context, l *LibgenLibrary, op func(IMirror) (T, error)) (T, error) {
	m, err := l.mirror(ctx)
	if err != nil {
		var zero T
		return zero, err
	}

	result, err := op(m)
	if err == nil || ctx.Err() != nil {
		return result, err
	}

	next, ok := l.replace(ctx, m)
	if !ok {
		return result, err
	}
	log.Printf("libgen: %s failed, retrying on %s: %v", m.URL(), next.URL(), err)
	return op(next)
}

func (l *LibgenLibrary) mirror(ctx context.Context) (IMirror, error) {
	if current := l.current.Load(); current != nil {
		return current.IMirror, nil
	}
	if l.cooldownElapsed() {
		if m, ok := l.refresh(ctx, ""); ok {
			return m, nil
		}
	}
	return nil, models.ErrNoMirror
}

// replace returns a mirror other than failed: the current one if another request already
// replaced it, or a freshly discovered one if the cooldown allows a refresh.
func (l *LibgenLibrary) replace(ctx context.Context, failed IMirror) (IMirror, bool) {
	if current := l.current.Load(); current != nil && current.URL() != failed.URL() {
		return current.IMirror, true
	}
	if !l.cooldownElapsed() {
		return nil, false
	}
	return l.refresh(ctx, failed.URL())
}

func (l *LibgenLibrary) cooldownElapsed() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return time.Since(l.lastRefresh) >= l.cooldown
}

// refresh stores the fastest mirror other than exclude. Concurrent calls share one discovery,
// which outlives the ctx of the caller that started it.
func (l *LibgenLibrary) refresh(ctx context.Context, exclude string) (IMirror, bool) {
	result, _, _ := l.group.Do("refresh", func() (any, error) {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), refreshTimeout)
		defer cancel()

		candidates := slices.DeleteFunc(l.mirrors.Mirrors(ctx), func(m IMirror) bool {
			return m.URL() == exclude
		})
		m, ok := fastestMirror(ctx, candidates)

		l.mu.Lock()
		l.lastRefresh = time.Now()
		l.mu.Unlock()

		if !ok {
			log.Printf("libgen: no mirror answered out of %d", len(candidates))
			return nil, nil
		}
		if current := l.current.Load(); current == nil || current.URL() != m.URL() {
			log.Printf("libgen: using mirror %s", m.URL())
		}
		l.current.Store(&activeMirror{m})
		return m, nil
	})

	m, ok := result.(IMirror)
	return m, ok
}

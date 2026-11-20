package libgen

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"golang.org/x/sync/singleflight"
)

const refreshTimeout = 30 * time.Second

// LibgenLibrary searches and downloads on the fastest known mirror. It renews the mirror
// periodically (Start) and, when a request fails, replaces it and retries once; failure
// driven renewals are spaced by cooldown so a dead network doesn't flood the mirror list.
type LibgenLibrary struct {
	mirrors  IMirrorList
	cooldown time.Duration
	group    singleflight.Group

	mu          sync.Mutex
	current     IMirror
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

func (l *LibgenLibrary) Search(ctx context.Context, query string, field models.SearchField) ([]models.Book, error) {
	return withFailover(ctx, l, func(m IMirror) ([]models.Book, error) {
		return m.Search(ctx, query, field)
	})
}

// Download tries the current mirror and then every other known mirror: a file can fail on one
// (its download server answering 5xx) and work on another. It returns models.ErrBookNotFound
// only when no mirror has the book.
func (l *LibgenLibrary) Download(ctx context.Context, md5 models.MD5) (*models.BookDownload, error) {
	first, ok := l.mirror(ctx, "")
	if !ok {
		return nil, models.ErrNoMirror
	}
	book, err := first.Download(ctx, md5)
	if err == nil || ctx.Err() != nil {
		return book, err
	}

	errs := []error{err}
	notFound := errors.Is(err, models.ErrBookNotFound)
	for _, m := range l.mirrors.Mirrors(ctx) {
		if m.URL() == first.URL() {
			continue
		}
		book, err := m.Download(ctx, md5)
		if err == nil {
			log.Printf("libgen: %s downloaded on %s after %d failed mirrors", md5, m.URL(), len(errs))
			return book, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		errs = append(errs, err)
		notFound = notFound && errors.Is(err, models.ErrBookNotFound)
	}

	joined := errors.Join(errs...)
	if notFound {
		return nil, joined
	}
	// a mirror that failed for another reason may have it: not a permanent failure
	return nil, fmt.Errorf("every mirror failed:\n%s", joined.Error())
}

func withFailover[T any](ctx context.Context, l *LibgenLibrary, op func(IMirror) (T, error)) (T, error) {
	m, ok := l.mirror(ctx, "")
	if !ok {
		var zero T
		return zero, models.ErrNoMirror
	}

	result, err := op(m)
	if err == nil || ctx.Err() != nil {
		return result, err
	}

	next, ok := l.mirror(ctx, m.URL())
	if !ok {
		return result, err
	}
	log.Printf("libgen: %s failed, retrying on %s: %v", m.URL(), next.URL(), err)
	return op(next)
}

// mirror returns the current mirror unless it is exclude (another request may have already
// replaced it); otherwise it discovers a new one if the cooldown allows it.
func (l *LibgenLibrary) mirror(ctx context.Context, exclude string) (IMirror, bool) {
	l.mu.Lock()
	current, canRefresh := l.current, time.Since(l.lastRefresh) >= l.cooldown
	l.mu.Unlock()

	if current != nil && current.URL() != exclude {
		return current, true
	}
	if !canRefresh {
		return nil, false
	}
	return l.refresh(ctx, exclude)
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
		previous := l.current
		l.lastRefresh = time.Now()
		if ok {
			l.current = m
		}
		l.mu.Unlock()

		if !ok {
			log.Printf("libgen: no mirror answered out of %d", len(candidates))
			return nil, nil
		}
		if previous == nil || previous.URL() != m.URL() {
			log.Printf("libgen: using mirror %s", m.URL())
		}
		return m, nil
	})

	m, ok := result.(IMirror)
	return m, ok
}

package releasecache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/domain/ports"
	"golang.org/x/sync/singleflight"
)

var _ ports.IReleaseSource = (*CachedReleaseSource)(nil)

// CachedReleaseSource keeps the latest release for ttl so every app start doesn't hit the
// source (the GitHub API allows 60 anonymous requests per hour). When a refresh fails with the
// source unavailable, the stale release is served.
type CachedReleaseSource struct {
	source ports.IReleaseSource
	ttl    time.Duration
	now    func() time.Time
	group  singleflight.Group

	mu        sync.RWMutex
	release   *models.Release
	fetchedAt time.Time
}

func NewCachedReleaseSource(source ports.IReleaseSource, ttl time.Duration) *CachedReleaseSource {
	return &CachedReleaseSource{source: source, ttl: ttl, now: time.Now}
}

// WithClock replaces the clock; for tests.
func (c *CachedReleaseSource) WithClock(now func() time.Time) *CachedReleaseSource {
	c.now = now
	return c
}

func (c *CachedReleaseSource) Latest(ctx context.Context) (*models.Release, error) {
	if release, fresh := c.cached(); fresh {
		return release, nil
	}

	value, err, _ := c.group.Do("latest", func() (any, error) {
		// The request that triggered the refresh may be cancelled; the refresh serves everyone.
		return c.source.Latest(context.WithoutCancel(ctx))
	})
	if err != nil {
		if stale, _ := c.cached(); stale != nil && errors.Is(err, models.ErrSourceUnavailable) {
			return stale, nil
		}
		return nil, err
	}

	release := value.(*models.Release)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.release, c.fetchedAt = release, c.now()
	return release, nil
}

func (c *CachedReleaseSource) cached() (*models.Release, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.release == nil {
		return nil, false
	}
	copied := *c.release
	return &copied, c.now().Sub(c.fetchedAt) < c.ttl
}

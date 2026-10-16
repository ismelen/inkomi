package releasecache_test

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/releasecache"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type clock struct{ now time.Time }

func (c *clock) Now() time.Time          { return c.now }
func (c *clock) Advance(d time.Duration) { c.now = c.now.Add(d) }

func release(patch int) *models.Release {
	return &models.Release{Version: models.Version{Major: 1, Patch: patch}}
}

func TestCachedReleaseSource_Latest_WithinTTL_CallsSourceOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	var calls atomic.Int32
	source := &mocks.MockReleaseSource{LatestFn: func(ctx context.Context) (*models.Release, error) {
		calls.Add(1)
		return release(1), nil
	}}
	c := &clock{now: time.Now()}
	cached := releasecache.NewCachedReleaseSource(source, time.Minute).WithClock(c.Now)

	// Act
	_, err := cached.Latest(context.Background())
	require.NoError(t, err)
	c.Advance(30 * time.Second)
	got, err := cached.Latest(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int32(1), calls.Load())
	assert.Equal(t, 1, got.Version.Patch)
}

func TestCachedReleaseSource_Latest_Expired_Refreshes(t *testing.T) {
	t.Parallel()

	// Arrange
	var calls atomic.Int32
	source := &mocks.MockReleaseSource{LatestFn: func(ctx context.Context) (*models.Release, error) {
		return release(int(calls.Add(1))), nil
	}}
	c := &clock{now: time.Now()}
	cached := releasecache.NewCachedReleaseSource(source, time.Minute).WithClock(c.Now)

	// Act
	_, err := cached.Latest(context.Background())
	require.NoError(t, err)
	c.Advance(2 * time.Minute)
	got, err := cached.Latest(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, got.Version.Patch)
}

func TestCachedReleaseSource_Latest_RefreshFails_HandlesError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		err       error
		wantStale bool
	}{
		{"source unavailable serves stale", fmt.Errorf("%w: boom", models.ErrSourceUnavailable), true},
		{"other error is returned", models.ErrAPKNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fail := false
			source := &mocks.MockReleaseSource{LatestFn: func(ctx context.Context) (*models.Release, error) {
				if fail {
					return nil, tt.err
				}
				return release(1), nil
			}}
			c := &clock{now: time.Now()}
			cached := releasecache.NewCachedReleaseSource(source, time.Minute).WithClock(c.Now)
			_, err := cached.Latest(context.Background())
			require.NoError(t, err)
			fail = true
			c.Advance(2 * time.Minute)

			// Act
			got, err := cached.Latest(context.Background())

			// Assert
			if tt.wantStale {
				require.NoError(t, err)
				assert.Equal(t, 1, got.Version.Patch)
			} else {
				assert.True(t, errors.Is(err, tt.err))
				assert.Nil(t, got)
			}
		})
	}
}

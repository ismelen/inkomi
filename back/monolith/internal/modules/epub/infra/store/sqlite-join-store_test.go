package store_test

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/store"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/sqlitedb"
	"github.com/ismelen/inkomi/back/monolith/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ttl     = time.Hour
	lockTTL = 10 * time.Minute
)

// fakeClock is a clock the tests move forward by hand.
type fakeClock struct {
	now atomic.Int64
}

func (c *fakeClock) Now() time.Time {
	return time.UnixMilli(c.now.Load())
}

func (c *fakeClock) Advance(d time.Duration) {
	c.now.Add(d.Milliseconds())
}

func newStore(t *testing.T) (*store.SQLiteJoinStore, *fakeClock) {
	t.Helper()

	db, err := sqlitedb.Open(filepath.Join(t.TempDir(), "jobs.db"), migrations.Joins)
	require.NoError(t, err)
	t.Cleanup(func() {
		// the database lives in a temp dir that is removed anyway
		_ = db.Close()
	})

	clock := &fakeClock{}
	clock.now.Store(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli())
	return store.NewSQLiteJoinStore(db, ttl, lockTTL, clock.Now), clock
}

func childMsg(id string) jobs.Msg {
	return jobs.Msg{Id: id, UserId: 1, Filename: id + ".epub", Parent: &jobs.Parent{Id: "parent-1", Total: 3}}
}

// fill adds n distinct children and returns what the last Add returned.
func fill(t *testing.T, s *store.SQLiteJoinStore, n int) []jobs.Msg {
	t.Helper()

	var last []jobs.Msg
	for i := range n {
		var err error
		last, err = s.Add(context.Background(), "parent-1", n, childMsg(fmt.Sprint(i)))
		require.NoError(t, err)
	}
	return last
}

func TestSQLiteJoinStore_Add_UntilTotal_ReturnsAllAndClaims(t *testing.T) {
	t.Parallel()

	// Arrange
	s, _ := newStore(t)
	ctx := context.Background()

	// Act
	first, err1 := s.Add(ctx, "parent-1", 3, childMsg("a"))
	second, err2 := s.Add(ctx, "parent-1", 3, childMsg("b"))
	last, err3 := s.Add(ctx, "parent-1", 3, childMsg("c"))

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	assert.Nil(t, first)
	assert.Nil(t, second)
	assert.ElementsMatch(t, []jobs.Msg{childMsg("a"), childMsg("b"), childMsg("c")}, last)
}

func TestSQLiteJoinStore_Add_DuplicateChild_CountsOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	s, _ := newStore(t)
	ctx := context.Background()

	// Act
	_, err1 := s.Add(ctx, "parent-1", 2, childMsg("a"))
	dup, err2 := s.Add(ctx, "parent-1", 2, childMsg("a"))
	last, err3 := s.Add(ctx, "parent-1", 2, childMsg("b"))

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	assert.Nil(t, dup)
	assert.Len(t, last, 2)
}

func TestSQLiteJoinStore_Add_ClaimedGroup_ReturnsNil(t *testing.T) {
	t.Parallel()

	// Arrange
	s, _ := newStore(t)
	require.Len(t, fill(t, s, 2), 2)

	// Act
	redelivered, err := s.Add(context.Background(), "parent-1", 2, childMsg("1"))

	// Assert
	require.NoError(t, err)
	assert.Nil(t, redelivered)
}

func TestSQLiteJoinStore_Release_ClaimedGroup_CanBeClaimedAgain(t *testing.T) {
	t.Parallel()

	// Arrange
	s, _ := newStore(t)
	ctx := context.Background()
	require.Len(t, fill(t, s, 2), 2)

	// Act
	err := s.Release(ctx, "parent-1")
	redelivered, addErr := s.Add(ctx, "parent-1", 2, childMsg("1"))

	// Assert
	require.NoError(t, err)
	require.NoError(t, addErr)
	assert.Len(t, redelivered, 2)
}

func TestSQLiteJoinStore_Add_ExpiredLock_CanBeClaimedAgain(t *testing.T) {
	t.Parallel()

	// Arrange
	s, clock := newStore(t)
	require.Len(t, fill(t, s, 2), 2)

	// Act
	clock.Advance(lockTTL)
	redelivered, err := s.Add(context.Background(), "parent-1", 2, childMsg("1"))

	// Assert
	require.NoError(t, err)
	assert.Len(t, redelivered, 2)
}

func TestSQLiteJoinStore_Add_ExpiredGroup_StartsOver(t *testing.T) {
	t.Parallel()

	// Arrange
	s, clock := newStore(t)
	ctx := context.Background()
	_, err := s.Add(ctx, "parent-1", 2, childMsg("a"))
	require.NoError(t, err)

	// Act
	clock.Advance(ttl)
	msgs, err := s.Add(ctx, "parent-1", 2, childMsg("b"))

	// Assert
	require.NoError(t, err)
	assert.Nil(t, msgs)
}

func TestSQLiteJoinStore_Clear_ClaimedGroup_RemovesGroupAndClaim(t *testing.T) {
	t.Parallel()

	// Arrange
	s, _ := newStore(t)
	ctx := context.Background()
	require.Len(t, fill(t, s, 2), 2)

	// Act
	err := s.Clear(ctx, "parent-1")
	afterClear, addErr := s.Add(ctx, "parent-1", 2, childMsg("0"))

	// Assert
	require.NoError(t, err)
	require.NoError(t, addErr)
	assert.Nil(t, afterClear)
}

func TestSQLiteJoinStore_Add_Concurrent_ClaimsOnce(t *testing.T) {
	t.Parallel()

	// Arrange
	s, _ := newStore(t)
	const total = 20
	var wg sync.WaitGroup
	results := make(chan []jobs.Msg, total)

	// Act
	for i := range total {
		wg.Go(func() {
			msgs, err := s.Add(context.Background(), "parent-1", total, childMsg(fmt.Sprint(i)))
			assert.NoError(t, err)
			results <- msgs
		})
	}
	wg.Wait()
	close(results)

	// Assert
	claimed := 0
	for msgs := range results {
		if msgs != nil {
			claimed++
			assert.Len(t, msgs, total)
		}
	}
	assert.Equal(t, 1, claimed)
}

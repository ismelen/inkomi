package store_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/infra/store"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ttl = time.Hour

func newStore(t *testing.T) (*store.RedisJoinStore, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() {
		// closing a client of the in-memory server can't fail in a meaningful way
		_ = client.Close()
	})
	return store.NewRedisJoinStore(client, ttl), mr
}

func childMsg(id string) jobs.Msg {
	return jobs.Msg{Id: id, UserId: 1, Filename: id + ".epub", Parent: &jobs.Parent{Id: "parent-1", Total: 3}}
}

func TestRedisJoinStore_Add_UntilTotal_ReturnsAllAndClears(t *testing.T) {
	t.Parallel()

	// Arrange
	s, mr := newStore(t)
	ctx := context.Background()

	// Act
	first, err1 := s.Add(ctx, "parent-1", 3, childMsg("a"))
	second, err2 := s.Add(ctx, "parent-1", 3, childMsg("b"))
	ttlLeft := mr.TTL("join:parent-1")
	last, err3 := s.Add(ctx, "parent-1", 3, childMsg("c"))

	// Assert
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)
	assert.Nil(t, first)
	assert.Nil(t, second)
	assert.Equal(t, ttl, ttlLeft)
	assert.ElementsMatch(t, []jobs.Msg{childMsg("a"), childMsg("b"), childMsg("c")}, last)
	assert.False(t, mr.Exists("join:parent-1"))
}

func TestRedisJoinStore_Add_DuplicateChild_CountsOnce(t *testing.T) {
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

func TestRedisJoinStore_Add_Concurrent_CompletesOnce(t *testing.T) {
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
	completed := 0
	for msgs := range results {
		if msgs != nil {
			completed++
			assert.Len(t, msgs, total)
		}
	}
	assert.Equal(t, 1, completed)
}

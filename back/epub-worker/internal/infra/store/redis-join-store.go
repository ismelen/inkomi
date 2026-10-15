package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/redis/go-redis/v9"
)

const (
	joinKeyPrefix = "join:"
	lockKeySuffix = ":lock"
)

// addScript stores a child payload in the parent hash and, when the hash holds
// total entries and nobody holds the lock, takes the lock and returns every payload.
var addScript = redis.NewScript(`
redis.call("HSET", KEYS[1], ARGV[1], ARGV[2])
redis.call("EXPIRE", KEYS[1], ARGV[4])
if redis.call("HLEN", KEYS[1]) < tonumber(ARGV[3]) then
	return false
end
if not redis.call("SET", KEYS[2], "1", "NX", "EX", ARGV[5]) then
	return false
end
return redis.call("HVALS", KEYS[1])
`)

type RedisJoinStore struct {
	client  *redis.Client
	ttl     time.Duration
	lockTTL time.Duration
}

func NewRedisJoinStore(client *redis.Client, ttl, lockTTL time.Duration) *RedisJoinStore {
	return &RedisJoinStore{
		client:  client,
		ttl:     ttl,
		lockTTL: lockTTL,
	}
}

func (s *RedisJoinStore) Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("join store: marshal msg: %w", err)
	}

	values, err := addScript.Run(ctx, s.client,
		[]string{groupKey(parentId), lockKey(parentId)},
		msg.Id, payload, total, int(s.ttl.Seconds()), int(s.lockTTL.Seconds()),
	).StringSlice()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("join store: add %s: %w", parentId, err)
	}

	msgs := make([]jobs.Msg, len(values))
	for i, v := range values {
		if err := json.Unmarshal([]byte(v), &msgs[i]); err != nil {
			return nil, fmt.Errorf("join store: unmarshal msg: %w", err)
		}
	}

	return msgs, nil
}

func (s *RedisJoinStore) Clear(ctx context.Context, parentId string) error {
	if err := s.client.Del(ctx, groupKey(parentId), lockKey(parentId)).Err(); err != nil {
		return fmt.Errorf("join store: clear %s: %w", parentId, err)
	}
	return nil
}

func (s *RedisJoinStore) Release(ctx context.Context, parentId string) error {
	if err := s.client.Del(ctx, lockKey(parentId)).Err(); err != nil {
		return fmt.Errorf("join store: release %s: %w", parentId, err)
	}
	return nil
}

func groupKey(parentId string) string {
	return joinKeyPrefix + parentId
}

func lockKey(parentId string) string {
	return joinKeyPrefix + parentId + lockKeySuffix
}

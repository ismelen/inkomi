package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/redis/go-redis/v9"
)

const joinKeyPrefix = "join:"

// addScript stores a child payload in the parent hash and, when the hash holds
// total entries, returns every payload and deletes the hash atomically.
var addScript = redis.NewScript(`
redis.call("HSET", KEYS[1], ARGV[1], ARGV[2])
redis.call("EXPIRE", KEYS[1], ARGV[4])
if redis.call("HLEN", KEYS[1]) < tonumber(ARGV[3]) then
	return false
end
local values = redis.call("HVALS", KEYS[1])
redis.call("DEL", KEYS[1])
return values
`)

type RedisJoinStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisJoinStore(client *redis.Client, ttl time.Duration) *RedisJoinStore {
	return &RedisJoinStore{
		client: client,
		ttl:    ttl,
	}
}

func (s *RedisJoinStore) Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error) {
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("join store: marshal msg: %w", err)
	}

	values, err := addScript.Run(ctx, s.client,
		[]string{joinKeyPrefix + parentId},
		msg.Id, payload, total, int(s.ttl.Seconds()),
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

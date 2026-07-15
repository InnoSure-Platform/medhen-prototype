// Package idempotency provides distributed idempotency-key storage (ADR-PC-007).
package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrReplay = errors.New("idempotency key already used")

type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

func NewStore(rdb *redis.Client, ttl time.Duration) *Store {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}
	return &Store{rdb: rdb, ttl: ttl}
}

// Begin reserves the key. If already present, returns the cached response and ErrReplay.
func (s *Store) Begin(ctx context.Context, scope, key string) (cached []byte, replay bool, err error) {
	if key == "" {
		return nil, false, nil
	}
	redisKey := "idem:" + scope + ":" + key
	ok, err := s.rdb.SetNX(ctx, redisKey+":lock", "1", s.ttl).Result()
	if err != nil {
		return nil, false, err
	}
	if !ok {
		val, err := s.rdb.Get(ctx, redisKey+":body").Bytes()
		if err == redis.Nil {
			return nil, true, ErrReplay
		}
		return val, true, err
	}
	return nil, false, nil
}

func (s *Store) Complete(ctx context.Context, scope, key string, body any) error {
	if key == "" {
		return nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	redisKey := "idem:" + scope + ":" + key
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, redisKey+":body", b, s.ttl)
	pipe.Del(ctx, redisKey+":lock")
	_, err = pipe.Exec(ctx)
	return err
}

// MemoryStore is a process-local fallback for single-node demos without Valkey.
type MemoryStore struct {
	m map[string][]byte
}

func NewMemoryStore() *MemoryStore { return &MemoryStore{m: map[string][]byte{}} }

func (s *MemoryStore) Begin(_ context.Context, scope, key string) ([]byte, bool, error) {
	if key == "" {
		return nil, false, nil
	}
	k := scope + ":" + key
	if v, ok := s.m[k]; ok {
		return v, true, nil
	}
	s.m[k] = nil // reserved
	return nil, false, nil
}

func (s *MemoryStore) Complete(_ context.Context, scope, key string, body any) error {
	if key == "" {
		return nil
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	s.m[scope+":"+key] = b
	return nil
}

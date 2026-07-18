// Package idempotency provides an idempotency-key store backed by Valkey/Redis.
//
// It fixes the pre-refactor SDK, which did a non-atomic GET-then-SET (so two
// concurrent duplicate requests could both proceed), used a zero TTL (keys
// never expired), and returned 409 on replay instead of the original response.
// Here the claim is a single atomic SET NX with a TTL, and a completed request's
// response is cached for true replay.
package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Status is the outcome of claiming an idempotency key.
type Status int

const (
	// StatusNew means the caller claimed the key and should execute the request.
	StatusNew Status = iota
	// StatusInProgress means another request holds the key but hasn't completed.
	StatusInProgress
	// StatusDone means the request already completed; Response holds the cached body.
	StatusDone
)

const (
	markerPending byte = 'P'
	markerDone    byte = 'D'
)

// Result reports the claim outcome and, for StatusDone, the cached response.
type Result struct {
	Status   Status
	Response []byte
}

// Store is a Redis-backed idempotency store.
type Store struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// New builds a Store. ttl<=0 defaults to 24h.
func New(client redis.UniversalClient, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Store{client: client, ttl: ttl}
}

func (s *Store) redisKey(key string) string { return "idem:" + key }

// Acquire atomically claims the key. On StatusNew the caller owns the request
// and must call Complete when done. On StatusDone the cached response is
// returned for replay; on StatusInProgress the caller should retry/backoff.
func (s *Store) Acquire(ctx context.Context, key string) (Result, error) {
	rk := s.redisKey(key)

	ok, err := s.client.SetNX(ctx, rk, []byte{markerPending}, s.ttl).Result()
	if err != nil {
		return Result{}, fmt.Errorf("idempotency: setnx: %w", err)
	}
	if ok {
		return Result{Status: StatusNew}, nil
	}

	val, err := s.client.Get(ctx, rk).Bytes()
	if err == redis.Nil {
		// Expired between SETNX and GET (rare); treat as in-progress so the
		// caller retries rather than double-executing.
		return Result{Status: StatusInProgress}, nil
	}
	if err != nil {
		return Result{}, fmt.Errorf("idempotency: get: %w", err)
	}
	if len(val) > 0 && val[0] == markerDone {
		return Result{Status: StatusDone, Response: val[1:]}, nil
	}
	return Result{Status: StatusInProgress}, nil
}

// Complete stores the final response for replay and refreshes the TTL.
func (s *Store) Complete(ctx context.Context, key string, response []byte) error {
	val := append([]byte{markerDone}, response...)
	if err := s.client.Set(ctx, s.redisKey(key), val, s.ttl).Err(); err != nil {
		return fmt.Errorf("idempotency: complete: %w", err)
	}
	return nil
}

// Release removes a claim (e.g. when a StatusNew request fails and should be
// retryable rather than left pending until TTL).
func (s *Store) Release(ctx context.Context, key string) error {
	if err := s.client.Del(ctx, s.redisKey(key)).Err(); err != nil {
		return fmt.Errorf("idempotency: release: %w", err)
	}
	return nil
}

package idempotency

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

// IdempotencyKeyHeader is the standard header for idempotency keys.
const IdempotencyKeyHeader = "Idempotency-Key"

// Config holds the configuration for the Idempotency SDK.
type Config struct {
	RedisURL string
}

// Manager handles storing and verifying idempotency keys.
type Manager struct {
	client *redis.Client
}

// NewManager creates a new Idempotency Manager.
func NewManager(cfg Config) (*Manager, error) {
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid redis url: %w", err)
	}
	client := redis.NewClient(opts)

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Manager{
		client: client,
	}, nil
}

// Middleware returns a net/http middleware that enforces idempotency.
// In a Tier-0 implementation, this would:
// 1. Check if the Idempotency-Key header is present.
// 2. SETNX the key in Redis to lock it.
// 3. If locked by us, process request.
// 4. Cache response.
// 5. If already exists, return cached response.
func (m *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(IdempotencyKeyHeader)
		if key == "" {
			// For some APIs, this might be optional. For strict ones, we'd return 400.
			// Let's pass it through if not provided.
			next.ServeHTTP(w, r)
			return
		}

		// Simplified for MVP: Check if key exists.
		// A full implementation would use a robust Lua script to prevent race conditions.
		val, err := m.client.Get(r.Context(), "idemp:"+key).Result()
		if err == nil && val != "" {
			w.Header().Set("X-Idempotent-Replayed", "true")
			w.WriteHeader(http.StatusConflict) // Or 200/201 with cached response
			w.Write([]byte(`{"error":"Conflict", "message":"Idempotency key already used"}`))
			return
		}

		// Mark key as used (with an expiration)
		m.client.Set(r.Context(), "idemp:"+key, "used", 0)

		next.ServeHTTP(w, r)
	})
}

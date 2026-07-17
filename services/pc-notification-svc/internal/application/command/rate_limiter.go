package command

import (
	"context"
	"fmt"
	"time"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RateLimiter struct {
	rdb *redis.Client
}

func NewRateLimiter(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb}
}

// AllowMarketing checks if the party has exceeded the max daily marketing messages (e.g., 3)
func (rl *RateLimiter) AllowMarketing(ctx context.Context, partyID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("rate_limit:marketing:%s:%s", partyID.String(), time.Now().UTC().Format("2006-01-02"))
	
	count, err := rl.rdb.Incr(ctx, key).Result()
	if err != nil {
		return true, err // Fail open so we don't break notifications if redis is down
	}

	if count == 1 {
		rl.rdb.Expire(ctx, key, 24*time.Hour)
	}

	if count > 3 {
		return false, nil
	}
	return true, nil
}

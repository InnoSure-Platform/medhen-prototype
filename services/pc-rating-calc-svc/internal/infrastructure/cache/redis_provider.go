package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/singleflight"
)

// RedisRateProvider implements the RateTableProvider using Redis
type RedisRateProvider struct {
	client *redis.Client
	group  singleflight.Group
}

func NewRedisRateProvider(client *redis.Client) *RedisRateProvider {
	return &RedisRateProvider{client: client}
}

// GetBaseRate retrieves the base rate using Singleflight cache stampede protection
func (r *RedisRateProvider) GetBaseRate(ctx context.Context, productCode string, coverageCode string, dims map[string]string) (decimal.Decimal, string, error) {
	key := fmt.Sprintf("base:%s:%s", productCode, coverageCode)
	
	// singleflight ensures only 1 goroutine executes this simultaneously for the same key
	val, err, _ := r.group.Do(key, func() (interface{}, error) {
		// Mock Redis lookup
		d, _ := decimal.NewFromString("1000.00")
		return d, nil
	})

	if err != nil {
		return decimal.Zero, "", err
	}
	return val.(decimal.Decimal), "v1-2026", nil
}

// GetFactor retrieves a multiplier factor using Singleflight
func (r *RedisRateProvider) GetFactor(ctx context.Context, productCode string, coverageCode string, factorType string, dims map[string]string) (decimal.Decimal, string, error) {
	key := fmt.Sprintf("factor:%s:%s:%s", productCode, coverageCode, factorType)
	
	val, err, _ := r.group.Do(key, func() (interface{}, error) {
		// Mock Redis lookup
		d, _ := decimal.NewFromString("1.50")
		return d, nil
	})

	if err != nil {
		return decimal.Zero, "", err
	}
	return val.(decimal.Decimal), "v1-2026", nil
}

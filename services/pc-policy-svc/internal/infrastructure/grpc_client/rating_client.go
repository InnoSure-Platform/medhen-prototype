package grpc_client

import (
	"context"
	"errors"
	"time"

	"github.com/sony/gobreaker"
)

type RatingClient struct {
	cb *gobreaker.CircuitBreaker
	// conn *grpc.ClientConn
}

func NewRatingClient() *RatingClient {
	st := gobreaker.Settings{
		Name:        "RatingService",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}

	return &RatingClient{
		cb: gobreaker.NewCircuitBreaker(st),
	}
}

// CalculatePremium wraps the gRPC call in a circuit breaker.
func (c *RatingClient) CalculatePremium(ctx context.Context, payload []byte) (float64, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		// Mock synchronous gRPC call to pc-rating-calc-svc
		// if err != nil { return nil, err }
		// return resp.Premium, nil
		
		// Simulate successful call for now
		return 500.00, nil
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return 0, errors.New("rating service is currently unavailable (circuit open)")
		}
		return 0, err
	}

	return result.(float64), nil
}

package acl

import (
	"context"
	"errors"
	"time"
	"pc-notification-svc/internal/domain/notification"
	"github.com/sony/gobreaker"
)

type ResilientClient struct {
	baseClient Client
	cb         *gobreaker.CircuitBreaker
}

func NewResilientClient(base Client) *ResilientClient {
	st := gobreaker.Settings{
		Name:        "pc-integration-acl",
		MaxRequests: 1,
		Interval:    10 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 5 && failureRatio >= 0.6
		},
	}

	return &ResilientClient{
		baseClient: base,
		cb:         gobreaker.NewCircuitBreaker(st),
	}
}

func (c *ResilientClient) Dispatch(ctx context.Context, channel notification.Channel, address string, content string) (string, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.baseClient.Dispatch(ctx, channel, address, content)
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", errors.New("CIRCUIT_BREAKER_OPEN: integration ACL is down")
		}
		return "", err
	}

	return result.(string), nil
}

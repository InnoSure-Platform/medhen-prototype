package grpc_client

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UWClient struct {
	cb   *gobreaker.CircuitBreaker
	conn *grpc.ClientConn
	// In a real implementation with compiled protos, we would have:
	// client pb.UnderwritingServiceClient
}

func NewUWClient(target string) (*UWClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	st := gobreaker.Settings{
		Name:        "UnderwritingService",
		MaxRequests: 3,
		Interval:    5 * time.Second,
		Timeout:     15 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.5
		},
	}

	return &UWClient{
		cb:   gobreaker.NewCircuitBreaker(st),
		conn: conn,
	}, nil
}

// AssessQuote calls pc-underwriting-svc to check if the risk triggers any referral rules.
func (c *UWClient) AssessQuote(ctx context.Context, tenantID string, quoteID uuid.UUID, productID string, payload []byte) (string, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		// Simulate the actual gRPC call. Since we don't have the compiled proto in pc-contracts yet,
		// we represent the call pattern here.
		// req := &pb.AssessQuoteRequest{ ... }
		// resp, err := c.client.AssessQuote(ctx, req)
		// return resp.Status, err
		
		// For the sake of execution flow, we simulate a successful "APPROVED" response from the real service.
		return "APPROVED", nil
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", errors.New("underwriting service is currently unavailable (circuit open)")
		}
		return "", err
	}

	return result.(string), nil
}

func (c *UWClient) Close() error {
	return c.conn.Close()
}

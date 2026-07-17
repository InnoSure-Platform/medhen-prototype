package acl

import (
	"context"
	"pc-notification-svc/internal/domain/notification"
)

type Client interface {
	Dispatch(ctx context.Context, channel notification.Channel, address string, content string) (string, error)
}

type grpcClient struct{}

func NewClient() Client {
	return &grpcClient{}
}

func (c *grpcClient) Dispatch(ctx context.Context, channel notification.Channel, address string, content string) (string, error) {
	// Mock dispatch to pc-integration-acl
	return "receipt-12345", nil
}

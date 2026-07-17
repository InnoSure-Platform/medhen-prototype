package iam

import (
	"context"
	"log/slog"
)

// Client is a gRPC stub for communicating with pc-iam-svc.
type Client struct {
	logger *slog.Logger
	// conn *grpc.ClientConn
}

func NewClient(logger *slog.Logger) *Client {
	return &Client{logger: logger}
}

// ResolveRoleToUsers queries pc-iam-svc to return a list of eligible User IDs.
func (c *Client) ResolveRoleToUsers(ctx context.Context, roleExpression string, contextPayload []byte) ([]string, error) {
	c.logger.Info("Resolving role to users via IAM", "role_expr", roleExpression)
	// In a real implementation, this would make a gRPC call to pc-iam-svc
	// Example: return c.grpcClient.ResolveAssignees(ctx, ...)
	
	// Stub response
	return []string{"usr-123"}, nil
}

// GetManager returns the manager of a given user.
func (c *Client) GetManager(ctx context.Context, userID string) (string, error) {
	c.logger.Info("Fetching manager for user", "user_id", userID)
	return "mgr-999", nil
}

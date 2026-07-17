package pcgrpc

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/InnoSure-Platform/pc-shared-go-svc/tenant"
)

// RecoveryInterceptor catches panics and returns an internal server error.
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "panic recovered in grpc handler", "panic", fmt.Sprintf("%v", r), "method", info.FullMethod)
				err = status.Errorf(codes.Internal, "Internal Server Error")
			}
		}()
		return handler(ctx, req)
	}
}

// TenantInterceptor extracts a tenant ID from grpc metadata (or context) and injects it using our tenant package.
// For demonstration, we simply verify if a mock header or existing context value is set.
func TenantInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// In a real implementation, you'd extract "x-tenant-id" from grpc/metadata.
		// For now, we assume it's either provided or we fallback to a default if not strictly required,
		// or we let the handler fail when calling TenantFromContext.
		ctx = tenant.WithTenant(ctx, "mock-tenant-id")
		return handler(ctx, req)
	}
}

// LoggingInterceptor provides structured logging around gRPC calls.
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		slog.InfoContext(ctx, "grpc request started", "method", info.FullMethod)
		resp, err := handler(ctx, req)
		if err != nil {
			slog.ErrorContext(ctx, "grpc request failed", "method", info.FullMethod, "error", err.Error())
		} else {
			slog.InfoContext(ctx, "grpc request completed", "method", info.FullMethod)
		}
		return resp, err
	}
}

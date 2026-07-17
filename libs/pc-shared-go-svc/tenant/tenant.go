package tenant

import (
	"context"
	"errors"
)

type contextKey string

const (
	tenantCtxKey contextKey = "tenant_id"
	actorCtxKey  contextKey = "actor_id"
)

var (
	ErrTenantNotFound = errors.New("tenant ID not found in context")
	ErrActorNotFound  = errors.New("actor ID not found in context")
)

// WithTenant injects a tenant ID into the context.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantCtxKey, tenantID)
}

// TenantFromContext extracts the tenant ID from the context.
func TenantFromContext(ctx context.Context) (string, error) {
	val, ok := ctx.Value(tenantCtxKey).(string)
	if !ok || val == "" {
		return "", ErrTenantNotFound
	}
	return val, nil
}

// WithActor injects an actor ID into the context.
func WithActor(ctx context.Context, actorID string) context.Context {
	return context.WithValue(ctx, actorCtxKey, actorID)
}

// ActorFromContext extracts the actor ID from the context.
func ActorFromContext(ctx context.Context) (string, error) {
	val, ok := ctx.Value(actorCtxKey).(string)
	if !ok || val == "" {
		return "", ErrActorNotFound
	}
	return val, nil
}

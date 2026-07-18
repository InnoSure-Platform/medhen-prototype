package middleware

import (
	"net/http"

	idempotency "github.com/medhen/pc-idempotency-mgmt-sdk"
)

// IdempotencyMiddleware ensures that requests with the same Idempotency-Key
// are not processed multiple times.
func IdempotencyMiddleware(manager *idempotency.Manager) func(http.Handler) http.Handler {
	return manager.Middleware
}

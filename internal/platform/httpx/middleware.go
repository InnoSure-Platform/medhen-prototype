// Package httpx provides HTTP edge primitives shared by the monolith: request
// correlation, panic recovery, and middleware chaining. It intentionally has no
// third-party dependencies so the composition root stays lightweight.
package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
)

type ctxKey string

// RequestIDKey is the context key under which the request correlation ID is stored.
const RequestIDKey ctxKey = "request_id"

// HeaderRequestID is the header used to read/emit the correlation ID.
const HeaderRequestID = "X-Request-Id"

// Middleware is the standard net/http middleware shape.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares to h in order, so the first listed runs outermost.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// RequestID ensures every request carries a correlation ID (honouring an
// inbound one) and echoes it on the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = newID()
		}
		w.Header().Set(HeaderRequestID, id)
		ctx := context.WithValue(r.Context(), RequestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the correlation ID, if present.
func RequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// Recover converts panics into 500s and logs them with the correlation ID,
// keeping one panicking request from taking down the process.
func Recover(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"request_id", RequestIDFromContext(r.Context()),
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-unknown"
	}
	return hex.EncodeToString(b)
}

// Package middleware provides auth-lite and recovery middleware for Phase 0.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	pcerr "github.com/InnoSure-Platform/pc-shared-go/errors"
	"github.com/InnoSure-Platform/pc-shared-go/httpx"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic", "recover", rec, "stack", string(debug.Stack()))
				httpx.WriteError(w, pcerr.E(pcerr.CodeInternal, "internal error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// DemoAuth accepts Bearer tokens or X-User-ID for the pilot.
// Production replaces this with Keycloak JWT validation at the gateway.
func DemoAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}
		// Phase 0: allow unauthenticated demo traffic when no Bearer token.
		if strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("X-User-ID") == "" {
			r.Header.Set("X-User-ID", "demo-agent")
		}
		if r.Header.Get("X-Tenant-ID") == "" {
			r.Header.Set("X-Tenant-ID", "eic")
		}
		next.ServeHTTP(w, r)
	})
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "request_id", r.Header.Get("X-Request-ID"))
		next.ServeHTTP(w, r)
	})
}

package auth

import (
	"net/http"
	"strings"
)

// Middleware returns chi-compatible JWT validation when Keycloak is configured.
func Middleware(v *KeycloakValidator) func(http.Handler) http.Handler {
	if v == nil || !v.Enabled() {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/api/v1/health" {
				next.ServeHTTP(w, r)
				return
			}
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Bearer ") {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"Bearer token required"}`, http.StatusUnauthorized)
				return
			}
			claims, err := v.Validate(r.Context(), strings.TrimPrefix(authz, "Bearer "))
			if err != nil {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-User-ID", claims.Sub)
			if claims.Email != "" {
				r.Header.Set("X-User-Email", claims.Email)
			}
			next.ServeHTTP(w, r)
		})
	}
}

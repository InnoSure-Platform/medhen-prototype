package auth

import (
	"context"
	"net/http"
	"strings"
)

// AccessRule grants access to a route prefix. AnyOf lists roles of which the
// principal must hold at least one; an empty AnyOf means "any authenticated
// principal". The longest matching prefix wins.
type AccessRule struct {
	Prefix string
	AnyOf  []string
}

// EdgeMiddleware enforces authentication and role-based access at the HTTP edge.
//
//   - Paths under any public prefix bypass authentication entirely (health
//     checks, HMAC-authenticated webhooks).
//   - When v is nil (auth disabled for local dev), requests pass through
//     unauthenticated — handlers then read the tenant from X-Tenant-ID.
//   - Otherwise a valid Bearer token is required (401 on missing/invalid); the
//     claims are placed on the context, and the longest-matching rule's roles are
//     enforced (403 when the principal lacks them).
func EdgeMiddleware(v *Validator, public []string, rules []AccessRule) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			if matchesAny(path, public) {
				next.ServeHTTP(w, r)
				return
			}
			if v == nil {
				next.ServeHTTP(w, r) // dev: auth disabled
				return
			}

			token, ok := bearer(r)
			if !ok {
				http.Error(w, "missing or malformed authorization header", http.StatusUnauthorized)
				return
			}
			claims, err := v.Validate(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, RolesKey, claims.EffectiveRoles())
			ctx = context.WithValue(ctx, ClaimsKey, claims)
			r = r.WithContext(ctx)

			if rule, ok := longestMatch(rules, path); ok && len(rule.AnyOf) > 0 && !hasAnyRole(ctx, rule.AnyOf) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func bearer(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func matchesAny(path string, prefixes []string) bool {
	for _, p := range prefixes {
		if path == p || strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func longestMatch(rules []AccessRule, path string) (AccessRule, bool) {
	best, found := AccessRule{}, false
	for _, rule := range rules {
		if strings.HasPrefix(path, rule.Prefix) && len(rule.Prefix) > len(best.Prefix) {
			best, found = rule, true
		}
	}
	return best, found
}

func hasAnyRole(ctx context.Context, roles []string) bool {
	for _, role := range roles {
		if HasRole(ctx, role) {
			return true
		}
	}
	return false
}

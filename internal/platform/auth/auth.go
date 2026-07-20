// Package auth is the platform authentication kernel. It validates Keycloak
// (OIDC) access tokens using RS256 signatures verified against the provider's
// JWKS endpoint, and fails closed: any request without a valid, signed,
// non-expired token for the configured issuer/audience is rejected with 401.
package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	RolesKey    contextKey = "roles"
	ClaimsKey   contextKey = "claims"
)

// JWTConfig configures the authentication middleware.
type JWTConfig struct {
	// IssuerURL is the expected `iss` claim, e.g.
	// http://localhost:8081/realms/medhen. Required.
	IssuerURL string
	// Audience, if non-empty, is required to appear in the token `aud` claim.
	Audience string
	// JWKSURL overrides the derived JWKS endpoint. When empty it defaults to
	// IssuerURL + "/protocol/openid-connect/certs" (Keycloak convention).
	JWKSURL string
	// HTTPClient is used to fetch the JWKS. Optional.
	HTTPClient *http.Client
	// JWKSRefreshInterval bounds how often an unknown-kid triggers a JWKS
	// refresh. Optional (defaults to 5m).
	JWKSRefreshInterval time.Duration

	// keyFunc, when set, overrides JWKS resolution. It exists solely for tests
	// that mint tokens with a known key; it still requires a valid signature,
	// so it is not an authentication bypass.
	keyFunc jwt.Keyfunc
}

// ConfigFromEnv builds a JWTConfig from the standard Keycloak environment
// variables (KEYCLOAK_URL, KEYCLOAK_REALM, KEYCLOAK_AUDIENCE).
func ConfigFromEnv() JWTConfig {
	base := strings.TrimRight(os.Getenv("KEYCLOAK_URL"), "/")
	realm := os.Getenv("KEYCLOAK_REALM")
	issuer := ""
	if base != "" && realm != "" {
		issuer = fmt.Sprintf("%s/realms/%s", base, realm)
	}
	return JWTConfig{
		IssuerURL: issuer,
		Audience:  os.Getenv("KEYCLOAK_AUDIENCE"),
	}
}

// CustomClaims are the claims extracted from a validated token.
type CustomClaims struct {
	TenantID    string      `json:"tenant_id,omitempty"`
	Roles       []string    `json:"roles,omitempty"`
	BranchCode  string      `json:"branch_code,omitempty"`
	RealmAccess realmAccess `json:"realm_access,omitempty"`
	jwt.RegisteredClaims
}

type realmAccess struct {
	Roles []string `json:"roles,omitempty"`
}

// EffectiveRoles merges custom `roles` and Keycloak realm roles.
func (c *CustomClaims) EffectiveRoles() []string {
	if len(c.Roles) > 0 {
		return c.Roles
	}
	return c.RealmAccess.Roles
}

// Validator verifies access tokens against a configured issuer and JWKS.
type Validator struct {
	cfg     JWTConfig
	keys    *keySet
	parser  *jwt.Parser
	keyFunc jwt.Keyfunc
}

// NewValidator constructs a Validator. It returns an error when the config is
// insufficient to validate tokens, so callers fail closed at startup rather
// than silently accepting unauthenticated traffic.
func NewValidator(cfg JWTConfig) (*Validator, error) {
	if cfg.keyFunc == nil && cfg.IssuerURL == "" {
		return nil, errors.New("auth: IssuerURL is required (set KEYCLOAK_URL and KEYCLOAK_REALM)")
	}

	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{"RS256"}),
		jwt.WithExpirationRequired(),
	}
	if cfg.IssuerURL != "" {
		opts = append(opts, jwt.WithIssuer(cfg.IssuerURL))
	}
	if cfg.Audience != "" {
		opts = append(opts, jwt.WithAudience(cfg.Audience))
	}

	v := &Validator{
		cfg:     cfg,
		parser:  jwt.NewParser(opts...),
		keyFunc: cfg.keyFunc,
	}

	if cfg.keyFunc == nil {
		jwksURL := cfg.JWKSURL
		if jwksURL == "" {
			jwksURL = strings.TrimRight(cfg.IssuerURL, "/") + "/protocol/openid-connect/certs"
		}
		v.keys = newKeySet(jwksURL, cfg.HTTPClient, cfg.JWKSRefreshInterval)
		v.keyFunc = v.jwksKeyFunc
	}
	return v, nil
}

// NewValidatorWithKeyFunc builds a Validator that resolves signing keys via the
// supplied jwt.Keyfunc instead of a JWKS endpoint. It is intended for tests
// that mint tokens with a locally generated key; it still enforces the
// signature, issuer, audience and expiry, so it is not an auth bypass.
func NewValidatorWithKeyFunc(cfg JWTConfig, keyFunc jwt.Keyfunc) (*Validator, error) {
	cfg.keyFunc = keyFunc
	return NewValidator(cfg)
}

// jwksKeyFunc resolves the RSA public key for a token from the JWKS cache,
// enforcing that the token is RS256-signed (defence against alg confusion).
func (v *Validator) jwksKeyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	kid, _ := token.Header["kid"].(string)
	return v.keys.keyByID(context.Background(), kid)
}

// Validate parses and verifies a raw bearer token string.
func (v *Validator) Validate(tokenString string) (*CustomClaims, error) {
	claims := &CustomClaims{}
	token, err := v.parser.ParseWithClaims(tokenString, claims, v.keyFunc)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("token invalid")
	}
	return claims, nil
}

// Handler is the http.Handler middleware form of the Validator.
func (v *Validator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		claims, err := v.Validate(parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, RolesKey, claims.EffectiveRoles())
		ctx = context.WithValue(ctx, ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns middleware that admits only principals holding role. It
// must be applied inside Handler (which populates the claims context).
func (v *Validator) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasRole(r.Context(), role) {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetTenantID retrieves the tenant ID from the authenticated context.
func GetTenantID(ctx context.Context) (string, error) {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok && tenantID != "" {
		return tenantID, nil
	}
	return "", errors.New("tenant ID not found in context (unauthenticated)")
}

// TenantOrHeader returns the authenticated tenant, falling back to the
// X-Tenant-ID header when no authenticated principal is present. The fallback is
// a development affordance for running without Keycloak; once auth is enforced at
// the edge (Phase 6), the authenticated tenant is authoritative and the header is
// ignored for authenticated requests.
func TenantOrHeader(r *http.Request) string {
	if tid, err := GetTenantID(r.Context()); err == nil {
		return tid
	}
	return r.Header.Get("X-Tenant-ID")
}

// GetClaims retrieves the full CustomClaims from context.
func GetClaims(ctx context.Context) (*CustomClaims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*CustomClaims)
	return claims, ok
}

// HasRole reports whether the authenticated principal holds the given role.
func HasRole(ctx context.Context, role string) bool {
	claims, ok := GetClaims(ctx)
	if !ok {
		return false
	}
	for _, r := range claims.EffectiveRoles() {
		if r == role {
			return true
		}
	}
	return false
}

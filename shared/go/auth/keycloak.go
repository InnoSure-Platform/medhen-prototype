package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// KeycloakValidator validates JWTs from a Keycloak realm (ADR-PC-015).
type KeycloakValidator struct {
	issuer   string
	audience string
	keys     map[string]*rsa.PublicKey
	mu       sync.RWMutex
	client   *http.Client
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// Claims extracted from a validated token.
type Claims struct {
	Sub   string
	Email string
	Roles []string
}

func NewKeycloakValidator(baseURL, realm, audience string) *KeycloakValidator {
	if audience == "" {
		audience = "pc-gateway"
	}
	return &KeycloakValidator{
		issuer:   strings.TrimRight(baseURL, "/") + "/realms/" + realm,
		audience: audience,
		keys:     map[string]*rsa.PublicKey{},
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (v *KeycloakValidator) Enabled() bool { return v != nil && v.issuer != "" }

func (v *KeycloakValidator) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	if err := v.refreshKeys(ctx); err != nil {
		return nil, fmt.Errorf("jwks: %w", err)
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected alg %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		v.mu.RLock()
		key := v.keys[kid]
		v.mu.RUnlock()
		if key == nil {
			return nil, fmt.Errorf("unknown kid %s", kid)
		}
		return key, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	m, _ := token.Claims.(jwt.MapClaims)
	c := &Claims{Sub: str(m["sub"]), Email: str(m["email"])}
	if ra, ok := m["realm_access"].(map[string]any); ok {
		if roles, ok := ra["roles"].([]any); ok {
			for _, r := range roles {
				if s, ok := r.(string); ok {
					c.Roles = append(c.Roles, s)
				}
			}
		}
	}
	if c.Sub == "" {
		return nil, fmt.Errorf("missing sub")
	}
	return c, nil
}

func (v *KeycloakValidator) refreshKeys(ctx context.Context) error {
	v.mu.RLock()
	if len(v.keys) > 0 {
		v.mu.RUnlock()
		return nil
	}
	v.mu.RUnlock()
	url := v.issuer + "/protocol/openid-connect/certs"
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	res, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var doc jwks
	if err := json.NewDecoder(res.Body).Decode(&doc); err != nil {
		return err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	for _, k := range doc.Keys {
		if k.Kty != "RSA" {
			continue
		}
		pub, err := parseRSAPublic(k.N, k.E)
		if err != nil {
			continue
		}
		v.keys[k.Kid] = pub
	}
	return nil
}

func parseRSAPublic(nB64, eB64 string) (*rsa.PublicKey, error) {
	nb, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nb)
	var ei int
	for _, b := range eb {
		ei = ei<<8 + int(b)
	}
	return &rsa.PublicKey{N: n, E: ei}, nil
}

func str(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// Middleware returns HTTP middleware that validates Bearer tokens when Keycloak is configured.
func MiddlewareLegacy(v *KeycloakValidator, next http.Handler) http.Handler {
	return Middleware(v)(next)
}

// NewValidatorFromEnv returns nil if KEYCLOAK_URL unset (demo mode).
func NewValidatorFromEnv() *KeycloakValidator {
	base := getenv("KEYCLOAK_URL")
	if base == "" {
		return nil
	}
	realm := getenv("KEYCLOAK_REALM")
	if realm == "" {
		realm = "medhen"
	}
	return NewKeycloakValidator(base, realm, getenv("KEYCLOAK_AUDIENCE"))
}

func getenv(k string) string {
	// filled by env.go
	return envGet(k)
}

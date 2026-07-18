package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"
)

// jwksResponse mirrors the JSON Web Key Set returned by Keycloak's
// /protocol/openid-connect/certs endpoint.
type jwksResponse struct {
	Keys []jsonWebKey `json:"keys"`
}

type jsonWebKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"` // modulus (base64url)
	E   string `json:"e"` // exponent (base64url)
}

// keySet caches RSA public keys fetched from a JWKS endpoint, keyed by `kid`.
// It refreshes on demand (when an unknown kid is seen) with a minimum interval
// to avoid hammering the identity provider under a key-confusion attack.
type keySet struct {
	url         string
	httpClient  *http.Client
	minRefresh  time.Duration
	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastRefresh time.Time
}

func newKeySet(url string, httpClient *http.Client, minRefresh time.Duration) *keySet {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	if minRefresh <= 0 {
		minRefresh = 5 * time.Minute
	}
	return &keySet{
		url:        url,
		httpClient: httpClient,
		minRefresh: minRefresh,
		keys:       make(map[string]*rsa.PublicKey),
	}
}

// keyByID returns the RSA public key for the given kid, refreshing the cache
// from the JWKS endpoint if the kid is unknown (bounded by minRefresh).
func (ks *keySet) keyByID(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, fmt.Errorf("token missing kid header")
	}

	ks.mu.RLock()
	key, ok := ks.keys[kid]
	ks.mu.RUnlock()
	if ok {
		return key, nil
	}

	if err := ks.refresh(ctx); err != nil {
		return nil, err
	}

	ks.mu.RLock()
	key, ok = ks.keys[kid]
	ks.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no JWKS key matches kid %q", kid)
	}
	return key, nil
}

func (ks *keySet) refresh(ctx context.Context) error {
	ks.mu.Lock()
	defer ks.mu.Unlock()

	if !ks.lastRefresh.IsZero() && time.Since(ks.lastRefresh) < ks.minRefresh {
		return fmt.Errorf("JWKS refresh rate-limited; unknown kid")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ks.url, nil)
	if err != nil {
		return fmt.Errorf("build JWKS request: %w", err)
	}
	resp, err := ks.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS endpoint returned %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	next := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" {
			continue
		}
		if k.Use != "" && k.Use != "sig" {
			continue
		}
		pub, err := k.toRSAPublicKey()
		if err != nil {
			continue
		}
		next[k.Kid] = pub
	}
	if len(next) == 0 {
		return fmt.Errorf("JWKS contained no usable RSA signing keys")
	}

	ks.keys = next
	ks.lastRefresh = time.Now()
	return nil
}

func (k jsonWebKey) toRSAPublicKey() (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() || e.Int64() > int64(^uint32(0)) {
		return nil, fmt.Errorf("exponent too large")
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(e.Int64()),
	}, nil
}

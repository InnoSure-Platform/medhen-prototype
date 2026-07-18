package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testIssuer   = "https://test.local/realms/medhen"
	testAudience = "pc-gateway"
)

func testValidator(t *testing.T, key *rsa.PrivateKey) *Validator {
	t.Helper()
	v, err := NewValidatorWithKeyFunc(
		JWTConfig{IssuerURL: testIssuer, Audience: testAudience},
		func(*jwt.Token) (interface{}, error) { return &key.PublicKey, nil },
	)
	if err != nil {
		t.Fatalf("NewValidatorWithKeyFunc: %v", err)
	}
	return v
}

func signRS256(t *testing.T, key *rsa.PrivateKey, claims CustomClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = "test-key"
	s, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func validClaims() CustomClaims {
	return CustomClaims{
		TenantID: "eic",
		Roles:    []string{"agent"},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    testIssuer,
			Audience:  jwt.ClaimStrings{testAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		},
	}
}

func TestValidate_ValidToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	claims, err := v.Validate(signRS256(t, key, validClaims()))
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.TenantID != "eic" {
		t.Fatalf("tenant not extracted, got %q", claims.TenantID)
	}
}

func TestValidate_RejectsAlgConfusion(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
	hs, err := tok.SignedString([]byte("attacker-secret"))
	if err != nil {
		t.Fatalf("sign hs256: %v", err)
	}
	if _, err := v.Validate(hs); err == nil {
		t.Fatal("expected HS256 token to be rejected (alg confusion)")
	}
}

func TestValidate_RejectsExpired(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	c := validClaims()
	c.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))
	if _, err := v.Validate(signRS256(t, key, c)); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestValidate_RejectsWrongIssuer(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	c := validClaims()
	c.Issuer = "https://evil.example/realms/medhen"
	if _, err := v.Validate(signRS256(t, key, c)); err == nil {
		t.Fatal("expected wrong-issuer token to be rejected")
	}
}

func TestValidate_RejectsWrongAudience(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	c := validClaims()
	c.Audience = jwt.ClaimStrings{"some-other-client"}
	if _, err := v.Validate(signRS256(t, key, c)); err == nil {
		t.Fatal("expected wrong-audience token to be rejected")
	}
}

func TestValidate_RejectsWrongKey(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	other, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	if _, err := v.Validate(signRS256(t, other, validClaims())); err == nil {
		t.Fatal("expected token signed by untrusted key to be rejected")
	}
}

func TestHandler_FailsClosedWithoutHeader(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	h := v.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without Authorization header, got %d", rec.Code)
	}
}

func TestHandler_RejectsXTenantIDHeader(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	h := v.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "eic")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for X-Tenant-ID-only request, got %d", rec.Code)
	}
}

func TestHandler_AcceptsValidTokenAndPropagatesTenant(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)
	var gotTenant string
	h := v.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant, _ = GetTenantID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signRS256(t, key, validClaims()))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid token, got %d", rec.Code)
	}
	if gotTenant != "eic" {
		t.Fatalf("tenant not propagated, got %q", gotTenant)
	}
}

func TestRequireRole(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	v := testValidator(t, key)

	protected := v.Handler(v.RequireRole("admin")(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })))

	// agent role → forbidden
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signRS256(t, key, validClaims()))
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", rec.Code)
	}

	// admin role → ok
	adminClaims := validClaims()
	adminClaims.Roles = []string{"admin"}
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("Authorization", "Bearer "+signRS256(t, key, adminClaims))
	rec2 := httptest.NewRecorder()
	protected.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", rec2.Code)
	}
}

func TestNewValidator_FailsClosedOnMissingIssuer(t *testing.T) {
	if _, err := NewValidator(JWTConfig{}); err == nil {
		t.Fatal("expected error when IssuerURL is unset (must fail closed)")
	}
}

package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func claimsWithRoles(roles ...string) CustomClaims {
	c := validClaims()
	c.Roles = roles
	return c
}

func edgeRules() []AccessRule {
	return []AccessRule{
		{Prefix: "/iam/", AnyOf: []string{"admin"}},
		{Prefix: "/claims/", AnyOf: []string{"claims", "admin"}},
	}
}

func doEdge(t *testing.T, mw func(http.Handler) http.Handler, method, path, token string) int {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	mw(okHandler()).ServeHTTP(rec, req)
	return rec.Code
}

func TestEdge_PublicBypassesAuth(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	mw := EdgeMiddleware(testValidator(t, key), []string{"/healthz", "/billing/webhooks/telebirr"}, edgeRules())

	if code := doEdge(t, mw, "GET", "/healthz", ""); code != http.StatusOK {
		t.Fatalf("public /healthz = %d, want 200", code)
	}
	if code := doEdge(t, mw, "POST", "/billing/webhooks/telebirr", ""); code != http.StatusOK {
		t.Fatalf("public webhook = %d, want 200", code)
	}
}

func TestEdge_DevModePassesThrough(t *testing.T) {
	mw := EdgeMiddleware(nil, nil, edgeRules())
	if code := doEdge(t, mw, "GET", "/policy/quotes/x", ""); code != http.StatusOK {
		t.Fatalf("dev mode (nil validator) = %d, want 200", code)
	}
}

func TestEdge_RequiresToken(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	mw := EdgeMiddleware(testValidator(t, key), []string{"/healthz"}, edgeRules())
	if code := doEdge(t, mw, "GET", "/party/parties/x", ""); code != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", code)
	}
	if code := doEdge(t, mw, "GET", "/party/parties/x", "garbage"); code != http.StatusUnauthorized {
		t.Fatalf("bad token = %d, want 401", code)
	}
}

func TestEdge_AuthenticatedNoRuleAllowed(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	mw := EdgeMiddleware(testValidator(t, key), []string{"/healthz"}, edgeRules())
	// /product/ has no rule → any authenticated principal is allowed.
	token := signRS256(t, key, claimsWithRoles("agent"))
	if code := doEdge(t, mw, "GET", "/product/products", token); code != http.StatusOK {
		t.Fatalf("authenticated + no rule = %d, want 200", code)
	}
}

func TestEdge_RBACForbidsMissingRole(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	mw := EdgeMiddleware(testValidator(t, key), []string{"/healthz"}, edgeRules())

	agent := signRS256(t, key, claimsWithRoles("agent"))
	if code := doEdge(t, mw, "POST", "/claims/claims", agent); code != http.StatusForbidden {
		t.Fatalf("agent on /claims = %d, want 403", code)
	}
	if code := doEdge(t, mw, "POST", "/iam/users", agent); code != http.StatusForbidden {
		t.Fatalf("agent on /iam = %d, want 403", code)
	}
}

func TestEdge_RBACAdmitsMatchingRole(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	mw := EdgeMiddleware(testValidator(t, key), []string{"/healthz"}, edgeRules())

	if code := doEdge(t, mw, "POST", "/claims/claims", signRS256(t, key, claimsWithRoles("claims"))); code != http.StatusOK {
		t.Fatalf("claims role on /claims = %d, want 200", code)
	}
	if code := doEdge(t, mw, "POST", "/iam/users", signRS256(t, key, claimsWithRoles("admin"))); code != http.StatusOK {
		t.Fatalf("admin on /iam = %d, want 200", code)
	}
}

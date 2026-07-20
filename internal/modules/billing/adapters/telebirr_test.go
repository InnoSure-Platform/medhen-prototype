package adapters_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/billing/adapters"
)

func sign(secret string, body []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(body)
	return hex.EncodeToString(m.Sum(nil))
}

func TestVerifyTelebirrSignature(t *testing.T) {
	secret := "telebirr-sandbox-secret"
	body := []byte(`{"invoice_id":"inv-1","amount_minor":268000,"status":"SUCCESS"}`)

	if !adapters.VerifyTelebirrSignature(secret, body, sign(secret, body)) {
		t.Fatal("valid signature must verify")
	}
	if adapters.VerifyTelebirrSignature(secret, body, sign("wrong-secret", body)) {
		t.Fatal("signature under wrong secret must be rejected")
	}
	if adapters.VerifyTelebirrSignature(secret, append(body, '!'), sign(secret, body)) {
		t.Fatal("tampered body must be rejected")
	}
	// The pre-refactor bug: a non-empty but bogus signature was accepted.
	if adapters.VerifyTelebirrSignature(secret, body, "deadbeef") {
		t.Fatal("arbitrary non-empty signature must be rejected")
	}
	if adapters.VerifyTelebirrSignature("", body, "anything") {
		t.Fatal("empty secret must fail closed")
	}
	if adapters.VerifyTelebirrSignature(secret, body, "") {
		t.Fatal("empty signature must fail closed")
	}
}

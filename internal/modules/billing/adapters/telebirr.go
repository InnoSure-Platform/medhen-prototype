package adapters

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifyTelebirrSignature checks an HMAC-SHA256 signature over the raw webhook
// body using the shared secret. This fixes the pre-refactor webhook, which
// accepted any request whose signature header was merely non-empty. Empty secret
// or signature fails closed; comparison is constant-time.
func VerifyTelebirrSignature(secret string, body []byte, signatureHex string) bool {
	if secret == "" || signatureHex == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signatureHex))
}

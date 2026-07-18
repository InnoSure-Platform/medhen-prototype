package extractor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type CryptoSealer struct {
	secretKey []byte
}

func NewCryptoSealer(secretKey string) *CryptoSealer {
	return &CryptoSealer{secretKey: []byte(secretKey)}
}

// Seal payload with HMAC-SHA256
func (s *CryptoSealer) Seal(payload []byte) (string, error) {
	h := hmac.New(sha256.New, s.secretKey)
	_, err := h.Write(payload)
	if err != nil {
		return "", fmt.Errorf("failed to generate hmac: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

type SealedDocument struct {
	Payload []byte
	Hash    string
}

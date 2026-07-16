package postgres

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type KMS interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertext string) ([]byte, error)
}

// MockKMS implements AES-256 encryption using an injected 32-byte key.
type MockKMS struct {
	key []byte
}

func NewMockKMS(key string) (*MockKMS, error) {
	if len(key) != 32 {
		return nil, errors.New("invalid key length: must be 32 bytes for AES-256")
	}
	return &MockKMS{key: []byte(key)}, nil
}

func (k *MockKMS) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(k.key)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (k *MockKMS) Decrypt(cryptoText string) ([]byte, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(k.key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext) // In-place decryption

	return ciphertext, nil
}

package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// AESTokenCipher encrypts tokens at rest with AES-256-GCM as base64(nonce || ciphertext).
type AESTokenCipher struct {
	aead cipher.AEAD
}

// NewAESTokenCipher takes a base64 encoded 32 byte key.
func NewAESTokenCipher(keyBase64 string) (*AESTokenCipher, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode token encryption key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("token encryption key must have 32 bytes, got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm: %w", err)
	}
	return &AESTokenCipher{aead: aead}, nil
}

func (a *AESTokenCipher) Encrypt(plain string) (string, error) {
	nonce := make([]byte, a.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	sealed := a.aead.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (a *AESTokenCipher) Decrypt(encrypted string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil || len(raw) < a.aead.NonceSize() {
		return "", ErrInvalidCiphertext
	}

	nonce, sealed := raw[:a.aead.NonceSize()], raw[a.aead.NonceSize():]
	plain, err := a.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	return string(plain), nil
}

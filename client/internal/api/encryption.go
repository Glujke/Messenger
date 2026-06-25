package api

import (
	"encoding/base64"
	"errors"
)

var ErrEmptyEncryptionKey = errors.New("encryption key must not be empty")

// MessageEncrypter encrypts and decrypts message bodies.
type MessageEncrypter interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// XOREncrypter is a toy encrypter for development and tests only.
type XOREncrypter struct {
	key []byte
}

// NewXOREncrypter creates a toy XOR encrypter.
func NewXOREncrypter(key string) *XOREncrypter {
	return &XOREncrypter{key: []byte(key)}
}

// Encrypt XOR-encodes plaintext and returns base64 ciphertext.
func (e *XOREncrypter) Encrypt(plaintext string) (string, error) {
	if len(e.key) == 0 {
		return "", ErrEmptyEncryptionKey
	}

	data := []byte(plaintext)
	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ e.key[i%len(e.key)]
	}
	return base64.StdEncoding.EncodeToString(out), nil
}

// Decrypt decodes base64 ciphertext and XOR-decodes it back to plaintext.
func (e *XOREncrypter) Decrypt(ciphertext string) (string, error) {
	if len(e.key) == 0 {
		return "", ErrEmptyEncryptionKey
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	out := make([]byte, len(data))
	for i := range data {
		out[i] = data[i] ^ e.key[i%len(e.key)]
	}
	return string(out), nil
}

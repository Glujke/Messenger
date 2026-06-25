package domain

import "testing"

func TestXOREncrypter_RoundTrip(t *testing.T) {
	enc := NewXOREncrypter("secret-key")

	ciphertext, err := enc.Encrypt("hello")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if ciphertext == "hello" {
		t.Fatal("ciphertext must not equal plaintext")
	}

	plaintext, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "hello" {
		t.Fatalf("plaintext = %q, want %q", plaintext, "hello")
	}
}

func TestXOREncrypter_EmptyString(t *testing.T) {
	enc := NewXOREncrypter("secret-key")

	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	plaintext, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "" {
		t.Fatalf("plaintext = %q, want empty string", plaintext)
	}
}

func TestXOREncrypter_Unicode(t *testing.T) {
	enc := NewXOREncrypter("secret-key")
	input := "Привет, Сибирь! 🐻"

	ciphertext, err := enc.Encrypt(input)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	plaintext, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != input {
		t.Fatalf("plaintext = %q, want %q", plaintext, input)
	}
}

func TestXOREncrypter_EmptyKey(t *testing.T) {
	enc := NewXOREncrypter("")

	_, err := enc.Encrypt("hello")
	if err != ErrEmptyEncryptionKey {
		t.Fatalf("Encrypt() error = %v, want %v", err, ErrEmptyEncryptionKey)
	}
}

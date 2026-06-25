package config

import (
	"testing"
)

func TestLoad_EncryptionKeyFromEnv(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "my-secret-key")

	cfg := Load()
	if cfg.EncryptionKey != "my-secret-key" {
		t.Fatalf("EncryptionKey = %q, want %q", cfg.EncryptionKey, "my-secret-key")
	}
}

func TestLoad_EncryptionKeyDefault(t *testing.T) {
	t.Setenv("ENCRYPTION_KEY", "")

	cfg := Load()
	if cfg.EncryptionKey != "dev-encryption-key-change-me" {
		t.Fatalf("EncryptionKey = %q, want default", cfg.EncryptionKey)
	}
}

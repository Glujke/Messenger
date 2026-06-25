package config

import "os"

// Config holds client-side settings.
type Config struct {
	ServerURL     string
	EncryptionKey string
}

// Default returns the initial client configuration.
func Default() Config {
	return Config{
		ServerURL:     env("SERVER_URL", "http://localhost:8080"),
		EncryptionKey: env("ENCRYPTION_KEY", "dev-encryption-key-change-me"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

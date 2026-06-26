package config

import "os"

// Config holds client-side settings.
type Config struct {
	ServerURL     string
	AppName       string
	EncryptionKey string
}

// Default returns the initial client configuration.
func Default() Config {
	return Config{
		ServerURL:     env("SERVER_URL", "http://itc05:8080"),
		AppName:       env("APP_NAME", "my_messenger"),
		EncryptionKey: env("ENCRYPTION_KEY", "dev-encryption-key-change-me"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

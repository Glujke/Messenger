package config

import "os"

// Config holds application configuration loaded from environment variables.
type Config struct {
	HTTPPort    string
	DatabaseURL string
}

// Load reads configuration from the environment with sensible defaults for local development.
func Load() Config {
	return Config{
		HTTPPort:    env("HTTP_PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://messenger:messenger@localhost:5432/messenger?sslmode=disable"),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

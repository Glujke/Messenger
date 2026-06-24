package config

import (
	"os"
	"time"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	HTTPPort    string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

// Load reads configuration from the environment with sensible defaults for local development.
func Load() Config {
	return Config{
		HTTPPort:    env("HTTP_PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://messenger:messenger@localhost:5432/messenger?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "dev-secret-change-me"),
		JWTTTL:      envDuration("JWT_TTL", 24*time.Hour),
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

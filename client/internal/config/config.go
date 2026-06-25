package config

// Config holds client-side settings.
type Config struct {
	ServerURL string
}

// Default returns the initial client configuration.
func Default() Config {
	return Config{
		ServerURL: "http://ITC05:8080",
	}
}

package handler

import (
	"log/slog"
	"net/http"

	"messenger/backend/internal/repository"
)

// NewRouter registers HTTP routes and middleware.
func NewRouter(logger *slog.Logger, store repository.Store, auth *AuthHandler) http.Handler {
	mux := http.NewServeMux()

	health := NewHealthHandler(store)
	mux.Handle("GET /health", health)
	mux.Handle("GET /health/live", health)
	mux.Handle("GET /health/ready", health)

	mux.Handle("POST /auth/register", auth)
	mux.Handle("POST /auth/login", auth)

	return loggingMiddleware(logger, mux)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

package handler

import (
	"log/slog"
	"net/http"

	"messenger/backend/internal/repository"
)

// NewRouter registers HTTP routes and middleware.
func NewRouter(
	logger *slog.Logger,
	store repository.Store,
	auth *AuthHandler,
	me *MeHandler,
	rooms *RoomsHandler,
	messages *MessagesHandler,
	attachments *AttachmentsHandler,
	ws *WSHandler,
	jwtSecret string,
) http.Handler {
	mux := http.NewServeMux()

	health := NewHealthHandler(store)
	mux.Handle("GET /health", health)
	mux.Handle("GET /health/live", health)
	mux.Handle("GET /health/ready", health)

	mux.Handle("POST /auth/register", auth)
	mux.Handle("POST /auth/login", auth)
	mux.Handle("GET /auth/me", RequireAuth(jwtSecret, me))
	mux.Handle("POST /rooms/direct", RequireAuth(jwtSecret, rooms))
	mux.Handle("GET /rooms", RequireAuth(jwtSecret, rooms))
	mux.Handle("POST /rooms/{id}/messages", RequireAuth(jwtSecret, messages))
	mux.Handle("GET /rooms/{id}/messages", RequireAuth(jwtSecret, messages))
	mux.Handle("POST /rooms/{id}/attachments", RequireAuth(jwtSecret, http.HandlerFunc(attachments.ServeUpload)))
	mux.Handle("GET /attachments/{id}", RequireAuth(jwtSecret, http.HandlerFunc(attachments.ServeDownload)))
	mux.Handle("GET /ws", ws)

	return loggingMiddleware(logger, mux)
}

func loggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"messenger/backend/internal/repository"
)

// HealthHandler serves liveness checks.
type HealthHandler struct {
	store repository.Store
}

// NewHealthHandler creates a health check handler.
func NewHealthHandler(store repository.Store) *HealthHandler {
	return &HealthHandler{store: store}
}

// ServeHTTP implements http.Handler.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/health/live":
		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	case "/health/ready":
		if err := h.store.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status": "degraded",
				"error":  err.Error(),
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	default:
		h.serveLegacyHealth(w, r.Context())
	}
}

func (h *HealthHandler) serveLegacyHealth(w http.ResponseWriter, ctx context.Context) {
	status := http.StatusOK
	body := map[string]string{
		"status": "ok",
	}

	if err := h.store.Ping(ctx); err != nil {
		status = http.StatusServiceUnavailable
		body["status"] = "degraded"
		body["database"] = "unreachable"
	} else {
		body["database"] = "ok"
	}

	writeJSON(w, status, body)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

package handler

import "net/http"

// MeHandler returns the authenticated user's profile.
type MeHandler struct{}

// NewMeHandler creates a profile handler.
func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

type meResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

// ServeHTTP implements http.Handler.
func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/auth/me" {
		http.NotFound(w, r)
		return
	}

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}

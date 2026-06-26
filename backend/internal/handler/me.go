package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
	"messenger/backend/internal/service"
)

// ProfileUpdater updates the authenticated user's profile.
type ProfileUpdater interface {
	UpdateUsername(ctx context.Context, userID int64, username string) (service.ProfileResult, error)
}

// PasswordChanger changes the authenticated user's password.
type PasswordChanger interface {
	ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error
}

// MeHandler serves profile read and username update endpoints.
type MeHandler struct {
	profile ProfileUpdater
}

// NewMeHandler creates a profile handler.
func NewMeHandler(profile ProfileUpdater) *MeHandler {
	return &MeHandler{profile: profile}
}

type meResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type updateMeRequest struct {
	Username string `json:"username"`
}

// ServeHTTP implements http.Handler.
func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/auth/me" {
		http.NotFound(w, r)
		return
	}

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, meResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		})
	case http.MethodPatch:
		h.serveUpdate(w, r, user.ID)
	default:
		http.NotFound(w, r)
	}
}

func (h *MeHandler) serveUpdate(w http.ResponseWriter, r *http.Request, userID int64) {
	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	result, err := h.profile.UpdateUsername(r.Context(), userID, req.Username)
	if errors.Is(err, domain.ErrInvalidUsername) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if errors.Is(err, repository.ErrUsernameTaken) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "username already taken"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, meResponse{
		ID:       result.ID,
		Email:    result.Email,
		Username: result.Username,
	})
}

// PasswordHandler serves password change requests.
type PasswordHandler struct {
	profile PasswordChanger
}

// NewPasswordHandler creates a password change handler.
func NewPasswordHandler(profile PasswordChanger) *PasswordHandler {
	return &PasswordHandler{profile: profile}
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ServeHTTP implements http.Handler.
func (h *PasswordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || r.URL.Path != "/auth/password" {
		http.NotFound(w, r)
		return
	}

	user, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	err := h.profile.ChangePassword(r.Context(), user.ID, req.OldPassword, req.NewPassword)
	if errors.Is(err, service.ErrWrongPassword) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid current password"})
		return
	}
	if errors.Is(err, domain.ErrInvalidPassword) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

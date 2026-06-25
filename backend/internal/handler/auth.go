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

// AuthRegistrar registers new users.
type AuthRegistrar interface {
	Register(ctx context.Context, email, username, password string) (service.RegisterResult, error)
}

// AuthAuthenticator logs users in.
type AuthAuthenticator interface {
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthHandler serves registration and login endpoints.
type AuthHandler struct {
	registrar AuthRegistrar
	login     AuthAuthenticator
}

// NewAuthHandler creates an auth HTTP handler.
func NewAuthHandler(registrar AuthRegistrar, login AuthAuthenticator) *AuthHandler {
	return &AuthHandler{
		registrar: registrar,
		login:     login,
	}
}

type authRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// ServeHTTP implements http.Handler.
func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/auth/register":
		h.serveRegister(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/auth/login":
		h.serveLogin(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AuthHandler) serveRegister(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	result, err := h.registrar.Register(r.Context(), req.Email, req.Username, req.Password)
	if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrInvalidUsername) || errors.Is(err, domain.ErrInvalidPassword) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if errors.Is(err, repository.ErrEmailTaken) {
		writeJSON(w, http.StatusConflict, errorResponse{Error: "email already registered"})
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

	writeJSON(w, http.StatusCreated, registerResponse{
		ID:       result.ID,
		Email:    result.Email,
		Username: result.Username,
	})
}

func (h *AuthHandler) serveLogin(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	token, err := h.login.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}
	if errors.Is(err, service.ErrNotVerified) {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "account not verified"})
		return
	}
	if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrInvalidPassword) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: token})
}

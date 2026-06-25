package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
	"messenger/backend/internal/service"
)

type mockAuthRegistrar struct {
	result service.RegisterResult
	err    error
}

func (m *mockAuthRegistrar) Register(_ context.Context, email, username, password string) (service.RegisterResult, error) {
	if m.err != nil {
		return service.RegisterResult{}, m.err
	}
	if m.result.Email == "" {
		m.result.Email = email
	}
	if m.result.Username == "" {
		m.result.Username = username
	}
	return m.result, nil
}

type mockAuthAuthenticator struct {
	token string
	err   error
}

func (m *mockAuthAuthenticator) Login(context.Context, string, string) (string, error) {
	return m.token, m.err
}

func TestAuthHandler_Register(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{result: service.RegisterResult{ID: 1, Email: "user@example.com", Username: "user"}},
		&mockAuthAuthenticator{},
	)

	body := `{"email":"user@example.com","username":"user","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp registerResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != 1 || resp.Email != "user@example.com" || resp.Username != "user" {
		t.Fatalf("response = %+v", resp)
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(&mockAuthRegistrar{}, &mockAuthAuthenticator{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(`{bad`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{err: domain.ErrInvalidEmail},
		&mockAuthAuthenticator{},
	)

	body := `{"email":"bad","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Register_EmailTaken(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{err: repository.ErrEmailTaken},
		&mockAuthAuthenticator{},
	)

	body := `{"email":"user@example.com","username":"user","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestAuthHandler_Register_UsernameTaken(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{err: repository.ErrUsernameTaken},
		&mockAuthAuthenticator{},
	)

	body := `{"email":"user@example.com","username":"user","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestAuthHandler_Login(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{},
		&mockAuthAuthenticator{token: "jwt-token"},
	)

	body := `{"email":"user@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp loginResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Token != "jwt-token" {
		t.Fatalf("token = %q, want jwt-token", resp.Token)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{},
		&mockAuthAuthenticator{err: service.ErrInvalidCredentials},
	)

	body := `{"email":"user@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Login_NotVerified(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{},
		&mockAuthAuthenticator{err: service.ErrNotVerified},
	)

	body := `{"email":"user@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	h := NewAuthHandler(&mockAuthRegistrar{}, &mockAuthAuthenticator{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{bad`))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_UnknownRoute(t *testing.T) {
	h := NewAuthHandler(&mockAuthRegistrar{}, &mockAuthAuthenticator{})

	req := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAuthHandler_Login_InternalError(t *testing.T) {
	h := NewAuthHandler(
		&mockAuthRegistrar{},
		&mockAuthAuthenticator{err: errors.New("db down")},
	)

	body := `{"email":"user@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMeHandler_Authenticated(t *testing.T) {
	h := NewMeHandler()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp meResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != 1 || resp.Email != "test@local" {
		t.Fatalf("response = %+v", resp)
	}
}

func TestMeHandler_Unauthorized(t *testing.T) {
	h := NewMeHandler()

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMeHandler_WrongMethod(t *testing.T) {
	h := NewMeHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/me", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

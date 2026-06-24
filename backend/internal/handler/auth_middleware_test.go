package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"messenger/backend/internal/service"
)

func TestRequireAuth_ValidToken(t *testing.T) {
	const secret = "test-secret"
	token, err := service.IssueToken(7, "user@example.com", secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	var gotUser AuthUser
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := AuthUserFromContext(r.Context())
		if !ok {
			t.Fatal("auth user missing from context")
		}
		gotUser = user
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	RequireAuth(secret, next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUser.ID != 7 || gotUser.Email != "user@example.com" {
		t.Fatalf("user = %+v", gotUser)
	}
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	RequireAuth("secret", next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler must not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer bad-token")
	rec := httptest.NewRecorder()

	RequireAuth("secret", next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_EmptyBearer(t *testing.T) {
	next := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler must not be called")
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	RequireAuth("secret", next).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ContextPreserved(t *testing.T) {
	const secret = "test-secret"
	token, err := service.IssueToken(1, "user@example.com", secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	type ctxKey struct{}
	baseCtx := context.WithValue(context.Background(), ctxKey{}, "kept")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(ctxKey{}) != "kept" {
			t.Fatal("original context value was lost")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(baseCtx)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	RequireAuth(secret, next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

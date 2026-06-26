package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/service"
)

type mockProfileService struct {
	updateUsernameFn func(ctx context.Context, userID int64, username string) (service.ProfileResult, error)
	changePasswordFn func(ctx context.Context, userID int64, oldPassword, newPassword string) error
}

func (m *mockProfileService) UpdateUsername(ctx context.Context, userID int64, username string) (service.ProfileResult, error) {
	return m.updateUsernameFn(ctx, userID, username)
}

func (m *mockProfileService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	return m.changePasswordFn(ctx, userID, oldPassword, newPassword)
}

func TestMeHandler_Authenticated(t *testing.T) {
	h := NewMeHandler(&mockProfileService{})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local", Username: "testuser"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp meResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != 1 || resp.Email != "test@local" || resp.Username != "testuser" {
		t.Fatalf("response = %+v", resp)
	}
}

func TestMeHandler_Unauthorized(t *testing.T) {
	h := NewMeHandler(&mockProfileService{})

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMeHandler_UpdateUsername(t *testing.T) {
	h := NewMeHandler(&mockProfileService{
		updateUsernameFn: func(_ context.Context, userID int64, username string) (service.ProfileResult, error) {
			if userID != 1 || username != "newname" {
				t.Fatalf("userID=%d username=%q", userID, username)
			}
			return service.ProfileResult{ID: 1, Email: "test@local", Username: "newname"}, nil
		},
	})

	body := `{"username":"newname"}`
	req := httptest.NewRequest(http.MethodPatch, "/auth/me", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local", Username: "testuser"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestMeHandler_UpdateUsername_Invalid(t *testing.T) {
	h := NewMeHandler(&mockProfileService{
		updateUsernameFn: func(context.Context, int64, string) (service.ProfileResult, error) {
			return service.ProfileResult{}, domain.ErrInvalidUsername
		},
	})

	req := httptest.NewRequest(http.MethodPatch, "/auth/me", bytes.NewBufferString(`{"username":"ab"}`))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local", Username: "testuser"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPasswordHandler_Success(t *testing.T) {
	h := NewPasswordHandler(&mockProfileService{
		changePasswordFn: func(_ context.Context, userID int64, oldPassword, newPassword string) error {
			if userID != 1 || oldPassword != "oldsecret1" || newPassword != "newsecret1" {
				t.Fatalf("unexpected args: %d %q %q", userID, oldPassword, newPassword)
			}
			return nil
		},
	})

	body := `{"old_password":"oldsecret1","new_password":"newsecret1"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/password", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local", Username: "testuser"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestPasswordHandler_WrongOldPassword(t *testing.T) {
	h := NewPasswordHandler(&mockProfileService{
		changePasswordFn: func(context.Context, int64, string, string) error {
			return service.ErrWrongPassword
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/auth/password", bytes.NewBufferString(`{"old_password":"bad","new_password":"newsecret1"}`))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "test@local", Username: "testuser"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

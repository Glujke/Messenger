package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"messenger/backend/internal/service"
)

type mockContactManager struct {
	listContactsFn func(ctx context.Context, userID int64) ([]service.ContactItem, error)
}

func (m *mockContactManager) Invite(context.Context, int64, string) (service.ContactRequestItem, error) {
	panic("not implemented")
}
func (m *mockContactManager) AcceptRequest(context.Context, int64, int64) error {
	panic("not implemented")
}
func (m *mockContactManager) RejectRequest(context.Context, int64, int64) error {
	panic("not implemented")
}
func (m *mockContactManager) ListRequests(context.Context, int64) ([]service.ContactRequestItem, error) {
	panic("not implemented")
}
func (m *mockContactManager) ListContacts(ctx context.Context, userID int64) ([]service.ContactItem, error) {
	return m.listContactsFn(ctx, userID)
}

func TestContactsHandler_ListContacts(t *testing.T) {
	h := NewContactsHandler(&mockContactManager{
		listContactsFn: func(_ context.Context, userID int64) ([]service.ContactItem, error) {
			if userID != 1 {
				t.Fatalf("userID = %d, want 1", userID)
			}
			return []service.ContactItem{
				{ID: 2, Email: "friend@local", Username: "friend"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "me@local", Username: "me"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp struct {
		Contacts []service.ContactItem `json:"contacts"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Contacts) != 1 || resp.Contacts[0].Username != "friend" {
		t.Fatalf("response = %+v", resp)
	}
}

func TestContactsHandler_ListContacts_Unauthorized(t *testing.T) {
	h := NewContactsHandler(&mockContactManager{})
	req := httptest.NewRequest(http.MethodGet, "/contacts", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

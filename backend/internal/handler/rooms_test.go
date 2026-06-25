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

type mockDirectRoomOpener struct {
	result  service.DirectRoomResult
	created bool
	err     error
}

func (m *mockDirectRoomOpener) GetOrCreateDirect(_ context.Context, callerID, peerID int64) (service.DirectRoomResult, bool, error) {
	if m.err != nil {
		return service.DirectRoomResult{}, false, m.err
	}
	if m.result.PeerID == 0 {
		m.result.PeerID = peerID
	}
	return m.result, m.created, nil
}

type mockRoomLister struct {
	rooms []service.RoomListItem
	err   error
}

func (m *mockRoomLister) ListRooms(context.Context, int64) ([]service.RoomListItem, error) {
	return m.rooms, m.err
}

func TestRoomsHandler_CreateDirect(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{
			result:  service.DirectRoomResult{ID: 10, PeerID: 2},
			created: true,
		},
		&mockRoomLister{},
	)

	body := `{"user_id":2}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/direct", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var resp directRoomResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ID != 10 || resp.PeerID != 2 {
		t.Fatalf("response = %+v", resp)
	}
}

func TestRoomsHandler_ExistingDirect(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{
			result:  service.DirectRoomResult{ID: 10, PeerID: 2},
			created: false,
		},
		&mockRoomLister{},
	)

	body := `{"user_id":2}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/direct", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRoomsHandler_DirectSelfChat(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{err: domain.ErrSelfChat},
		&mockRoomLister{},
	)

	body := `{"user_id":1}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/direct", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRoomsHandler_DirectPeerNotFound(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{err: repository.ErrNotFound},
		&mockRoomLister{},
	)

	body := `{"user_id":2}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/direct", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRoomsHandler_DirectPeerNotVerified(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{err: service.ErrPeerNotVerified},
		&mockRoomLister{},
	)

	body := `{"user_id":2}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/direct", bytes.NewBufferString(body))
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestRoomsHandler_ListRooms(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{},
		&mockRoomLister{
			rooms: []service.RoomListItem{
				{ID: 10, PeerID: 2, PeerEmail: "peer@example.com"},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp roomListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Rooms) != 1 {
		t.Fatalf("rooms = %+v", resp.Rooms)
	}
}

func TestRoomsHandler_ListRooms_Empty(t *testing.T) {
	h := NewRoomsHandler(&mockDirectRoomOpener{}, &mockRoomLister{})

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp roomListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Rooms) != 0 {
		t.Fatalf("rooms = %+v, want empty slice", resp.Rooms)
	}
}

func TestRoomsHandler_Unauthorized(t *testing.T) {
	h := NewRoomsHandler(&mockDirectRoomOpener{}, &mockRoomLister{})

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRoomsHandler_ListRooms_InternalError(t *testing.T) {
	h := NewRoomsHandler(
		&mockDirectRoomOpener{},
		&mockRoomLister{err: errors.New("db down")},
	)

	req := httptest.NewRequest(http.MethodGet, "/rooms", nil)
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 1, Email: "user1@example.com", Username: "user1"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

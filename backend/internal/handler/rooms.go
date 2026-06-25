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

// DirectRoomOpener opens or creates a direct room.
type DirectRoomOpener interface {
	GetOrCreateDirect(ctx context.Context, callerID, peerID int64) (service.DirectRoomResult, bool, error)
}

// RoomLister returns rooms for the authenticated user.
type RoomLister interface {
	ListRooms(ctx context.Context, userID int64) ([]service.RoomListItem, error)
}

// GroupRoomCreator creates a new group room.
type GroupRoomCreator interface {
	CreateGroup(ctx context.Context, creatorID int64, name string, memberIDs []int64) (int64, error)
}

// RoomsHandler serves room endpoints.
type RoomsHandler struct {
	direct DirectRoomOpener
	list   RoomLister
	groups GroupRoomCreator
}

// NewRoomsHandler creates a rooms HTTP handler.
func NewRoomsHandler(direct DirectRoomOpener, list RoomLister, groups GroupRoomCreator) *RoomsHandler {
	return &RoomsHandler{
		direct: direct,
		list:   list,
		groups: groups,
	}
}

type directRoomRequest struct {
	UserID int64 `json:"user_id"`
}

type directRoomResponse struct {
	ID     int64 `json:"id"`
	PeerID int64 `json:"peer_id"`
}

type createGroupRequest struct {
	Name    string  `json:"name"`
	UserIDs []int64 `json:"user_ids"`
}

type createGroupResponse struct {
	ID int64 `json:"id"`
}

type roomListResponse struct {
	Rooms []service.RoomListItem `json:"rooms"`
}

// ServeHTTP implements http.Handler.
func (h *RoomsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/rooms/direct":
		h.serveDirect(w, r)
	case r.Method == http.MethodPost && r.URL.Path == "/rooms/group":
		h.serveGroup(w, r)
	case r.Method == http.MethodGet && r.URL.Path == "/rooms":
		h.serveList(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *RoomsHandler) serveDirect(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req directRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	result, created, err := h.direct.GetOrCreateDirect(r.Context(), caller.ID, req.UserID)
	if errors.Is(err, domain.ErrSelfChat) || errors.Is(err, domain.ErrInvalidPeerID) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if errors.Is(err, repository.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "peer not found"})
		return
	}
	if errors.Is(err, service.ErrPeerNotVerified) {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: "peer is not verified"})
		return
	}
	if errors.Is(err, domain.ErrNotContact) {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, directRoomResponse{
		ID:     result.ID,
		PeerID: result.PeerID,
	})
}

func (h *RoomsHandler) serveGroup(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid json"})
		return
	}

	roomID, err := h.groups.CreateGroup(r.Context(), caller.ID, req.Name, req.UserIDs)
	if errors.Is(err, domain.ErrInvalidRoomName) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	if errors.Is(err, service.ErrPeerNotVerified) || errors.Is(err, domain.ErrNotContact) {
		writeJSON(w, http.StatusForbidden, errorResponse{Error: err.Error()})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, createGroupResponse{ID: roomID})
}

func (h *RoomsHandler) serveList(w http.ResponseWriter, r *http.Request) {
	caller, ok := AuthUserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}

	rooms, err := h.list.ListRooms(r.Context(), caller.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	if rooms == nil {
		rooms = []service.RoomListItem{}
	}

	writeJSON(w, http.StatusOK, roomListResponse{Rooms: rooms})
}

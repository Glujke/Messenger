package service

import (
	"context"
	"errors"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

var ErrPeerNotVerified = errors.New("peer is not verified")

// RoomService handles direct room use cases.
type RoomService struct {
	users repository.UserStore
	rooms repository.RoomStore
}

// NewRoomService creates a room service.
func NewRoomService(users repository.UserStore, rooms repository.RoomStore) *RoomService {
	return &RoomService{
		users: users,
		rooms: rooms,
	}
}

// DirectRoomResult is returned when opening a direct chat.
type DirectRoomResult struct {
	ID     int64
	PeerID int64
}

// RoomListItem is a room entry for the authenticated user.
type RoomListItem struct {
	ID        int64  `json:"id"`
	PeerID    int64  `json:"peer_id"`
	PeerEmail string `json:"peer_email"`
}

// GetOrCreateDirect returns an existing or newly created 1:1 room.
func (s *RoomService) GetOrCreateDirect(ctx context.Context, callerID, peerID int64) (DirectRoomResult, bool, error) {
	if err := domain.ValidateDirectPeer(callerID, peerID); err != nil {
		return DirectRoomResult{}, false, err
	}

	peer, err := s.users.FindByID(ctx, peerID)
	if errors.Is(err, repository.ErrNotFound) {
		return DirectRoomResult{}, false, repository.ErrNotFound
	}
	if err != nil {
		return DirectRoomResult{}, false, err
	}
	if !peer.Verified {
		return DirectRoomResult{}, false, ErrPeerNotVerified
	}

	roomID, found, err := s.rooms.FindDirectRoom(ctx, callerID, peerID)
	if err != nil {
		return DirectRoomResult{}, false, err
	}
	if found {
		return DirectRoomResult{ID: roomID, PeerID: peerID}, false, nil
	}

	roomID, err = s.rooms.CreateDirectRoom(ctx, callerID, peerID)
	if err != nil {
		return DirectRoomResult{}, false, err
	}

	return DirectRoomResult{ID: roomID, PeerID: peerID}, true, nil
}

// ListRooms returns direct rooms for the authenticated user.
func (s *RoomService) ListRooms(ctx context.Context, userID int64) ([]RoomListItem, error) {
	records, err := s.rooms.ListUserRooms(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]RoomListItem, 0, len(records))
	for _, record := range records {
		items = append(items, RoomListItem{
			ID:        record.ID,
			PeerID:    record.PeerID,
			PeerEmail: record.PeerEmail,
		})
	}
	return items, nil
}

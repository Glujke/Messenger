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
	users    repository.UserStore
	rooms    repository.RoomStore
	contacts repository.ContactStore
}

// NewRoomService creates a room service.
func NewRoomService(users repository.UserStore, rooms repository.RoomStore, contacts repository.ContactStore) *RoomService {
	return &RoomService{
		users:    users,
		rooms:    rooms,
		contacts: contacts,
	}
}

// DirectRoomResult is returned when opening a direct chat.
type DirectRoomResult struct {
	ID     int64
	PeerID int64
}

// RoomListItem is a room entry for the authenticated user.
type RoomListItem struct {
	ID        int64           `json:"id"`
	Kind      domain.RoomKind `json:"kind"`
	Name      string          `json:"name,omitempty"`       // For groups
	PeerID    int64           `json:"peer_id,omitempty"`    // For direct
	PeerEmail string          `json:"peer_email,omitempty"` // For direct
}

// CreateGroup creates a new group room.
func (s *RoomService) CreateGroup(ctx context.Context, creatorID int64, name string, memberIDs []int64) (int64, error) {
	if err := domain.ValidateRoomName(name); err != nil {
		return 0, err
	}

	// Verify all members are verified and are contacts of the creator
	for _, memberID := range memberIDs {
		if memberID == creatorID {
			continue
		}
		peer, err := s.users.FindByID(ctx, memberID)
		if err != nil {
			return 0, err
		}
		if !peer.Verified {
			return 0, ErrPeerNotVerified
		}

		areContacts, err := s.contacts.AreContacts(ctx, creatorID, memberID)
		if err != nil {
			return 0, err
		}
		if !areContacts {
			return 0, domain.ErrNotContact
		}
	}

	return s.rooms.CreateGroupRoom(ctx, name, creatorID, memberIDs)
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

	// Check if users are contacts before allowing direct chat
	areContacts, err := s.contacts.AreContacts(ctx, callerID, peerID)
	if err != nil {
		return DirectRoomResult{}, false, err
	}
	if !areContacts {
		return DirectRoomResult{}, false, domain.ErrNotContact
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

// ListRooms returns direct and group rooms for the authenticated user.
func (s *RoomService) ListRooms(ctx context.Context, userID int64) ([]RoomListItem, error) {
	records, err := s.rooms.ListUserRooms(ctx, userID)
	if err != nil {
		return nil, err
	}

	items := make([]RoomListItem, 0, len(records))
	for _, record := range records {
		item := RoomListItem{
			ID:   record.ID,
			Kind: record.Kind,
		}
		if record.Kind == domain.RoomKindGroup && record.Name != nil {
			item.Name = *record.Name
		} else if record.Kind == domain.RoomKindDirect {
			item.PeerID = record.PeerID
			item.PeerEmail = record.PeerEmail
		}
		items = append(items, item)
	}
	return items, nil
}

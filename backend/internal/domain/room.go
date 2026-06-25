package domain

import (
	"errors"
	"strings"
)

type RoomKind string

const (
	RoomKindDirect RoomKind = "direct"
	RoomKindGroup  RoomKind = "group"
)

var (
	ErrSelfChat        = errors.New("cannot start direct chat with yourself")
	ErrInvalidPeerID   = errors.New("invalid peer id")
	ErrInvalidRoomName = errors.New("room name must be 3-64 characters")
)

// ValidateDirectPeer checks that a direct chat target is valid.
func ValidateDirectPeer(callerID, peerID int64) error {
	if peerID <= 0 {
		return ErrInvalidPeerID
	}
	if callerID == peerID {
		return ErrSelfChat
	}
	return nil
}

// ValidateRoomName checks that a group room name is valid.
func ValidateRoomName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 3 || len(name) > 64 {
		return ErrInvalidRoomName
	}
	return nil
}

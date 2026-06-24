package repository

import (
	"context"
	"time"
)

// RoomRecord is a persisted chat room.
type RoomRecord struct {
	ID        int64
	CreatedAt time.Time
}

// RoomListRecord is a room entry for the authenticated user.
type RoomListRecord struct {
	ID        int64
	PeerID    int64
	PeerEmail string
}

// RoomStore defines room persistence operations.
type RoomStore interface {
	FindDirectRoom(ctx context.Context, userA, userB int64) (int64, bool, error)
	CreateDirectRoom(ctx context.Context, userA, userB int64) (int64, error)
	ListUserRooms(ctx context.Context, userID int64) ([]RoomListRecord, error)
	IsRoomMember(ctx context.Context, roomID, userID int64) (bool, error)
}

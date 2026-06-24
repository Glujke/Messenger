package repository

import (
	"context"
	"time"
)

// MessageRecord is a persisted chat message.
type MessageRecord struct {
	ID        int64
	RoomID    int64
	SenderID  int64
	Type      string
	Body      string
	CreatedAt time.Time
}

// MessageStore defines message persistence operations.
type MessageStore interface {
	SaveMessage(ctx context.Context, roomID, senderID int64, messageType, body string) (MessageRecord, error)
	ListMessages(ctx context.Context, roomID int64, limit int, beforeID int64) ([]MessageRecord, error)
}

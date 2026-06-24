package service

import (
	"context"
	"errors"
	"strings"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

var ErrNotRoomMember = errors.New("not a room member")

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

// MessageService handles chat message use cases.
type MessageService struct {
	rooms    repository.RoomStore
	messages repository.MessageStore
}

// NewMessageService creates a message service.
func NewMessageService(rooms repository.RoomStore, messages repository.MessageStore) *MessageService {
	return &MessageService{
		rooms:    rooms,
		messages: messages,
	}
}

// MessageItem is a message returned to clients.
type MessageItem struct {
	ID        int64  `json:"id"`
	RoomID    int64  `json:"room_id"`
	SenderID  int64  `json:"sender_id"`
	Type      string `json:"type"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// SendText stores a text message in a room.
func (s *MessageService) SendText(ctx context.Context, roomID, senderID int64, body string) (MessageItem, error) {
	if err := domain.ValidateTextBody(body); err != nil {
		return MessageItem{}, err
	}
	body = strings.TrimSpace(body)

	isMember, err := s.rooms.IsRoomMember(ctx, roomID, senderID)
	if err != nil {
		return MessageItem{}, err
	}
	if !isMember {
		return MessageItem{}, ErrNotRoomMember
	}

	record, err := s.messages.SaveMessage(ctx, roomID, senderID, string(domain.MessageTypeText), body)
	if err != nil {
		return MessageItem{}, err
	}

	return toMessageItem(record), nil
}

// List returns recent messages for a room member.
func (s *MessageService) List(ctx context.Context, roomID, userID int64, limit int, beforeID int64) ([]MessageItem, error) {
	if limit <= 0 {
		limit = defaultMessageLimit
	}
	if limit > maxMessageLimit {
		limit = maxMessageLimit
	}

	isMember, err := s.rooms.IsRoomMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrNotRoomMember
	}

	records, err := s.messages.ListMessages(ctx, roomID, limit, beforeID)
	if err != nil {
		return nil, err
	}

	items := make([]MessageItem, 0, len(records))
	for _, record := range records {
		items = append(items, toMessageItem(record))
	}
	return items, nil
}

func toMessageItem(record repository.MessageRecord) MessageItem {
	return MessageItem{
		ID:        record.ID,
		RoomID:    record.RoomID,
		SenderID:  record.SenderID,
		Type:      record.Type,
		Body:      record.Body,
		CreatedAt: record.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}

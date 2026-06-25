package service

import (
	"context"
	"errors"
	"strings"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

var (
	ErrNotRoomMember      = errors.New("not a room member")
	ErrAttachmentNotFound = errors.New("attachment not found")
	ErrAttachmentUsed     = errors.New("attachment already used")
	ErrAttachmentMismatch = errors.New("attachment does not belong to room")
)

const (
	defaultMessageLimit = 50
	maxMessageLimit     = 100
)

// MessageBroadcaster pushes realtime message events to connected clients.
type MessageBroadcaster interface {
	BroadcastMessage(roomID int64, item MessageItem)
}

// MessageService handles chat message use cases.
type MessageService struct {
	rooms       repository.RoomStore
	messages    repository.MessageStore
	attachments repository.AttachmentStore
	broadcaster MessageBroadcaster
	encrypter   domain.MessageEncrypter
}

// NewMessageService creates a message service.
func NewMessageService(
	rooms repository.RoomStore,
	messages repository.MessageStore,
	attachments repository.AttachmentStore,
	broadcaster MessageBroadcaster,
	encrypter domain.MessageEncrypter,
) *MessageService {
	return &MessageService{
		rooms:       rooms,
		messages:    messages,
		attachments: attachments,
		broadcaster: broadcaster,
		encrypter:   encrypter,
	}
}

// AttachmentItem describes a message attachment in API responses.
type AttachmentItem struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

// MessageItem is a message returned to clients.
type MessageItem struct {
	ID         int64           `json:"id"`
	RoomID     int64           `json:"room_id"`
	SenderID   int64           `json:"sender_id"`
	Type       string          `json:"type"`
	Body       string          `json:"body"`
	Attachment *AttachmentItem `json:"attachment,omitempty"`
	CreatedAt  string          `json:"created_at"`
}

// SendText stores a text message in a room.
func (s *MessageService) SendText(ctx context.Context, roomID, senderID int64, body string) (MessageItem, error) {
	if err := domain.ValidateTextBody(body); err != nil {
		return MessageItem{}, err
	}
	body = strings.TrimSpace(body)

	return s.saveAndBroadcast(ctx, roomID, senderID, string(domain.MessageTypeText), body, nil)
}

// SendAttachment links an uploaded attachment to a new message.
func (s *MessageService) SendAttachment(ctx context.Context, roomID, senderID, attachmentID int64, body string) (MessageItem, error) {
	if err := domain.ValidateCaption(body); err != nil {
		return MessageItem{}, err
	}

	attachment, err := s.attachments.FindAttachment(ctx, attachmentID)
	if errors.Is(err, repository.ErrNotFound) {
		return MessageItem{}, ErrAttachmentNotFound
	}
	if err != nil {
		return MessageItem{}, err
	}
	if attachment.RoomID != roomID {
		return MessageItem{}, ErrAttachmentMismatch
	}

	used, err := s.attachments.IsAttachmentUsed(ctx, attachmentID)
	if err != nil {
		return MessageItem{}, err
	}
	if used {
		return MessageItem{}, ErrAttachmentUsed
	}

	messageType, err := domain.MessageTypeForContentType(attachment.ContentType)
	if err != nil {
		return MessageItem{}, err
	}

	return s.saveAndBroadcast(ctx, roomID, senderID, string(messageType), body, &attachmentID)
}

func (s *MessageService) saveAndBroadcast(
	ctx context.Context,
	roomID, senderID int64,
	messageType, body string,
	attachmentID *int64,
) (MessageItem, error) {
	isMember, err := s.rooms.IsRoomMember(ctx, roomID, senderID)
	if err != nil {
		return MessageItem{}, err
	}
	if !isMember {
		return MessageItem{}, ErrNotRoomMember
	}

	encryptedBody, err := s.encrypter.Encrypt(body)
	if err != nil {
		return MessageItem{}, err
	}

	record, err := s.messages.SaveMessage(ctx, roomID, senderID, messageType, encryptedBody, attachmentID)
	if err != nil {
		return MessageItem{}, err
	}

	item, err := toMessageItem(record, s.encrypter)
	if err != nil {
		return MessageItem{}, err
	}
	if s.broadcaster != nil {
		s.broadcaster.BroadcastMessage(roomID, item)
	}
	return item, nil
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
		item, err := toMessageItem(record, s.encrypter)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func toMessageItem(record repository.MessageRecord, encrypter domain.MessageEncrypter) (MessageItem, error) {
	body, err := encrypter.Decrypt(record.Body)
	if err != nil {
		return MessageItem{}, err
	}

	item := MessageItem{
		ID:        record.ID,
		RoomID:    record.RoomID,
		SenderID:  record.SenderID,
		Type:      record.Type,
		Body:      body,
		CreatedAt: record.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
	if record.Attachment != nil {
		item.Attachment = &AttachmentItem{
			ID:          record.Attachment.ID,
			Filename:    record.Attachment.Filename,
			ContentType: record.Attachment.ContentType,
			SizeBytes:   record.Attachment.SizeBytes,
		}
	}
	return item, nil
}

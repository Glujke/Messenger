package domain

import (
	"errors"
	"strings"
	"time"
)

const maxTextBodyLength = 4000

var (
	ErrEmptyMessageBody = errors.New("message body is required")
	ErrMessageTooLong   = errors.New("message body is too long")
)

// MessageType describes how a message should be rendered on the client.
type MessageType string

const (
	MessageTypeText  MessageType = "text"
	MessageTypeImage MessageType = "image"
	MessageTypeFile  MessageType = "file"
)

// Message represents a chat message stored by the API.
type Message struct {
	ID           int64
	RoomID       int64
	SenderID     int64
	Type         MessageType
	Body         string
	AttachmentID *int64
	CreatedAt    time.Time
}

// ValidateTextBody checks that a text message body is non-empty and within limits.
func ValidateTextBody(body string) error {
	body = strings.TrimSpace(body)
	if body == "" {
		return ErrEmptyMessageBody
	}
	if len(body) > maxTextBodyLength {
		return ErrMessageTooLong
	}
	return nil
}

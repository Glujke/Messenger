package domain

import "time"

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

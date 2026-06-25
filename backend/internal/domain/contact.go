package domain

import "time"

// ContactRequestStatus represents the state of a contact invitation.
type ContactRequestStatus string

const (
	ContactRequestPending  ContactRequestStatus = "pending"
	ContactRequestAccepted ContactRequestStatus = "accepted"
	ContactRequestRejected ContactRequestStatus = "rejected"
)

// ContactRequest represents an invitation to connect.
type ContactRequest struct {
	ID           int64
	FromUserID   int64
	ToUserID     int64
	Status       ContactRequestStatus
	CreatedAt    time.Time
	RespondedAt *time.Time
}

// Contact represents a confirmed friendship.
type Contact struct {
	UserID    int64
	ContactID int64
	CreatedAt time.Time
}

package repository

import (
	"context"
	"time"

	"messenger/backend/internal/domain"
)

// ContactRequestRecord is a persisted contact request.
type ContactRequestRecord struct {
	ID           int64
	FromUserID   int64
	ToUserID     int64
	Status       domain.ContactRequestStatus
	CreatedAt    time.Time
	RespondedAt *time.Time
}

// ContactRecord is a persisted friendship.
type ContactRecord struct {
	UserID    int64
	ContactID int64
	CreatedAt time.Time
}

// ContactStore defines contact persistence operations.
type ContactStore interface {
	CreateRequest(ctx context.Context, fromID, toID int64) (ContactRequestRecord, error)
	GetRequest(ctx context.Context, id int64) (ContactRequestRecord, error)
	UpdateRequestStatus(ctx context.Context, id int64, status domain.ContactRequestStatus) error
	ListRequests(ctx context.Context, userID int64) ([]ContactRequestRecord, error)
	
	AddContact(ctx context.Context, userA, userB int64) error
	ListContacts(ctx context.Context, userID int64) ([]ContactRecord, error)
	AreContacts(ctx context.Context, userA, userB int64) (bool, error)
}

package repository

import (
	"context"
	"time"
)

// AttachmentRecord is a persisted uploaded file.
type AttachmentRecord struct {
	ID          int64
	RoomID      int64
	UploaderID  int64
	Filename    string
	ContentType string
	SizeBytes   int64
	StorageKey  string
	CreatedAt   time.Time
}

// AttachmentStore defines attachment persistence operations.
type AttachmentStore interface {
	CreateAttachment(ctx context.Context, record AttachmentRecord) (AttachmentRecord, error)
	FindAttachment(ctx context.Context, id int64) (AttachmentRecord, error)
	IsAttachmentUsed(ctx context.Context, attachmentID int64) (bool, error)
}

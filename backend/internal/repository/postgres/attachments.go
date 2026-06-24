package postgres

import (
	"context"
	"database/sql"
	"errors"

	"messenger/backend/internal/repository"
)

// CreateAttachment inserts attachment metadata.
func (s *Store) CreateAttachment(ctx context.Context, record repository.AttachmentRecord) (repository.AttachmentRecord, error) {
	const query = `
		INSERT INTO attachments (room_id, uploader_id, filename, content_type, size_bytes, storage_key)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, room_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at
	`

	var created repository.AttachmentRecord
	err := s.db.QueryRowContext(ctx, query,
		record.RoomID,
		record.UploaderID,
		record.Filename,
		record.ContentType,
		record.SizeBytes,
		record.StorageKey,
	).Scan(
		&created.ID,
		&created.RoomID,
		&created.UploaderID,
		&created.Filename,
		&created.ContentType,
		&created.SizeBytes,
		&created.StorageKey,
		&created.CreatedAt,
	)
	if err != nil {
		return repository.AttachmentRecord{}, err
	}
	return created, nil
}

// FindAttachment loads attachment metadata by id.
func (s *Store) FindAttachment(ctx context.Context, id int64) (repository.AttachmentRecord, error) {
	const query = `
		SELECT id, room_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at
		FROM attachments
		WHERE id = $1
	`

	var record repository.AttachmentRecord
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.RoomID,
		&record.UploaderID,
		&record.Filename,
		&record.ContentType,
		&record.SizeBytes,
		&record.StorageKey,
		&record.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.AttachmentRecord{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.AttachmentRecord{}, err
	}
	return record, nil
}

// IsAttachmentUsed reports whether a message already references the attachment.
func (s *Store) IsAttachmentUsed(ctx context.Context, attachmentID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM messages
			WHERE attachment_id = $1
		)
	`

	var used bool
	err := s.db.QueryRowContext(ctx, query, attachmentID).Scan(&used)
	if err != nil {
		return false, err
	}
	return used, nil
}

package postgres

import (
	"context"
	"database/sql"

	"messenger/backend/internal/repository"
)

const messageSelectColumns = `
	m.id, m.room_id, m.sender_id, m.type, m.body, m.attachment_id, m.created_at,
	a.id, a.room_id, a.uploader_id, a.filename, a.content_type, a.size_bytes, a.storage_key, a.created_at
`

// SaveMessage inserts a message into a room.
func (s *Store) SaveMessage(ctx context.Context, roomID, senderID int64, messageType, body string, attachmentID *int64) (repository.MessageRecord, error) {
	const query = `
		INSERT INTO messages (room_id, sender_id, type, body, attachment_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, room_id, sender_id, type, body, attachment_id, created_at
	`

	var record repository.MessageRecord
	err := s.db.QueryRowContext(ctx, query, roomID, senderID, messageType, body, attachmentID).Scan(
		&record.ID,
		&record.RoomID,
		&record.SenderID,
		&record.Type,
		&record.Body,
		&record.AttachmentID,
		&record.CreatedAt,
	)
	if err != nil {
		return repository.MessageRecord{}, err
	}

	if attachmentID != nil {
		attachment, err := s.FindAttachment(ctx, *attachmentID)
		if err != nil {
			return repository.MessageRecord{}, err
		}
		record.Attachment = &attachment
	}

	return record, nil
}

// ListMessages returns recent messages for a room.
func (s *Store) ListMessages(ctx context.Context, roomID int64, limit int, beforeID int64) ([]repository.MessageRecord, error) {
	var (
		rows *sql.Rows
		err  error
	)

	baseQuery := `
		SELECT ` + messageSelectColumns + `
		FROM messages m
		LEFT JOIN attachments a ON a.id = m.attachment_id
		WHERE m.room_id = $1
	`

	if beforeID > 0 {
		rows, err = s.db.QueryContext(ctx, baseQuery+`
			AND m.id < $2
			ORDER BY m.id DESC
			LIMIT $3
		`, roomID, beforeID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, baseQuery+`
			ORDER BY m.id DESC
			LIMIT $2
		`, roomID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMessageRows(rows)
}

func scanMessageRows(rows *sql.Rows) ([]repository.MessageRecord, error) {
	var messages []repository.MessageRecord
	for rows.Next() {
		var (
			record     repository.MessageRecord
			attachment repository.AttachmentRecord
			aID        sql.NullInt64
			aRoomID    sql.NullInt64
			aUploader  sql.NullInt64
			aFilename  sql.NullString
			aType      sql.NullString
			aSize      sql.NullInt64
			aKey       sql.NullString
			aCreated   sql.NullTime
		)

		if err := rows.Scan(
			&record.ID,
			&record.RoomID,
			&record.SenderID,
			&record.Type,
			&record.Body,
			&record.AttachmentID,
			&record.CreatedAt,
			&aID,
			&aRoomID,
			&aUploader,
			&aFilename,
			&aType,
			&aSize,
			&aKey,
			&aCreated,
		); err != nil {
			return nil, err
		}

		if aID.Valid {
			attachment = repository.AttachmentRecord{
				ID:          aID.Int64,
				RoomID:      aRoomID.Int64,
				UploaderID:  aUploader.Int64,
				Filename:    aFilename.String,
				ContentType: aType.String,
				SizeBytes:   aSize.Int64,
				StorageKey:  aKey.String,
				CreatedAt:   aCreated.Time,
			}
			record.Attachment = &attachment
		}

		messages = append(messages, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

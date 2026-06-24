package postgres

import (
	"context"
	"database/sql"

	"messenger/backend/internal/repository"
)

// SaveMessage inserts a text message into a room.
func (s *Store) SaveMessage(ctx context.Context, roomID, senderID int64, messageType, body string) (repository.MessageRecord, error) {
	const query = `
		INSERT INTO messages (room_id, sender_id, type, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, room_id, sender_id, type, body, created_at
	`

	var record repository.MessageRecord
	err := s.db.QueryRowContext(ctx, query, roomID, senderID, messageType, body).Scan(
		&record.ID,
		&record.RoomID,
		&record.SenderID,
		&record.Type,
		&record.Body,
		&record.CreatedAt,
	)
	if err != nil {
		return repository.MessageRecord{}, err
	}

	return record, nil
}

// ListMessages returns recent messages for a room.
func (s *Store) ListMessages(ctx context.Context, roomID int64, limit int, beforeID int64) ([]repository.MessageRecord, error) {
	var (
		rows *sql.Rows
		err  error
	)

	if beforeID > 0 {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, room_id, sender_id, type, body, created_at
			FROM messages
			WHERE room_id = $1 AND id < $2
			ORDER BY id DESC
			LIMIT $3
		`, roomID, beforeID, limit)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, room_id, sender_id, type, body, created_at
			FROM messages
			WHERE room_id = $1
			ORDER BY id DESC
			LIMIT $2
		`, roomID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []repository.MessageRecord
	for rows.Next() {
		var record repository.MessageRecord
		if err := rows.Scan(
			&record.ID,
			&record.RoomID,
			&record.SenderID,
			&record.Type,
			&record.Body,
			&record.CreatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

package postgres

import (
	"context"
	"database/sql"
	"errors"

	"messenger/backend/internal/repository"
)

// FindDirectRoom returns an existing 1:1 room shared by two users.
func (s *Store) FindDirectRoom(ctx context.Context, userA, userB int64) (int64, bool, error) {
	const query = `
		SELECT r.id
		FROM rooms r
		JOIN room_members m1 ON m1.room_id = r.id AND m1.user_id = $1
		JOIN room_members m2 ON m2.room_id = r.id AND m2.user_id = $2
		WHERE r.kind = 'direct' AND (
			SELECT COUNT(*)
			FROM room_members
			WHERE room_id = r.id
		) = 2
		LIMIT 1
	`

	var roomID int64
	err := s.db.QueryRowContext(ctx, query, userA, userB).Scan(&roomID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	return roomID, true, nil
}

// CreateDirectRoom creates a new 1:1 room for two users.
func (s *Store) CreateDirectRoom(ctx context.Context, userA, userB int64) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var roomID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO rooms (kind) VALUES ('direct')
		RETURNING id
	`).Scan(&roomID)
	if err != nil {
		return 0, err
	}

	const insertMember = `
		INSERT INTO room_members (room_id, user_id)
		VALUES ($1, $2)
	`
	if _, err := tx.ExecContext(ctx, insertMember, roomID, userA); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, insertMember, roomID, userB); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return roomID, nil
}

// CreateGroupRoom creates a new group room with multiple members.
func (s *Store) CreateGroupRoom(ctx context.Context, name string, creatorID int64, memberIDs []int64) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var roomID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO rooms (kind, name, created_by)
		VALUES ('group', $1, $2)
		RETURNING id
	`, name, creatorID).Scan(&roomID)
	if err != nil {
		return 0, err
	}

	const insertMember = `
		INSERT INTO room_members (room_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	// Add creator
	if _, err := tx.ExecContext(ctx, insertMember, roomID, creatorID); err != nil {
		return 0, err
	}

	// Add other members
	for _, memberID := range memberIDs {
		if _, err := tx.ExecContext(ctx, insertMember, roomID, memberID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return roomID, nil
}

// ListUserRooms returns direct and group rooms for the given user.
func (s *Store) ListUserRooms(ctx context.Context, userID int64) ([]repository.RoomListRecord, error) {
	const query = `
		SELECT 
			r.id, r.kind, r.name, 
			COALESCE(u.id, 0), COALESCE(u.email, '')
		FROM rooms r
		JOIN room_members me ON me.room_id = r.id AND me.user_id = $1
		LEFT JOIN room_members peer ON r.kind = 'direct' AND peer.room_id = r.id AND peer.user_id <> $1
		LEFT JOIN users u ON u.id = peer.user_id
		ORDER BY r.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []repository.RoomListRecord
	for rows.Next() {
		var record repository.RoomListRecord
		if err := rows.Scan(&record.ID, &record.Kind, &record.Name, &record.PeerID, &record.PeerEmail); err != nil {
			return nil, err
		}
		rooms = append(rooms, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

// IsRoomMember reports whether the user belongs to the room.
func (s *Store) IsRoomMember(ctx context.Context, roomID, userID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM room_members
			WHERE room_id = $1 AND user_id = $2
		)
	`

	var isMember bool
	err := s.db.QueryRowContext(ctx, query, roomID, userID).Scan(&isMember)
	if err != nil {
		return false, err
	}
	return isMember, nil
}

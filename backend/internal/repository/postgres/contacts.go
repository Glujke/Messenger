package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

// CreateRequest inserts a new contact invitation.
func (s *Store) CreateRequest(ctx context.Context, fromID, toID int64) (repository.ContactRequestRecord, error) {
	const query = `
		INSERT INTO contact_requests (from_user_id, to_user_id)
		VALUES ($1, $2)
		RETURNING id, from_user_id, to_user_id, status, created_at, responded_at
	`

	var record repository.ContactRequestRecord
	err := s.db.QueryRowContext(ctx, query, fromID, toID).Scan(
		&record.ID,
		&record.FromUserID,
		&record.ToUserID,
		&record.Status,
		&record.CreatedAt,
		&record.RespondedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ContactRequestRecord{}, domain.ErrRequestPending
		}
		return repository.ContactRequestRecord{}, err
	}

	return record, nil
}

// GetRequest loads a contact request by ID.
func (s *Store) GetRequest(ctx context.Context, id int64) (repository.ContactRequestRecord, error) {
	const query = `
		SELECT id, from_user_id, to_user_id, status, created_at, responded_at
		FROM contact_requests
		WHERE id = $1
	`

	var record repository.ContactRequestRecord
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.FromUserID,
		&record.ToUserID,
		&record.Status,
		&record.CreatedAt,
		&record.RespondedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.ContactRequestRecord{}, domain.ErrRequestNotFound
	}
	if err != nil {
		return repository.ContactRequestRecord{}, err
	}

	return record, nil
}

// UpdateRequestStatus changes the status of an invitation.
func (s *Store) UpdateRequestStatus(ctx context.Context, id int64, status domain.ContactRequestStatus) error {
	const query = `
		UPDATE contact_requests
		SET status = $1, responded_at = now()
		WHERE id = $2
	`

	res, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrRequestNotFound
	}

	return nil
}

// ListRequests returns all requests (incoming and outgoing) for a user.
func (s *Store) ListRequests(ctx context.Context, userID int64) ([]repository.ContactRequestRecord, error) {
	const query = `
		SELECT id, from_user_id, to_user_id, status, created_at, responded_at
		FROM contact_requests
		WHERE from_user_id = $1 OR to_user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []repository.ContactRequestRecord
	for rows.Next() {
		var r repository.ContactRequestRecord
		if err := rows.Scan(&r.ID, &r.FromUserID, &r.ToUserID, &r.Status, &r.CreatedAt, &r.RespondedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, rows.Err()
}

// AddContact creates a bidirectional friendship.
func (s *Store) AddContact(ctx context.Context, userA, userB int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const query = `
		INSERT INTO contacts (user_id, contact_id)
		VALUES ($1, $2), ($2, $1)
		ON CONFLICT DO NOTHING
	`

	if _, err := tx.ExecContext(ctx, query, userA, userB); err != nil {
		return err
	}

	return tx.Commit()
}

// ListContacts returns all confirmed friends for a user.
func (s *Store) ListContacts(ctx context.Context, userID int64) ([]repository.ContactRecord, error) {
	const query = `
		SELECT user_id, contact_id, created_at
		FROM contacts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []repository.ContactRecord
	for rows.Next() {
		var r repository.ContactRecord
		if err := rows.Scan(&r.UserID, &r.ContactID, &r.CreatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, rows.Err()
}

// AreContacts checks if two users are confirmed friends.
func (s *Store) AreContacts(ctx context.Context, userA, userB int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM contacts
			WHERE user_id = $1 AND contact_id = $2
		)
	`

	var exists bool
	err := s.db.QueryRowContext(ctx, query, userA, userB).Scan(&exists)
	return exists, err
}

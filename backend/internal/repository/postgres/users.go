package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"messenger/backend/internal/repository"
)

// CreateUser inserts a new user and returns the persisted record.
func (s *Store) CreateUser(ctx context.Context, email, passwordHash string) (repository.UserRecord, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, verified, created_at
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email), passwordHash).Scan(
		&record.ID,
		&record.Email,
		&record.PasswordHash,
		&record.Verified,
		&record.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.UserRecord{}, repository.ErrEmailTaken
		}
		return repository.UserRecord{}, err
	}

	return record, nil
}

// FindByEmail loads a user by email address.
func (s *Store) FindByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	const query = `
		SELECT id, email, password_hash, verified, created_at
		FROM users
		WHERE email = $1
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&record.ID,
		&record.Email,
		&record.PasswordHash,
		&record.Verified,
		&record.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.UserRecord{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.UserRecord{}, err
	}

	return record, nil
}

// FindByID loads a user by primary key.
func (s *Store) FindByID(ctx context.Context, id int64) (repository.UserRecord, error) {
	const query = `
		SELECT id, email, password_hash, verified, created_at
		FROM users
		WHERE id = $1
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.Email,
		&record.PasswordHash,
		&record.Verified,
		&record.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.UserRecord{}, repository.ErrNotFound
	}
	if err != nil {
		return repository.UserRecord{}, err
	}

	return record, nil
}

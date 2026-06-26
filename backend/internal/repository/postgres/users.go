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
func (s *Store) CreateUser(ctx context.Context, email, username, passwordHash string) (repository.UserRecord, error) {
	const query = `
		INSERT INTO users (email, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, username, password_hash, verified, created_at
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email), strings.ToLower(username), passwordHash).Scan(
		&record.ID,
		&record.Email,
		&record.Username,
		&record.PasswordHash,
		&record.Verified,
		&record.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return repository.UserRecord{}, repository.ErrUsernameTaken
			}
			return repository.UserRecord{}, repository.ErrEmailTaken
		}
		return repository.UserRecord{}, err
	}

	return record, nil
}

// FindByEmail loads a user by email address.
func (s *Store) FindByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	const query = `
		SELECT id, email, username, password_hash, verified, created_at
		FROM users
		WHERE email = $1
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(email)).Scan(
		&record.ID,
		&record.Email,
		&record.Username,
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

// FindByUsername loads a user by username.
func (s *Store) FindByUsername(ctx context.Context, username string) (repository.UserRecord, error) {
	const query = `
		SELECT id, email, username, password_hash, verified, created_at
		FROM users
		WHERE username = $1
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, strings.ToLower(username)).Scan(
		&record.ID,
		&record.Email,
		&record.Username,
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
		SELECT id, email, username, password_hash, verified, created_at
		FROM users
		WHERE id = $1
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.Email,
		&record.Username,
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

// UpdateUsername changes the username for an existing user.
func (s *Store) UpdateUsername(ctx context.Context, id int64, username string) (repository.UserRecord, error) {
	const query = `
		UPDATE users
		SET username = $2
		WHERE id = $1
		RETURNING id, email, username, password_hash, verified, created_at
	`

	var record repository.UserRecord
	err := s.db.QueryRowContext(ctx, query, id, strings.ToLower(username)).Scan(
		&record.ID,
		&record.Email,
		&record.Username,
		&record.PasswordHash,
		&record.Verified,
		&record.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return repository.UserRecord{}, repository.ErrNotFound
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && strings.Contains(pgErr.ConstraintName, "username") {
			return repository.UserRecord{}, repository.ErrUsernameTaken
		}
		return repository.UserRecord{}, err
	}

	return record, nil
}

// UpdatePasswordHash replaces the password hash for an existing user.
func (s *Store) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error {
	const query = `
		UPDATE users
		SET password_hash = $2
		WHERE id = $1
	`

	res, err := s.db.ExecContext(ctx, query, id, passwordHash)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrNotFound
	}
	return nil
}

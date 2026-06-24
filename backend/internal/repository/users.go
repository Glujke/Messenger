package repository

import (
	"context"
	"errors"
	"time"
)

var (
	ErrEmailTaken = errors.New("email already taken")
	ErrNotFound   = errors.New("user not found")
)

// UserRecord is a persisted user with credentials metadata.
type UserRecord struct {
	ID           int64
	Email        string
	PasswordHash string
	Verified     bool
	CreatedAt    time.Time
}

// UserStore defines user persistence operations.
type UserStore interface {
	CreateUser(ctx context.Context, email, passwordHash string) (UserRecord, error)
	FindByEmail(ctx context.Context, email string) (UserRecord, error)
	FindByID(ctx context.Context, id int64) (UserRecord, error)
}

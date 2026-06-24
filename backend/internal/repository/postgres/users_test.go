package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"

	"messenger/backend/internal/repository"
)

func TestStore_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	createdAt := time.Now()

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "verified", "created_at"}).
			AddRow(int64(1), "user@example.com", "hash", false, createdAt))

	record, err := store.CreateUser(context.Background(), "User@Example.com", "hash")
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if record.ID != 1 {
		t.Fatalf("ID = %d, want 1", record.ID)
	}
	if record.Email != "user@example.com" {
		t.Fatalf("Email = %q, want %q", record.Email, "user@example.com")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_CreateUser_DuplicateEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("user@example.com", "hash").
		WillReturnError(&pgconn.PgError{Code: "23505"})

	_, err = store.CreateUser(context.Background(), "user@example.com", "hash")
	if !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("CreateUser() error = %v, want %v", err, repository.ErrEmailTaken)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_FindByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	createdAt := time.Now()

	mock.ExpectQuery(`SELECT id, email, password_hash, verified, created_at`).
		WithArgs("user@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password_hash", "verified", "created_at"}).
			AddRow(int64(1), "user@example.com", "hash", true, createdAt))

	record, err := store.FindByEmail(context.Background(), "User@Example.com")
	if err != nil {
		t.Fatalf("FindByEmail() error = %v", err)
	}
	if !record.Verified {
		t.Fatal("Verified = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_FindByEmail_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT id, email, password_hash, verified, created_at`).
		WithArgs("missing@example.com").
		WillReturnError(sql.ErrNoRows)

	_, err = store.FindByEmail(context.Background(), "missing@example.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("FindByEmail() error = %v, want %v", err, repository.ErrNotFound)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

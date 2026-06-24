package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"messenger/backend/internal/repository"
)

func TestStore_CreateAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	createdAt := time.Now()

	mock.ExpectQuery(`INSERT INTO attachments`).
		WithArgs(int64(1), int64(2), "photo.png", "image/png", int64(100), "abc.png").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_id", "uploader_id", "filename", "content_type", "size_bytes", "storage_key", "created_at",
		}).AddRow(int64(5), int64(1), int64(2), "photo.png", "image/png", int64(100), "abc.png", createdAt))

	record, err := store.CreateAttachment(context.Background(), repository.AttachmentRecord{
		RoomID:      1,
		UploaderID:  2,
		Filename:    "photo.png",
		ContentType: "image/png",
		SizeBytes:   100,
		StorageKey:  "abc.png",
	})
	if err != nil {
		t.Fatalf("CreateAttachment() error = %v", err)
	}
	if record.ID != 5 {
		t.Fatalf("record = %+v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_FindAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	createdAt := time.Now()

	mock.ExpectQuery(`SELECT id, room_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_id", "uploader_id", "filename", "content_type", "size_bytes", "storage_key", "created_at",
		}).AddRow(int64(5), int64(1), int64(2), "photo.png", "image/png", int64(100), "abc.png", createdAt))

	record, err := store.FindAttachment(context.Background(), 5)
	if err != nil {
		t.Fatalf("FindAttachment() error = %v", err)
	}
	if record.Filename != "photo.png" {
		t.Fatalf("record = %+v", record)
	}
}

func TestStore_FindAttachment_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT id, room_id, uploader_id, filename, content_type, size_bytes, storage_key, created_at`).
		WithArgs(int64(99)).
		WillReturnError(sql.ErrNoRows)

	_, err = store.FindAttachment(context.Background(), 99)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("FindAttachment() error = %v", err)
	}
}

func TestStore_IsAttachmentUsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(5)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	used, err := store.IsAttachmentUsed(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if !used {
		t.Fatal("used = false, want true")
	}
}

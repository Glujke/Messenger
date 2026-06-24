package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStore_SaveMessage(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	createdAt := time.Now()

	mock.ExpectQuery(`INSERT INTO messages`).
		WithArgs(int64(1), int64(2), "text", "hello", nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_id", "sender_id", "type", "body", "attachment_id", "created_at"}).
			AddRow(int64(10), int64(1), int64(2), "text", "hello", nil, createdAt))

	record, err := store.SaveMessage(context.Background(), 1, 2, "text", "hello", nil)
	if err != nil {
		t.Fatalf("SaveMessage() error = %v", err)
	}
	if record.ID != 10 || record.Body != "hello" {
		t.Fatalf("record = %+v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_ListMessages(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	createdAt := time.Now()

	mock.ExpectQuery(`FROM messages m`).
		WithArgs(int64(1), 50).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_id", "sender_id", "type", "body", "attachment_id", "created_at",
			"id", "room_id", "uploader_id", "filename", "content_type", "size_bytes", "storage_key", "created_at",
		}).AddRow(
			int64(10), int64(1), int64(2), "text", "hello", nil, createdAt,
			nil, nil, nil, nil, nil, nil, nil, nil,
		))

	messages, err := store.ListMessages(context.Background(), 1, 50, 0)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != 10 {
		t.Fatalf("messages = %+v", messages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_ListMessages_BeforeID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`FROM messages m`).
		WithArgs(int64(1), int64(20), 50).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_id", "sender_id", "type", "body", "attachment_id", "created_at",
			"id", "room_id", "uploader_id", "filename", "content_type", "size_bytes", "storage_key", "created_at",
		}))

	messages, err := store.ListMessages(context.Background(), 1, 50, 20)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 0 {
		t.Fatalf("messages = %+v, want empty", messages)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_IsRoomMember(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	isMember, err := store.IsRoomMember(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("IsRoomMember() error = %v", err)
	}
	if !isMember {
		t.Fatal("isMember = false, want true")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

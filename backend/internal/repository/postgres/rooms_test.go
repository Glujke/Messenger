package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStore_FindDirectRoom_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT r.id`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(10)))

	roomID, found, err := store.FindDirectRoom(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("FindDirectRoom() error = %v", err)
	}
	if !found {
		t.Fatal("found = false, want true")
	}
	if roomID != 10 {
		t.Fatalf("roomID = %d, want 10", roomID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_FindDirectRoom_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT r.id`).
		WithArgs(int64(1), int64(2)).
		WillReturnError(sql.ErrNoRows)

	roomID, found, err := store.FindDirectRoom(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("FindDirectRoom() error = %v", err)
	}
	if found {
		t.Fatal("found = true, want false")
	}
	if roomID != 0 {
		t.Fatalf("roomID = %d, want 0", roomID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_CreateDirectRoom(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO rooms`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(5)))
	mock.ExpectExec(`INSERT INTO room_members`).
		WithArgs(int64(5), int64(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO room_members`).
		WithArgs(int64(5), int64(2)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	roomID, err := store.CreateDirectRoom(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("CreateDirectRoom() error = %v", err)
	}
	if roomID != 5 {
		t.Fatalf("roomID = %d, want 5", roomID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStore_ListUserRooms(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT r.id, u.id, u.email`).
		WithArgs(int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "id", "email"}).
			AddRow(int64(10), int64(2), "peer@example.com"))

	rooms, err := store.ListUserRooms(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListUserRooms() error = %v", err)
	}
	if len(rooms) != 1 {
		t.Fatalf("len(rooms) = %d, want 1", len(rooms))
	}
	if rooms[0].ID != 10 || rooms[0].PeerID != 2 || rooms[0].PeerEmail != "peer@example.com" {
		t.Fatalf("rooms[0] = %+v", rooms[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

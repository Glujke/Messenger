package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"messenger/backend/internal/domain"
)

func TestStore_CreateRequest(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)
	now := time.Now()

	mock.ExpectQuery(`INSERT INTO contact_requests`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "from_user_id", "to_user_id", "status", "created_at", "responded_at"}).
			AddRow(int64(10), int64(1), int64(2), "pending", now, nil))

	record, err := store.CreateRequest(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if record.ID != 10 || record.Status != domain.ContactRequestPending {
		t.Fatalf("record = %+v", record)
	}
}

func TestStore_AreContacts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := NewStore(db)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(int64(1), int64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	are, err := store.AreContacts(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !are {
		t.Fatal("want true")
	}
}

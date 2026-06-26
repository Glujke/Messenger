package service

import (
	"context"
	"testing"
	"time"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

type mockContactStore struct {
	areContactsFn    func(int64, int64) (bool, error)
	listContactsFn   func(ctx context.Context, userID int64) ([]repository.ContactRecord, error)
	listRequestsFn   func(ctx context.Context, userID int64) ([]repository.ContactRequestRecord, error)
}

func (m *mockContactStore) CreateRequest(ctx context.Context, fromID, toID int64) (repository.ContactRequestRecord, error) {
	return repository.ContactRequestRecord{}, nil
}
func (m *mockContactStore) GetRequest(ctx context.Context, id int64) (repository.ContactRequestRecord, error) {
	return repository.ContactRequestRecord{}, nil
}
func (m *mockContactStore) UpdateRequestStatus(ctx context.Context, id int64, status domain.ContactRequestStatus) error {
	return nil
}
func (m *mockContactStore) ListRequests(ctx context.Context, userID int64) ([]repository.ContactRequestRecord, error) {
	if m.listRequestsFn != nil {
		return m.listRequestsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockContactStore) AddContact(ctx context.Context, userA, userB int64) error {
	return nil
}
func (m *mockContactStore) ListContacts(ctx context.Context, userID int64) ([]repository.ContactRecord, error) {
	if m.listContactsFn != nil {
		return m.listContactsFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockContactStore) AreContacts(ctx context.Context, userA, userB int64) (bool, error) {
	if m.areContactsFn != nil {
		return m.areContactsFn(userA, userB)
	}
	return false, nil
}

type mockContactUserStore struct {
	findByIDFn func(ctx context.Context, id int64) (repository.UserRecord, error)
}

func (m *mockContactUserStore) CreateUser(context.Context, string, string, string) (repository.UserRecord, error) {
	panic("not implemented")
}
func (m *mockContactUserStore) FindByEmail(context.Context, string) (repository.UserRecord, error) {
	panic("not implemented")
}
func (m *mockContactUserStore) FindByUsername(context.Context, string) (repository.UserRecord, error) {
	panic("not implemented")
}
func (m *mockContactUserStore) FindByID(ctx context.Context, id int64) (repository.UserRecord, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockContactUserStore) UpdateUsername(context.Context, int64, string) (repository.UserRecord, error) {
	panic("not implemented")
}
func (m *mockContactUserStore) UpdatePasswordHash(context.Context, int64, string) error {
	panic("not implemented")
}

func TestContactService_ListContacts_ReturnsPeerProfile(t *testing.T) {
	now := time.Now()
	svc := NewContactService(
		&mockContactUserStore{
			findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
				if id != 2 {
					t.Fatalf("id = %d, want 2", id)
				}
				return repository.UserRecord{ID: 2, Email: "friend@local", Username: "friend"}, nil
			},
		},
		&mockContactStore{
			listContactsFn: func(_ context.Context, userID int64) ([]repository.ContactRecord, error) {
				return []repository.ContactRecord{
					{UserID: userID, ContactID: 2, CreatedAt: now},
				}, nil
			},
		},
	)

	items, err := svc.ListContacts(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListContacts() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ID != 2 || items[0].Email != "friend@local" || items[0].Username != "friend" {
		t.Fatalf("item = %+v", items[0])
	}
}

func TestContactService_ListRequests_IncludesPeerProfile(t *testing.T) {
	now := time.Now()
	svc := NewContactService(
		&mockContactUserStore{
			findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
				return repository.UserRecord{ID: id, Email: "peer@local", Username: "peer"}, nil
			},
		},
		&mockContactStore{
			listRequestsFn: func(_ context.Context, userID int64) ([]repository.ContactRequestRecord, error) {
				return []repository.ContactRequestRecord{
					{ID: 5, FromUserID: 2, ToUserID: userID, Status: domain.ContactRequestPending, CreatedAt: now},
				}, nil
			},
		},
	)

	items, err := svc.ListRequests(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListRequests() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].PeerEmail != "peer@local" || items[0].PeerUsername != "peer" {
		t.Fatalf("item = %+v", items[0])
	}
}

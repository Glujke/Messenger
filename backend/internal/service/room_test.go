package service

import (
	"context"
	"errors"
	"testing"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

type mockRoomStore struct {
	findDirectFn func(ctx context.Context, userA, userB int64) (int64, bool, error)
	createFn     func(ctx context.Context, userA, userB int64) (int64, error)
	listFn       func(ctx context.Context, userID int64) ([]repository.RoomListRecord, error)
}

func (m *mockRoomStore) FindDirectRoom(ctx context.Context, userA, userB int64) (int64, bool, error) {
	return m.findDirectFn(ctx, userA, userB)
}

func (m *mockRoomStore) CreateDirectRoom(ctx context.Context, userA, userB int64) (int64, error) {
	return m.createFn(ctx, userA, userB)
}

func (m *mockRoomStore) ListUserRooms(ctx context.Context, userID int64) ([]repository.RoomListRecord, error) {
	return m.listFn(ctx, userID)
}

func (m *mockRoomStore) IsRoomMember(context.Context, int64, int64) (bool, error) {
	return false, nil
}

type mockUserStoreWithID struct {
	findByIDFn func(ctx context.Context, id int64) (repository.UserRecord, error)
}

func (m *mockUserStoreWithID) CreateUser(context.Context, string, string, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockUserStoreWithID) FindByEmail(context.Context, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockUserStoreWithID) FindByUsername(context.Context, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockUserStoreWithID) FindByID(ctx context.Context, id int64) (repository.UserRecord, error) {
	return m.findByIDFn(ctx, id)
}

func TestRoomService_GetOrCreateDirect_Creates(t *testing.T) {
	users := &mockUserStoreWithID{
		findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
			return repository.UserRecord{ID: id, Email: "peer@example.com", Verified: true}, nil
		},
	}
	rooms := &mockRoomStore{
		findDirectFn: func(context.Context, int64, int64) (int64, bool, error) {
			return 0, false, nil
		},
		createFn: func(_ context.Context, userA, userB int64) (int64, error) {
			if userA != 1 || userB != 2 {
				t.Fatalf("users = %d and %d", userA, userB)
			}
			return 10, nil
		},
	}
	contacts := &mockContactStore{
		areContactsFn: func(int64, int64) (bool, error) { return true, nil },
	}

	svc := NewRoomService(users, rooms, contacts)
	result, created, err := svc.GetOrCreateDirect(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("GetOrCreateDirect() error = %v", err)
	}
	if !created {
		t.Fatal("created = false, want true")
	}
	if result.ID != 10 || result.PeerID != 2 {
		t.Fatalf("result = %+v", result)
	}
}

func TestRoomService_GetOrCreateDirect_NotContacts(t *testing.T) {
	users := &mockUserStoreWithID{
		findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
			return repository.UserRecord{ID: id, Verified: true}, nil
		},
	}
	contacts := &mockContactStore{
		areContactsFn: func(int64, int64) (bool, error) { return false, nil },
	}

	svc := NewRoomService(users, &mockRoomStore{}, contacts)
	_, _, err := svc.GetOrCreateDirect(context.Background(), 1, 2)
	if !errors.Is(err, domain.ErrNotContact) {
		t.Fatalf("error = %v, want %v", err, domain.ErrNotContact)
	}
}

func TestRoomService_GetOrCreateDirect_Existing(t *testing.T) {
	users := &mockUserStoreWithID{
		findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
			return repository.UserRecord{ID: id, Verified: true}, nil
		},
	}
	rooms := &mockRoomStore{
		findDirectFn: func(context.Context, int64, int64) (int64, bool, error) {
			return 10, true, nil
		},
		createFn: func(context.Context, int64, int64) (int64, error) {
			t.Fatal("CreateDirectRoom must not be called")
			return 0, nil
		},
	}
	contacts := &mockContactStore{
		areContactsFn: func(int64, int64) (bool, error) { return true, nil },
	}

	svc := NewRoomService(users, rooms, contacts)
	result, created, err := svc.GetOrCreateDirect(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("GetOrCreateDirect() error = %v", err)
	}
	if created {
		t.Fatal("created = true, want false")
	}
	if result.ID != 10 {
		t.Fatalf("result = %+v", result)
	}
}

func TestRoomService_GetOrCreateDirect_SelfChat(t *testing.T) {
	svc := NewRoomService(&mockUserStoreWithID{}, &mockRoomStore{}, &mockContactStore{})
	_, _, err := svc.GetOrCreateDirect(context.Background(), 1, 1)
	if !errors.Is(err, domain.ErrSelfChat) {
		t.Fatalf("error = %v, want %v", err, domain.ErrSelfChat)
	}
}

func TestRoomService_GetOrCreateDirect_PeerNotFound(t *testing.T) {
	users := &mockUserStoreWithID{
		findByIDFn: func(context.Context, int64) (repository.UserRecord, error) {
			return repository.UserRecord{}, repository.ErrNotFound
		},
	}
	svc := NewRoomService(users, &mockRoomStore{}, &mockContactStore{})
	_, _, err := svc.GetOrCreateDirect(context.Background(), 1, 2)
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("error = %v, want %v", err, repository.ErrNotFound)
	}
}

func TestRoomService_GetOrCreateDirect_PeerNotVerified(t *testing.T) {
	users := &mockUserStoreWithID{
		findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
			return repository.UserRecord{ID: id, Verified: false}, nil
		},
	}
	svc := NewRoomService(users, &mockRoomStore{}, &mockContactStore{})
	_, _, err := svc.GetOrCreateDirect(context.Background(), 1, 2)
	if !errors.Is(err, ErrPeerNotVerified) {
		t.Fatalf("error = %v, want %v", err, ErrPeerNotVerified)
	}
}

func TestRoomService_ListRooms(t *testing.T) {
	rooms := &mockRoomStore{
		listFn: func(_ context.Context, userID int64) ([]repository.RoomListRecord, error) {
			if userID != 1 {
				t.Fatalf("userID = %d", userID)
			}
			return []repository.RoomListRecord{
				{ID: 10, PeerID: 2, PeerEmail: "peer@example.com"},
			}, nil
		},
	}

	svc := NewRoomService(&mockUserStoreWithID{}, rooms, &mockContactStore{})
	items, err := svc.ListRooms(context.Background(), 1)
	if err != nil {
		t.Fatalf("ListRooms() error = %v", err)
	}
	if len(items) != 1 || items[0].PeerEmail != "peer@example.com" {
		t.Fatalf("items = %+v", items)
	}
}

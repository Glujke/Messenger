package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

type mockMessageStore struct {
	saveFn func(ctx context.Context, roomID, senderID int64, messageType, body string) (repository.MessageRecord, error)
	listFn func(ctx context.Context, roomID int64, limit int, beforeID int64) ([]repository.MessageRecord, error)
}

func (m *mockMessageStore) SaveMessage(ctx context.Context, roomID, senderID int64, messageType, body string) (repository.MessageRecord, error) {
	return m.saveFn(ctx, roomID, senderID, messageType, body)
}

func (m *mockMessageStore) ListMessages(ctx context.Context, roomID int64, limit int, beforeID int64) ([]repository.MessageRecord, error) {
	return m.listFn(ctx, roomID, limit, beforeID)
}

type mockRoomStoreWithMember struct {
	isMemberFn func(ctx context.Context, roomID, userID int64) (bool, error)
}

func (m *mockRoomStoreWithMember) FindDirectRoom(context.Context, int64, int64) (int64, bool, error) {
	panic("not implemented")
}

func (m *mockRoomStoreWithMember) CreateDirectRoom(context.Context, int64, int64) (int64, error) {
	panic("not implemented")
}

func (m *mockRoomStoreWithMember) ListUserRooms(context.Context, int64) ([]repository.RoomListRecord, error) {
	panic("not implemented")
}

func (m *mockRoomStoreWithMember) IsRoomMember(ctx context.Context, roomID, userID int64) (bool, error) {
	return m.isMemberFn(ctx, roomID, userID)
}

func TestMessageService_SendText(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(_ context.Context, roomID, userID int64) (bool, error) {
			return roomID == 1 && userID == 2, nil
		},
	}
	messages := &mockMessageStore{
		saveFn: func(_ context.Context, roomID, senderID int64, messageType, body string) (repository.MessageRecord, error) {
			return repository.MessageRecord{
				ID:        10,
				RoomID:    roomID,
				SenderID:  senderID,
				Type:      messageType,
				Body:      body,
				CreatedAt: time.Unix(1, 0),
			}, nil
		},
	}

	svc := NewMessageService(rooms, messages)
	item, err := svc.SendText(context.Background(), 1, 2, "hello")
	if err != nil {
		t.Fatalf("SendText() error = %v", err)
	}
	if item.ID != 10 || item.Body != "hello" {
		t.Fatalf("item = %+v", item)
	}
}

func TestMessageService_SendText_NotMember(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return false, nil
		},
	}
	svc := NewMessageService(rooms, &mockMessageStore{})
	_, err := svc.SendText(context.Background(), 1, 2, "hello")
	if !errors.Is(err, ErrNotRoomMember) {
		t.Fatalf("SendText() error = %v, want %v", err, ErrNotRoomMember)
	}
}

func TestMessageService_SendText_InvalidBody(t *testing.T) {
	svc := NewMessageService(&mockRoomStoreWithMember{}, &mockMessageStore{})
	_, err := svc.SendText(context.Background(), 1, 2, "   ")
	if !errors.Is(err, domain.ErrEmptyMessageBody) {
		t.Fatalf("SendText() error = %v, want %v", err, domain.ErrEmptyMessageBody)
	}
}

func TestMessageService_List(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(_ context.Context, roomID, userID int64) (bool, error) {
			return roomID == 1 && userID == 2, nil
		},
	}
	messages := &mockMessageStore{
		listFn: func(_ context.Context, roomID int64, limit int, beforeID int64) ([]repository.MessageRecord, error) {
			return []repository.MessageRecord{
				{ID: 10, RoomID: roomID, SenderID: 2, Type: "text", Body: "hello", CreatedAt: time.Unix(1, 0)},
			}, nil
		},
	}

	svc := NewMessageService(rooms, messages)
	items, err := svc.List(context.Background(), 1, 2, 0, 0)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].Body != "hello" {
		t.Fatalf("items = %+v", items)
	}
}

func TestMessageService_List_NotMember(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return false, nil
		},
	}
	svc := NewMessageService(rooms, &mockMessageStore{})
	_, err := svc.List(context.Background(), 1, 2, 0, 0)
	if !errors.Is(err, ErrNotRoomMember) {
		t.Fatalf("List() error = %v, want %v", err, ErrNotRoomMember)
	}
}

func TestMessageService_List_DefaultLimit(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return true, nil
		},
	}
	var gotLimit int
	messages := &mockMessageStore{
		listFn: func(_ context.Context, _ int64, limit int, _ int64) ([]repository.MessageRecord, error) {
			gotLimit = limit
			return nil, nil
		},
	}

	svc := NewMessageService(rooms, messages)
	if _, err := svc.List(context.Background(), 1, 2, 0, 0); err != nil {
		t.Fatal(err)
	}
	if gotLimit != defaultMessageLimit {
		t.Fatalf("limit = %d, want %d", gotLimit, defaultMessageLimit)
	}
}

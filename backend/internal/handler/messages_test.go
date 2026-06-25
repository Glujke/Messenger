package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
	"messenger/backend/internal/service"
)

type mockMessageSender struct {
	item service.MessageItem
	err  error
}

func (m *mockMessageSender) SendText(_ context.Context, _, _ int64, body string) (service.MessageItem, error) {
	if m.err != nil {
		return service.MessageItem{}, m.err
	}
	if m.item.Body == "" {
		m.item.Body = body
	}
	return m.item, nil
}

func (m *mockMessageSender) SendAttachment(_ context.Context, _, _ int64, attachmentID int64, body string) (service.MessageItem, error) {
	if m.err != nil {
		return service.MessageItem{}, m.err
	}
	item := m.item
	if item.ID == 0 {
		item.ID = 10
	}
	item.Body = body
	item.Type = "image"
	item.Attachment = &service.AttachmentItem{ID: attachmentID}
	return item, nil
}

type mockMessageLister struct {
	messages []service.MessageItem
	err      error
}

func (m *mockMessageLister) List(context.Context, int64, int64, int, int64) ([]service.MessageItem, error) {
	return m.messages, m.err
}

func TestMessagesHandler_Send(t *testing.T) {
	h := NewMessagesHandler(
		&mockMessageSender{item: service.MessageItem{ID: 10, RoomID: 1, SenderID: 2, Type: "text", Body: "hello"}},
		&mockMessageLister{},
	)

	body := `{"body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/messages", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var item service.MessageItem
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.ID != 10 || item.Body != "hello" {
		t.Fatalf("item = %+v", item)
	}
}

func TestMessagesHandler_SendAttachment(t *testing.T) {
	attachmentID := int64(5)
	h := NewMessagesHandler(
		&mockMessageSender{item: service.MessageItem{ID: 11, RoomID: 1, Type: "image"}},
		&mockMessageLister{},
	)

	body := `{"attachment_id":5,"body":"caption"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/messages", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}

	var item service.MessageItem
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.Type != "image" || item.Attachment == nil || item.Attachment.ID != attachmentID {
		t.Fatalf("item = %+v", item)
	}
}

func TestMessagesHandler_Send_NotMember(t *testing.T) {
	h := NewMessagesHandler(
		&mockMessageSender{err: service.ErrNotRoomMember},
		&mockMessageLister{},
	)

	body := `{"body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/messages", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestMessagesHandler_Send_InvalidBody(t *testing.T) {
	h := NewMessagesHandler(
		&mockMessageSender{err: domain.ErrEmptyMessageBody},
		&mockMessageLister{},
	)

	body := `{"body":"   "}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/messages", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMessagesHandler_List(t *testing.T) {
	h := NewMessagesHandler(
		&mockMessageSender{},
		&mockMessageLister{
			messages: []service.MessageItem{
				{ID: 10, RoomID: 1, SenderID: 2, Type: "text", Body: "hello"},
			},
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/rooms/1/messages", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp messageListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Messages) != 1 {
		t.Fatalf("messages = %+v", resp.Messages)
	}
}

func TestMessagesHandler_List_Empty(t *testing.T) {
	h := NewMessagesHandler(&mockMessageSender{}, &mockMessageLister{})

	req := httptest.NewRequest(http.MethodGet, "/rooms/1/messages", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var resp messageListResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Messages) != 0 {
		t.Fatalf("messages = %+v, want empty slice", resp.Messages)
	}
}

func TestMessagesHandler_InvalidRoomID(t *testing.T) {
	h := NewMessagesHandler(&mockMessageSender{}, &mockMessageLister{})

	req := httptest.NewRequest(http.MethodGet, "/rooms/bad/messages", nil)
	req.SetPathValue("id", "bad")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMessagesHandler_Unauthorized(t *testing.T) {
	h := NewMessagesHandler(&mockMessageSender{}, &mockMessageLister{})

	req := httptest.NewRequest(http.MethodGet, "/rooms/1/messages", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMessagesHandler_List_InvalidQuery(t *testing.T) {
	h := NewMessagesHandler(&mockMessageSender{}, &mockMessageLister{})

	req := httptest.NewRequest(http.MethodGet, "/rooms/1/messages?limit=bad", nil)
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMessagesHandler_Send_InternalError(t *testing.T) {
	h := NewMessagesHandler(
		&mockMessageSender{err: errors.New("db down")},
		&mockMessageLister{},
	)

	body := `{"body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/messages", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestMessagesHandler_Send_ReturnsDecryptedBody(t *testing.T) {
	enc := domain.NewXOREncrypter("handler-test-key")
	rooms := &handlerMessageRoomStore{}
	messages := &handlerMessageStore{}
	svc := service.NewMessageService(rooms, messages, nil, nil, enc)
	h := NewMessagesHandler(svc, svc)

	body := `{"body":"secret message"}`
	req := httptest.NewRequest(http.MethodPost, "/rooms/1/messages", bytes.NewBufferString(body))
	req.SetPathValue("id", "1")
	req = req.WithContext(WithAuthUser(req.Context(), AuthUser{ID: 2, Email: "user2@example.com", Username: "user2"}))
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if messages.savedBody == "secret message" {
		t.Fatal("saved body must be encrypted")
	}

	var item service.MessageItem
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil {
		t.Fatal(err)
	}
	if item.Body != "secret message" {
		t.Fatalf("item.Body = %q, want %q", item.Body, "secret message")
	}
}

type handlerMessageRoomStore struct{}

func (handlerMessageRoomStore) FindDirectRoom(context.Context, int64, int64) (int64, bool, error) {
	panic("not implemented")
}

func (handlerMessageRoomStore) CreateDirectRoom(context.Context, int64, int64) (int64, error) {
	panic("not implemented")
}

func (handlerMessageRoomStore) ListUserRooms(context.Context, int64) ([]repository.RoomListRecord, error) {
	panic("not implemented")
}

func (handlerMessageRoomStore) CreateGroupRoom(context.Context, string, int64, []int64) (int64, error) {
	panic("not implemented")
}

func (handlerMessageRoomStore) IsRoomMember(context.Context, int64, int64) (bool, error) {
	return true, nil
}

type handlerMessageStore struct {
	savedBody string
}

func (s *handlerMessageStore) SaveMessage(_ context.Context, roomID, senderID int64, messageType, body string, _ *int64) (repository.MessageRecord, error) {
	s.savedBody = body
	return repository.MessageRecord{
		ID: 10, RoomID: roomID, SenderID: senderID, Type: messageType, Body: body, CreatedAt: time.Unix(1, 0),
	}, nil
}

func (handlerMessageStore) ListMessages(context.Context, int64, int, int64) ([]repository.MessageRecord, error) {
	panic("not implemented")
}

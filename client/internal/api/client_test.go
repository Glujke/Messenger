package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Login_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/auth/login" {
			t.Errorf("expected /auth/login, got %s", r.URL.Path)
		}

		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		if req["email"] != "test@example.com" || req["password"] != "password123" {
			t.Errorf("unexpected request body: %+v", req)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"token": "fake-jwt-token",
		})
	}))
	defer server.Close()

	client := New(server.URL)
	token, err := client.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token != "fake-jwt-token" {
		t.Fatalf("token = %q, want %q", token, "fake-jwt-token")
	}
}

func TestClient_Login_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid credentials",
		})
	}))
	defer server.Close()

	client := New(server.URL)
	_, err := client.Login(context.Background(), "test@example.com", "wrong")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "invalid credentials" {
		t.Fatalf("error = %q, want %q", err.Error(), "invalid credentials")
	}
}

func TestClient_Register_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/register" {
			t.Errorf("expected /auth/register, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       1,
			"email":    "new@example.com",
			"username": "newuser",
		})
	}))
	defer server.Close()

	client := New(server.URL)
	err := client.Register(context.Background(), "new@example.com", "newuser", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
}

func TestClient_GetRooms_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			t.Errorf("expected Bearer fake-token, got %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"rooms": []map[string]interface{}{
				{"id": 1, "kind": "direct", "peer_id": 2, "peer_email": "friend@example.com"},
				{"id": 2, "kind": "group", "name": "Project X"},
			},
		})
	}))
	defer server.Close()

	client := New(server.URL)
	rooms, err := client.GetRooms(context.Background(), "fake-token")
	if err != nil {
		t.Fatalf("GetRooms() error = %v", err)
	}
	if len(rooms) != 2 {
		t.Fatalf("len(rooms) = %d, want 2", len(rooms))
	}
	if rooms[0].PeerEmail != "friend@example.com" || rooms[1].Name != "Project X" {
		t.Fatalf("unexpected rooms: %+v", rooms)
	}
}

func TestClient_GetRooms_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
	}))
	defer server.Close()

	client := New(server.URL)
	_, err := client.GetRooms(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "unauthorized" {
		t.Fatalf("error = %q, want %q", err.Error(), "unauthorized")
	}
}

func TestClient_GetMessages_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []map[string]interface{}{
				{"id": 1, "room_id": 10, "sender_id": 2, "type": "text", "body": "SGVsbG8="}, // "Hello" in base64 (if encrypted)
			},
		})
	}))
	defer server.Close()

	client := New(server.URL)
	messages, err := client.GetMessages(context.Background(), "fake-token", 10, 0, 0)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("len(messages) = %d, want 1", len(messages))
	}
	if messages[0].Body != "SGVsbG8=" {
		t.Fatalf("unexpected body: %q", messages[0].Body)
	}
}

func TestClient_GetMessages_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != "50" {
			t.Fatalf("limit = %q, want 50", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("before_id") != "20" {
			t.Fatalf("before_id = %q, want 20", r.URL.Query().Get("before_id"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": []map[string]interface{}{
				{"id": 15, "room_id": 10, "sender_id": 2, "type": "text", "body": "older"},
			},
		})
	}))
	defer server.Close()

	client := New(server.URL)
	messages, err := client.GetMessages(context.Background(), "fake-token", 10, 50, 20)
	if err != nil {
		t.Fatalf("GetMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != 15 {
		t.Fatalf("unexpected messages: %+v", messages)
	}
}

func TestClient_SendMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 100, "room_id": 10, "sender_id": 1, "type": "text", "body": "SGVsbG8=",
		})
	}))
	defer server.Close()

	client := New(server.URL)
	msg, err := client.SendMessage(context.Background(), "fake-token", 10, "SGVsbG8=")
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if msg.ID != 100 {
		t.Fatalf("msg.ID = %d, want 100", msg.ID)
	}
}

func TestClient_Contacts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/contacts/invite":
			var req map[string]string
			json.NewDecoder(r.Body).Decode(&req)
			if req["identifier"] != "friend@example.com" {
				t.Errorf("expected identifier=friend@example.com, got %v", req["identifier"])
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
		case "/contacts/requests":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"requests": []map[string]interface{}{
					{
						"id": 1, "from_user_id": 2, "to_user_id": 1, "status": "pending",
						"peer_email": "friend@example.com", "peer_username": "friend",
						"created_at": "2026-06-25T10:00:00Z",
					},
				},
			})
		case "/contacts/requests/1/accept":
			w.WriteHeader(http.StatusNoContent)
		case "/contacts":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"contacts": []map[string]interface{}{
					{"id": 2, "email": "friend@example.com", "username": "friend"},
				},
			})
		}
	}))
	defer server.Close()

	client := New(server.URL)
	ctx := context.Background()
	token := "fake-token"

	// Invite
	if err := client.InviteContact(ctx, token, "friend@example.com"); err != nil {
		t.Fatalf("InviteContact() error = %v", err)
	}

	// List Requests
	reqs, err := client.GetContactRequests(ctx, token)
	if err != nil {
		t.Fatalf("GetContactRequests() error = %v", err)
	}
	if len(reqs) != 1 || reqs[0].FromUserID != 2 || reqs[0].PeerUsername != "friend" {
		t.Fatalf("unexpected requests: %+v", reqs)
	}

	// Accept
	if err := client.AcceptContact(ctx, token, 1); err != nil {
		t.Fatalf("AcceptContact() error = %v", err)
	}

	// List Contacts
	contacts, err := client.ListContacts(ctx, token)
	if err != nil {
		t.Fatalf("ListContacts() error = %v", err)
	}
	if len(contacts) != 1 || contacts[0].ID != 2 || contacts[0].Username != "friend" || contacts[0].Email != "friend@example.com" {
		t.Fatalf("unexpected contacts: %+v", contacts)
	}
}

func TestClient_CreateDirectRoom_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rooms/direct" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req map[string]int64
		json.NewDecoder(r.Body).Decode(&req)
		if req["user_id"] != 2 {
			t.Fatalf("user_id = %v", req["user_id"])
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 42, "peer_id": 2})
	}))
	defer server.Close()

	client := New(server.URL)
	roomID, err := client.CreateDirectRoom(context.Background(), "fake-token", 2)
	if err != nil {
		t.Fatalf("CreateDirectRoom() error = %v", err)
	}
	if roomID != 42 {
		t.Fatalf("roomID = %d, want 42", roomID)
	}
}

func TestClient_CreateGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rooms/group" {
			t.Errorf("expected /rooms/group, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 500})
	}))
	defer server.Close()

	client := New(server.URL)
	id, err := client.CreateGroup(context.Background(), "fake-token", "Project X", []int64{2, 3})
	if err != nil {
		t.Fatalf("CreateGroup() error = %v", err)
	}
	if id != 500 {
		t.Fatalf("id = %d, want 500", id)
	}
}

func TestClient_UploadAttachment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		// Path is /rooms/10/attachments
		if r.URL.Path != "/rooms/10/attachments" {
			t.Errorf("expected /rooms/10/attachments, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"id": 777})
	}))
	defer server.Close()

	client := New(server.URL)
	data := []byte("fake file content")
	id, err := client.UploadAttachment(context.Background(), "fake-token", 10, "test.txt", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("UploadAttachment() error = %v", err)
	}
	if id != 777 {
		t.Fatalf("id = %d, want 777", id)
	}
}

func TestClient_DownloadAttachment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/attachments/777" {
			t.Errorf("expected /attachments/777, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("file data"))
	}))
	defer server.Close()

	client := New(server.URL)
	reader, err := client.DownloadAttachment(context.Background(), "fake-token", 777)
	if err != nil {
		t.Fatalf("DownloadAttachment() error = %v", err)
	}
	defer reader.Close()

	content, _ := io.ReadAll(reader)
	if string(content) != "file data" {
		t.Fatalf("content = %q, want %q", string(content), "file data")
	}
}

func TestClient_SendAttachmentMessage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rooms/10/messages" {
			t.Errorf("expected /rooms/10/messages, got %s", r.URL.Path)
		}
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		if req["attachment_id"] != float64(777) {
			t.Fatalf("unexpected attachment_id: %v", req["attachment_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 101, "room_id": 10, "sender_id": 1, "type": "image", "body": "caption",
			"attachment": map[string]interface{}{
				"id": 777, "filename": "photo.png", "content_type": "image/png", "size_bytes": 100,
			},
		})
	}))
	defer server.Close()

	client := New(server.URL)
	msg, err := client.SendAttachmentMessage(context.Background(), "fake-token", 10, 777, "caption")
	if err != nil {
		t.Fatalf("SendAttachmentMessage() error = %v", err)
	}
	if msg.Attachment == nil || msg.Attachment.Filename != "photo.png" {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestClient_GetMe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/me" || r.Method != http.MethodGet {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			t.Fatalf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 1, "email": "user@example.com", "username": "user",
		})
	}))
	defer server.Close()

	client := New(server.URL)
	profile, err := client.GetMe(context.Background(), "fake-token")
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if profile.Email != "user@example.com" || profile.Username != "user" {
		t.Fatalf("profile = %+v", profile)
	}
}

func TestClient_UpdateUsername_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/auth/me" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		if req["username"] != "newname" {
			t.Fatalf("username = %q", req["username"])
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 1, "email": "user@example.com", "username": "newname",
		})
	}))
	defer server.Close()

	client := New(server.URL)
	profile, err := client.UpdateUsername(context.Background(), "fake-token", "newname")
	if err != nil {
		t.Fatalf("UpdateUsername() error = %v", err)
	}
	if profile.Username != "newname" {
		t.Fatalf("profile = %+v", profile)
	}
}

func TestClient_ChangePassword_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/auth/password" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		if req["old_password"] != "oldsecret1" || req["new_password"] != "newsecret1" {
			t.Fatalf("unexpected body: %+v", req)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(server.URL)
	if err := client.ChangePassword(context.Background(), "fake-token", "oldsecret1", "newsecret1"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
}

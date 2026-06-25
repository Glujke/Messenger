package api

import (
	"context"
	"encoding/json"
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

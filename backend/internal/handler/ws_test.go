package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"log/slog"

	"github.com/gorilla/websocket"

	"messenger/backend/internal/service"
)

func TestWSHandler_MissingToken(t *testing.T) {
	h := NewWSHandler(service.NewHub(), "secret", nil, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWSHandler_InvalidToken(t *testing.T) {
	h := NewWSHandler(service.NewHub(), "secret", nil, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/ws?token=bad", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestWSHandler_ValidTokenUpgrades(t *testing.T) {
	const secret = "secret"
	token, err := service.IssueToken(1, "user@example.com", "user", secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	h := NewWSHandler(service.NewHub(), secret, nil, slog.Default())

	server := httptest.NewServer(h)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "?token=" + token
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error = %v", err)
	}
	defer conn.Close()

	if resp.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusSwitchingProtocols)
	}
}

func TestWSHandler_WrongMethod(t *testing.T) {
	h := NewWSHandler(service.NewHub(), "secret", nil, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/ws", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

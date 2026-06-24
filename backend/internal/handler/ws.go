package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"messenger/backend/internal/service"
)

const (
	wsReadTimeout  = 60 * time.Second
	wsWriteTimeout = 10 * time.Second
	wsPingInterval = 30 * time.Second
)

// RoomMembershipChecker validates room access for WebSocket subscriptions.
type RoomMembershipChecker interface {
	IsRoomMember(ctx context.Context, roomID, userID int64) (bool, error)
}

// WSHandler serves the WebSocket endpoint.
type WSHandler struct {
	hub       *service.Hub
	jwtSecret string
	rooms     RoomMembershipChecker
	logger    *slog.Logger
	upgrader  websocket.Upgrader
}

// NewWSHandler creates a WebSocket handler.
func NewWSHandler(hub *service.Hub, jwtSecret string, rooms RoomMembershipChecker, logger *slog.Logger) *WSHandler {
	return &WSHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
		rooms:     rooms,
		logger:    logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

// ServeHTTP implements http.Handler.
func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		writeUnauthorized(w)
		return
	}

	claims, err := service.ParseToken(token, h.jwtSecret)
	if err != nil {
		writeUnauthorized(w)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	client := service.NewHubClient(claims.UserID)
	h.hub.Register(client)

	go h.writePump(conn, client)
	h.readPump(r.Context(), conn, client)
}

func (h *WSHandler) readPump(ctx context.Context, conn *websocket.Conn, client *service.HubClient) {
	defer func() {
		h.hub.Unregister(client)
		_ = conn.Close()
	}()

	conn.SetReadLimit(4096)
	_ = conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(wsReadTimeout))
	})

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var event service.ClientEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			h.sendError(conn, "invalid json")
			continue
		}

		switch event.Type {
		case service.ClientEventSubscribe:
			h.handleSubscribe(ctx, conn, client, event.RoomID)
		case service.ClientEventUnsubscribe:
			if event.RoomID > 0 {
				h.hub.Unsubscribe(client, event.RoomID)
			}
		default:
			h.sendError(conn, "unknown event type")
		}
	}
}

func (h *WSHandler) handleSubscribe(ctx context.Context, conn *websocket.Conn, client *service.HubClient, roomID int64) {
	if roomID <= 0 {
		h.sendError(conn, "invalid room id")
		return
	}

	isMember, err := h.rooms.IsRoomMember(ctx, roomID, client.UserID)
	if err != nil {
		h.sendError(conn, "internal error")
		return
	}
	if !isMember {
		h.sendError(conn, "not a room member")
		return
	}

	h.hub.Subscribe(client, roomID)
}

func (h *WSHandler) writePump(conn *websocket.Conn, client *service.HubClient) {
	ticker := time.NewTicker(wsPingInterval)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case payload, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *WSHandler) sendError(conn *websocket.Conn, message string) {
	event := service.ServerEvent{
		Type:  service.ServerEventError,
		Error: message,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	_ = conn.SetWriteDeadline(time.Now().Add(wsWriteTimeout))
	_ = conn.WriteMessage(websocket.TextMessage, payload)
}

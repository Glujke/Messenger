package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

// ServerEvent represents an event received from the server.
type ServerEvent struct {
	Type    string  `json:"type"`
	Message Message `json:"message,omitempty"`
	Error   string  `json:"error,omitempty"`
}

const (
	ServerEventNewMessage = "new_message"
	ServerEventError      = "error"
)

// ClientEvent represents an event sent to the server.
type ClientEvent struct {
	Type   string `json:"type"`
	RoomID int64  `json:"room_id,omitempty"`
}

const (
	ClientEventSubscribe   = "subscribe"
	ClientEventUnsubscribe = "unsubscribe"
)

// WSClient handles the WebSocket connection.
type WSClient struct {
	conn *websocket.Conn
}

// Dial establishes a WebSocket connection.
func Dial(baseURL string, token string) (*WSClient, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	
	wsURL := fmt.Sprintf("%s://%s/ws?token=%s", scheme, u.Host, token)
	
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return nil, err
	}

	return &WSClient{conn: conn}, nil
}

// Subscribe sends a subscription request for a room.
func (c *WSClient) Subscribe(roomID int64) error {
	event := ClientEvent{
		Type:   ClientEventSubscribe,
		RoomID: roomID,
	}
	return c.conn.WriteJSON(event)
}

// ReadEvent reads the next event from the server.
func (c *WSClient) ReadEvent() (ServerEvent, error) {
	_, payload, err := c.conn.ReadMessage()
	if err != nil {
		return ServerEvent{}, err
	}

	var event ServerEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return ServerEvent{}, err
	}

	return event, nil
}

// Close closes the connection.
func (c *WSClient) Close() error {
	return c.conn.Close()
}

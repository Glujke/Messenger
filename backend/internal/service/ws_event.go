package service

// WebSocket client event types.
const (
	ClientEventSubscribe   = "subscribe"
	ClientEventUnsubscribe = "unsubscribe"
)

// WebSocket server event types.
const (
	ServerEventMessageNew = "message.new"
	ServerEventError      = "error"
)

// ClientEvent is a message sent by a WebSocket client.
type ClientEvent struct {
	Type   string `json:"type"`
	RoomID int64  `json:"room_id,omitempty"`
}

// ServerEvent is a message sent by the WebSocket server.
type ServerEvent struct {
	Type    string       `json:"type"`
	RoomID  int64        `json:"room_id,omitempty"`
	Message *MessageItem `json:"message,omitempty"`
	Error   string       `json:"error,omitempty"`
}

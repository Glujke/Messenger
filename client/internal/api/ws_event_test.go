package api

import (
	"encoding/json"
	"testing"
)

func TestServerEvent_ParseMessageNew(t *testing.T) {
	payload := []byte(`{
		"type": "message.new",
		"room_id": 10,
		"message": {
			"id": 1,
			"room_id": 10,
			"sender_id": 2,
			"type": "text",
			"body": "SGVsbG8=",
			"created_at": "2026-06-26T10:00:00Z"
		}
	}`)

	var event ServerEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if event.Type != ServerEventNewMessage {
		t.Fatalf("type = %q, want %q", event.Type, ServerEventNewMessage)
	}
	if event.Message.ID != 1 || event.Message.RoomID != 10 || event.Message.SenderID != 2 {
		t.Fatalf("message = %+v", event.Message)
	}
}

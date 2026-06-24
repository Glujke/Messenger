package service

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHub_BroadcastMessage_DeliversToSubscribedClient(t *testing.T) {
	hub := NewHub()
	client := NewHubClient(1)
	hub.Register(client)
	hub.Subscribe(client, 10)

	item := MessageItem{
		ID:        5,
		RoomID:    10,
		SenderID:  2,
		Type:      "text",
		Body:      "hello",
		CreatedAt: time.Unix(1, 0).UTC().Format(time.RFC3339),
	}
	hub.BroadcastMessage(10, item)

	select {
	case payload := <-client.Send:
		var event ServerEvent
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatal(err)
		}
		if event.Type != ServerEventMessageNew {
			t.Fatalf("type = %q, want %q", event.Type, ServerEventMessageNew)
		}
		if event.Message == nil || event.Message.Body != "hello" {
			t.Fatalf("event = %+v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for broadcast")
	}
}

func TestHub_BroadcastMessage_SkipsUnsubscribedClient(t *testing.T) {
	hub := NewHub()
	client := NewHubClient(1)
	hub.Register(client)

	hub.BroadcastMessage(10, MessageItem{ID: 1, RoomID: 10, Type: "text", Body: "hello"})

	select {
	case <-client.Send:
		t.Fatal("unexpected broadcast to unsubscribed client")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHub_UnregisterRemovesSubscriptions(t *testing.T) {
	hub := NewHub()
	client := NewHubClient(1)
	hub.Register(client)
	hub.Subscribe(client, 10)
	hub.Unregister(client)

	hub.BroadcastMessage(10, MessageItem{ID: 1, RoomID: 10, Type: "text", Body: "hello"})

	select {
	case <-client.Send:
		t.Fatal("unexpected broadcast after unregister")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHub_UnsubscribeStopsDelivery(t *testing.T) {
	hub := NewHub()
	client := NewHubClient(1)
	hub.Register(client)
	hub.Subscribe(client, 10)
	hub.Unsubscribe(client, 10)

	hub.BroadcastMessage(10, MessageItem{ID: 1, RoomID: 10, Type: "text", Body: "hello"})

	select {
	case <-client.Send:
		t.Fatal("unexpected broadcast after unsubscribe")
	case <-time.After(50 * time.Millisecond):
	}
}

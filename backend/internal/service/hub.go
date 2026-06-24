package service

import (
	"encoding/json"
	"sync"
)

// Hub tracks room subscriptions and broadcasts events to connected clients.
type Hub struct {
	mu      sync.RWMutex
	clients map[*HubClient]struct{}
	rooms   map[int64]map[*HubClient]struct{}
}

// HubClient is a WebSocket client registered in the hub.
type HubClient struct {
	UserID int64
	Send   chan []byte
	rooms  map[int64]struct{}
}

// NewHub creates an empty hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*HubClient]struct{}),
		rooms:   make(map[int64]map[*HubClient]struct{}),
	}
}

// Register adds a client to the hub.
func (h *Hub) Register(client *HubClient) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = struct{}{}
}

// Unregister removes a client and all room subscriptions.
func (h *Hub) Unregister(client *HubClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.clients, client)
	for roomID := range client.rooms {
		h.removeClientFromRoomLocked(client, roomID)
	}
	client.rooms = make(map[int64]struct{})
}

// Subscribe adds a client to a room channel.
func (h *Hub) Subscribe(client *HubClient, roomID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.rooms == nil {
		client.rooms = make(map[int64]struct{})
	}
	client.rooms[roomID] = struct{}{}

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*HubClient]struct{})
	}
	h.rooms[roomID][client] = struct{}{}
}

// Unsubscribe removes a client from a room channel.
func (h *Hub) Unsubscribe(client *HubClient, roomID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.removeClientFromRoomLocked(client, roomID)
	delete(client.rooms, roomID)
}

// BroadcastMessage sends a message.new event to room subscribers.
func (h *Hub) BroadcastMessage(roomID int64, item MessageItem) {
	event := ServerEvent{
		Type:    ServerEventMessageNew,
		RoomID:  roomID,
		Message: &item,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.broadcast(roomID, payload)
}

func (h *Hub) broadcast(roomID int64, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.rooms[roomID] {
		select {
		case client.Send <- payload:
		default:
		}
	}
}

func (h *Hub) removeClientFromRoomLocked(client *HubClient, roomID int64) {
	roomClients := h.rooms[roomID]
	if roomClients == nil {
		return
	}
	delete(roomClients, client)
	if len(roomClients) == 0 {
		delete(h.rooms, roomID)
	}
}

// NewHubClient creates a hub client with a buffered send channel.
func NewHubClient(userID int64) *HubClient {
	return &HubClient{
		UserID: userID,
		Send:   make(chan []byte, 16),
		rooms:  make(map[int64]struct{}),
	}
}

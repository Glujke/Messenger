package api

import (
	"log"
	"sync"
	"time"
)

// ConnectionStatus describes the websocket link state.
type ConnectionStatus string

const (
	ConnectionConnected    ConnectionStatus = "connected"
	ConnectionReconnecting ConnectionStatus = "reconnecting"
	ConnectionOffline      ConnectionStatus = "offline"
)

// WSManager maintains a websocket connection with automatic reconnect.
type WSManager struct {
	mu sync.Mutex

	serverURL string
	token     string
	conn      *WSClient

	activeRoomID int64
	status       ConnectionStatus

	OnEvent        func(ServerEvent)
	OnStatusChange func(ConnectionStatus)

	stopCh chan struct{}
	doneCh chan struct{}
}

// NewWSManager creates a websocket manager. Call Start to connect.
func NewWSManager(serverURL, token string) *WSManager {
	return &WSManager{
		serverURL: serverURL,
		token:     token,
		status:    ConnectionOffline,
		stopCh:    make(chan struct{}),
	}
}

// Start connects and begins the read/reconnect loop.
func (m *WSManager) Start() {
	m.mu.Lock()
	if m.doneCh != nil {
		m.mu.Unlock()
		return
	}
	m.doneCh = make(chan struct{})
	m.mu.Unlock()

	go m.run()
}

// Stop closes the connection and ends the reconnect loop.
func (m *WSManager) Stop() {
	m.mu.Lock()
	if m.doneCh == nil {
		m.mu.Unlock()
		return
	}
	close(m.stopCh)
	done := m.doneCh
	conn := m.conn
	m.conn = nil
	m.doneCh = nil
	m.stopCh = make(chan struct{})
	m.mu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
	<-done
	m.setStatus(ConnectionOffline)
}

// SetActiveRoom remembers the room to resubscribe after reconnect.
func (m *WSManager) SetActiveRoom(roomID int64) {
	m.mu.Lock()
	m.activeRoomID = roomID
	conn := m.conn
	m.mu.Unlock()

	if conn != nil && roomID > 0 {
		_ = conn.Subscribe(roomID)
	}
}

// Status returns the current connection status.
func (m *WSManager) Status() ConnectionStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *WSManager) run() {
	defer close(m.doneCh)

	backoff := time.Second
	for {
		select {
		case <-m.stopCh:
			return
		default:
		}

		m.setStatus(ConnectionReconnecting)
		conn, err := Dial(m.serverURL, m.token)
		if err != nil {
			log.Printf("ws: dial error: %v", err)
			if !m.sleep(backoff) {
				return
			}
			backoff = minDuration(backoff*2, 30*time.Second)
			continue
		}

		m.mu.Lock()
		m.conn = conn
		roomID := m.activeRoomID
		m.mu.Unlock()

		m.setStatus(ConnectionConnected)
		backoff = time.Second

		if roomID > 0 {
			_ = conn.Subscribe(roomID)
		}

		for {
			event, err := conn.ReadEvent()
			if err != nil {
				log.Printf("ws: read error: %v", err)
				m.mu.Lock()
				if m.conn == conn {
					m.conn = nil
				}
				m.mu.Unlock()
				_ = conn.Close()
				break
			}

			if m.OnEvent != nil {
				m.OnEvent(event)
			}
		}

		select {
		case <-m.stopCh:
			return
		default:
		}
	}
}

func (m *WSManager) setStatus(status ConnectionStatus) {
	m.mu.Lock()
	if m.status == status {
		m.mu.Unlock()
		return
	}
	m.status = status
	cb := m.OnStatusChange
	m.mu.Unlock()

	if cb != nil {
		cb(status)
	}
}

func (m *WSManager) sleep(d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-m.stopCh:
		return false
	case <-timer.C:
		return true
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

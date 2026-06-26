package state

import (
	"context"
	"fmt"

	"messenger/client/internal/api"
	"messenger/client/internal/storage"
	"sort"
	"sync"

	"fyne.io/fyne/v2"
)

// MessagePageSize matches the backend default page size.
const MessagePageSize = 50

// AppState holds the global application state and dependencies.
type AppState struct {
	API              *api.Client
	WSManager        *api.WSManager
	Encrypter        api.MessageEncrypter
	Prefs            *storage.Prefs
	ServerURL        string
	Token            string
	UserID           int64
	Username         string
	Email            string
	Rooms            []api.Room
	ActiveRoomID     int64
	Messages         []api.Message
	HasMoreMessages  bool
	ScrollToBottomOnUpdate bool
	WSStatus         api.ConnectionStatus
	UnreadCounts     map[int64]int
	unreadMu         sync.Mutex
	Contacts         []api.Contact
	ContactRequests  []api.ContactRequest
	Window           fyne.Window
	App              fyne.App
	AppName          string
	OnLogin          func()
	OnLogout         func()
	OnRoomsUpdate    func()
	OnMessagesUpdate func()
	OnContactsUpdate func()
	OnWSStatusChange func()
	OnUnreadChange   func()
	OnOpenRoom       func(roomID int64)
}

// New creates a new application state.
func New(a fyne.App, w fyne.Window, apiURL, encryptionKey, appName string, prefs *storage.Prefs) *AppState {
	return &AppState{
		API:             api.New(apiURL),
		Encrypter:       api.NewXOREncrypter(encryptionKey),
		Prefs:           prefs,
		ServerURL:       apiURL,
		App:             a,
		AppName:         appName,
		Window:          w,
		Rooms:           []api.Room{},
		Messages:        []api.Message{},
		Contacts:        []api.Contact{},
		ContactRequests: []api.ContactRequest{},
		UnreadCounts:    make(map[int64]int),
		WSStatus:        api.ConnectionOffline,
	}
}

// SetServerURL updates the API base URL and persists it.
func (s *AppState) SetServerURL(url string) {
	s.ServerURL = url
	s.API = api.New(url)
	if s.Prefs != nil {
		s.Prefs.SetServerURL(url)
	}
}

// SaveAuthSession stores or clears credentials based on remember-me choice.
func (s *AppState) SaveAuthSession(token string, remember bool) {
	if s.Prefs == nil {
		return
	}
	s.Prefs.SetRememberMe(remember)
	if remember {
		s.Prefs.SetAuthToken(token)
	} else {
		s.Prefs.SetAuthToken("")
	}
}

// SetRooms updates the room list and triggers callback on the UI thread.
func (s *AppState) SetRooms(rooms []api.Room) {
	s.Rooms = rooms
	if s.OnRoomsUpdate != nil {
		s.OnRoomsUpdate()
	}
}

// SetContacts updates the contact list and triggers callback on the UI thread.
func (s *AppState) SetContacts(contacts []api.Contact) {
	s.Contacts = contacts
	if s.OnContactsUpdate != nil {
		s.OnContactsUpdate()
	}
}

// SetContactRequests updates the contact requests list and triggers callback on the UI thread.
func (s *AppState) SetContactRequests(reqs []api.ContactRequest) {
	s.ContactRequests = reqs
	if s.OnContactsUpdate != nil {
		s.OnContactsUpdate()
	}
}

// SetContactsState updates contacts and requests with a single UI refresh.
func (s *AppState) SetContactsState(contacts []api.Contact, reqs []api.ContactRequest) {
	s.Contacts = contacts
	s.ContactRequests = reqs
	if s.OnContactsUpdate != nil {
		s.OnContactsUpdate()
	}
}

// RefreshContacts loads contacts and requests from the API.
func (s *AppState) RefreshContacts(ctx context.Context) error {
	contacts, err := s.API.ListContacts(ctx, s.Token)
	if err != nil {
		return err
	}
	requests, err := s.API.GetContactRequests(ctx, s.Token)
	if err != nil {
		return err
	}
	s.SetContactsState(contacts, requests)
	return nil
}

// SetMessages replaces messages for the active room and scrolls to the bottom.
func (s *AppState) SetMessages(messages []api.Message) {
	s.ScrollToBottomOnUpdate = true
	decryptMessages(s.Encrypter, messages)
	sortMessagesAsc(messages)
	s.Messages = messages
	if s.OnMessagesUpdate != nil {
		s.OnMessagesUpdate()
	}
}

// PrependMessages adds older messages to the start without scrolling to bottom.
func (s *AppState) PrependMessages(messages []api.Message) {
	if len(messages) == 0 {
		return
	}
	s.ScrollToBottomOnUpdate = false
	decryptMessages(s.Encrypter, messages)

	seen := make(map[int64]struct{}, len(s.Messages)+len(messages))
	for _, msg := range s.Messages {
		seen[msg.ID] = struct{}{}
	}

	newOnes := make([]api.Message, 0, len(messages))
	for _, msg := range messages {
		if _, ok := seen[msg.ID]; ok {
			continue
		}
		newOnes = append(newOnes, msg)
		seen[msg.ID] = struct{}{}
	}
	if len(newOnes) == 0 {
		return
	}
	s.Messages = append(newOnes, s.Messages...)
	sortMessagesAsc(s.Messages)
	if s.OnMessagesUpdate != nil {
		s.OnMessagesUpdate()
	}
}

// AddMessage adds a single message and scrolls to the bottom.
func (s *AppState) AddMessage(msg api.Message) {
	s.ScrollToBottomOnUpdate = true
	decrypted, err := s.Encrypter.Decrypt(msg.Body)
	if err == nil {
		msg.Body = decrypted
	}
	for _, existing := range s.Messages {
		if existing.ID == msg.ID {
			return
		}
	}
	s.Messages = append(s.Messages, msg)
	sortMessagesAsc(s.Messages)
	if s.OnMessagesUpdate != nil {
		s.OnMessagesUpdate()
	}
}

func decryptMessages(enc api.MessageEncrypter, messages []api.Message) {
	for i := range messages {
		decrypted, err := enc.Decrypt(messages[i].Body)
		if err == nil {
			messages[i].Body = decrypted
		}
	}
}

func sortMessagesAsc(messages []api.Message) {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})
}

// IncrementUnread increases the unread counter for a room.
func (s *AppState) IncrementUnread(roomID int64) {
	s.unreadMu.Lock()
	s.UnreadCounts[roomID]++
	s.unreadMu.Unlock()
	if s.OnUnreadChange != nil {
		s.OnUnreadChange()
	}
}

// ClearUnread resets unread count for a room.
func (s *AppState) ClearUnread(roomID int64) {
	s.unreadMu.Lock()
	delete(s.UnreadCounts, roomID)
	s.unreadMu.Unlock()
	if s.OnUnreadChange != nil {
		s.OnUnreadChange()
	}
}

// UnreadCount returns unread messages for a room.
func (s *AppState) UnreadCount(roomID int64) int {
	s.unreadMu.Lock()
	defer s.unreadMu.Unlock()
	return s.UnreadCounts[roomID]
}

// RoomTitle returns a display name for notifications.
func (s *AppState) RoomTitle(roomID int64) string {
	for _, room := range s.Rooms {
		if room.ID != roomID {
			continue
		}
		if room.Kind == "group" {
			return room.Name
		}
		if room.PeerEmail != "" {
			return room.PeerEmail
		}
		return room.Name
	}
	return "Messenger"
}

// NotifyMessage shows a desktop notification for a new message.
func (s *AppState) NotifyMessage(msg api.Message) {
	if s.App == nil {
		return
	}
	preview := msg.Body
	if msg.Attachment != nil && preview == "" {
		preview = "📎 " + msg.Attachment.Filename
	}
	if len(preview) > 120 {
		preview = preview[:117] + "..."
	}
	s.App.SendNotification(&fyne.Notification{
		Title:   s.RoomTitle(msg.RoomID),
		Content: s.SenderLabel(msg.SenderID) + ": " + preview,
	})
}

// SenderLabel returns a display name for a message author.
func (s *AppState) SenderLabel(senderID int64) string {
	if senderID == s.UserID {
		return "Вы"
	}
	for _, c := range s.Contacts {
		if c.ID == senderID {
			if c.Username != "" {
				return c.Username
			}
			return c.Email
		}
	}
	for _, room := range s.Rooms {
		if room.PeerID == senderID {
			if room.PeerEmail != "" {
				return room.PeerEmail
			}
			if room.Name != "" {
				return room.Name
			}
		}
	}
	return fmt.Sprintf("user:%d", senderID)
}

// SetToken saves the JWT token and triggers login callback.
func (s *AppState) SetToken(token string) {
	s.Token = token
	if uid, err := api.ParseUserIDFromJWT(token); err == nil {
		s.UserID = uid
	}
	if s.OnLogin != nil {
		s.OnLogin()
	}
}

// Logout clears session state and triggers logout callback.
func (s *AppState) Logout() {
	if s.WSManager != nil {
		s.WSManager.Stop()
		s.WSManager = nil
	}
	s.Token = ""
	s.UserID = 0
	s.Username = ""
	s.Email = ""
	s.Rooms = nil
	s.Messages = nil
	s.ActiveRoomID = 0
	s.HasMoreMessages = false
	s.unreadMu.Lock()
	s.UnreadCounts = make(map[int64]int)
	s.unreadMu.Unlock()
	s.WSStatus = api.ConnectionOffline
	if s.Prefs != nil {
		s.Prefs.ClearAuthSession()
	}
	if s.OnLogout != nil {
		s.OnLogout()
	}
}

// StartWSManager creates and starts the websocket manager if needed.
func (s *AppState) StartWSManager() {
	if s.WSManager != nil {
		return
	}
	mgr := api.NewWSManager(s.ServerURL, s.Token)
	mgr.OnStatusChange = func(status api.ConnectionStatus) {
		s.WSStatus = status
		if s.OnWSStatusChange != nil {
			s.OnWSStatusChange()
		}
	}
	mgr.OnEvent = func(event api.ServerEvent) {
		if event.Type != api.ServerEventNewMessage {
			return
		}
		msg := event.Message
		decrypted, err := s.Encrypter.Decrypt(msg.Body)
		if err == nil {
			msg.Body = decrypted
		}
		if msg.RoomID == s.ActiveRoomID {
			RunOnUI(func() {
				s.AddMessage(msg)
			})
			return
		}
		RunOnUI(func() {
			s.IncrementUnread(msg.RoomID)
			s.NotifyMessage(msg)
		})
	}
	s.WSManager = mgr
	mgr.Start()
}

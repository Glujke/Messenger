package state

import (
	"messenger/client/internal/api"
	"messenger/client/internal/storage"
	"sort"

	"fyne.io/fyne/v2"
)

// AppState holds the global application state and dependencies.
type AppState struct {
	API              *api.Client
	WS               *api.WSClient
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

// SetMessages updates the message list, decrypts them, and triggers callback on the UI thread.
func (s *AppState) SetMessages(messages []api.Message) {
	for i := range messages {
		decrypted, err := s.Encrypter.Decrypt(messages[i].Body)
		if err == nil {
			messages[i].Body = decrypted
		}
	}
	sortMessagesAsc(messages)
	s.Messages = messages
	if s.OnMessagesUpdate != nil {
		s.OnMessagesUpdate()
	}
}

// AddMessage adds a single message, decrypts it, and triggers callback on the UI thread.
func (s *AppState) AddMessage(msg api.Message) {
	decrypted, err := s.Encrypter.Decrypt(msg.Body)
	if err == nil {
		msg.Body = decrypted
	}
	s.Messages = append(s.Messages, msg)
	sortMessagesAsc(s.Messages)
	if s.OnMessagesUpdate != nil {
		s.OnMessagesUpdate()
	}
}

func sortMessagesAsc(messages []api.Message) {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})
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
	if s.WS != nil {
		_ = s.WS.Close()
		s.WS = nil
	}
	s.Token = ""
	s.UserID = 0
	s.Username = ""
	s.Email = ""
	s.Rooms = nil
	s.Messages = nil
	s.ActiveRoomID = 0
	if s.Prefs != nil {
		s.Prefs.ClearAuthSession()
	}
	if s.OnLogout != nil {
		s.OnLogout()
	}
}

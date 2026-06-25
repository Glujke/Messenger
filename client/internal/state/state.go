package state

import (
	"messenger/client/internal/api"

	"fyne.io/fyne/v2"
)

// AppState holds the global application state and dependencies.
type AppState struct {
	API          *api.Client
	Token        string
	Rooms        []api.Room
	ActiveRoomID int64
	Window       fyne.Window
	App          fyne.App
	OnLogin      func()
	OnLogout     func()
	OnRoomsUpdate func()
}

// New creates a new application state.
func New(a fyne.App, w fyne.Window, apiURL string) *AppState {
	return &AppState{
		API:   api.New(apiURL),
		App:   a,
		Window: w,
		Rooms: []api.Room{},
	}
}

// SetRooms updates the room list and triggers callback.
func (s *AppState) SetRooms(rooms []api.Room) {
	s.Rooms = rooms
	if s.OnRoomsUpdate != nil {
		s.OnRoomsUpdate()
	}
}

// SetToken saves the JWT token and triggers login callback.
func (s *AppState) SetToken(token string) {
	s.Token = token
	if s.OnLogin != nil {
		s.OnLogin()
	}
}

// Logout clears the token and triggers logout callback.
func (s *AppState) Logout() {
	s.Token = ""
	if s.OnLogout != nil {
		s.OnLogout()
	}
}

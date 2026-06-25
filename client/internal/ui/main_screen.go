package ui

import (
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

// MainScreen is the primary application interface.
type MainScreen struct {
	state    *state.AppState
	sidebar  *Sidebar
	chatArea *ChatArea
	split    *container.Split
}

// NewMainScreen creates a new main screen.
func NewMainScreen(s *state.AppState) *MainScreen {
	ms := &MainScreen{
		state:    s,
		sidebar:  NewSidebar(s),
		chatArea: NewChatArea(s),
	}

	// Update chat area when active room changes
	s.OnRoomsUpdate = func() {
		ms.sidebar.list.Refresh()
	}

	// We need a way to refresh the chat area when the room selection changes
	// Let's add a callback to AppState for this or just refresh the split
	return ms
}

// Content returns the main screen layout.
func (s *MainScreen) Content() fyne.CanvasObject {
	s.split = container.NewHSplit(
		s.sidebar.Content(),
		s.chatArea.Content(),
	)
	s.split.Offset = 0.3

	// Re-wrap the OnRoomsUpdate to also refresh sidebar
	oldOnRoomsUpdate := s.state.OnRoomsUpdate
	s.state.OnRoomsUpdate = func() {
		if oldOnRoomsUpdate != nil {
			oldOnRoomsUpdate()
		}
	}

	// When messages update, if we were in placeholder mode, we might need to refresh content
	oldOnMessagesUpdate := s.state.OnMessagesUpdate
	s.state.OnMessagesUpdate = func() {
		if oldOnMessagesUpdate != nil {
			oldOnMessagesUpdate()
		}
		// If the right side was a placeholder, replace it with the actual chat area
		s.split.Trailing = s.chatArea.Content()
		s.split.Refresh()
	}

	return s.split
}

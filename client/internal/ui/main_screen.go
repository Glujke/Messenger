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

	ms.sidebar.onRoomChanged = ms.refreshChatPanel

	s.OnRoomsUpdate = func() {
		ms.sidebar.list.Refresh()
	}
	s.OnOpenRoom = func(roomID int64) {
		ms.sidebar.OpenRoom(roomID)
	}

	return ms
}

func (ms *MainScreen) refreshChatPanel() {
	if ms.split == nil {
		return
	}
	ms.split.Trailing = ms.chatArea.Content()
	ms.split.Refresh()
}

// EnableInteraction turns on sidebar room selection after the window content is mounted.
func (ms *MainScreen) EnableInteraction() {
	ms.sidebar.EnableSelection()
}

// Content returns the main screen layout.
func (s *MainScreen) Content() fyne.CanvasObject {
	s.split = container.NewHSplit(
		s.sidebar.Content(),
		s.chatArea.Content(),
	)
	s.split.Offset = 0.3
	return s.split
}

package ui

import (
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MainScreen is the primary application interface.
type MainScreen struct {
	state   *state.AppState
	sidebar *Sidebar
}

// NewMainScreen creates a new main screen.
func NewMainScreen(s *state.AppState) *MainScreen {
	return &MainScreen{
		state:   s,
		sidebar: NewSidebar(s),
	}
}

// Content returns the main screen layout.
func (s *MainScreen) Content() fyne.CanvasObject {
	chatPlaceholder := container.NewCenter(widget.NewLabel("Select a chat to start messaging"))
	
	// Split layout: Sidebar left, Chat right
	split := container.NewHSplit(
		s.sidebar.Content(),
		chatPlaceholder,
	)
	split.Offset = 0.3 // Sidebar takes 30%

	return split
}

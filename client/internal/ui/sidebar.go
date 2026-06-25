package ui

import (
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Sidebar is the left panel with room list.
type Sidebar struct {
	state *state.AppState
	list  *widget.List
}

// NewSidebar creates a new sidebar.
func NewSidebar(s *state.AppState) *Sidebar {
	sb := &Sidebar{state: s}
	sb.setupList()
	
	// Refresh list when rooms are updated
	s.OnRoomsUpdate = func() {
		sb.list.Refresh()
	}
	
	return sb
}

func (s *Sidebar) setupList() {
	s.list = widget.NewList(
		func() int {
			return len(s.state.Rooms)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Room Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			room := s.state.Rooms[id]
			label := obj.(*widget.Label)
			if room.Kind == "group" {
				label.SetText("👥 " + room.Name)
			} else {
				label.SetText("👤 " + room.PeerEmail)
			}
		},
	)

	s.list.OnSelected = func(id widget.ListItemID) {
		room := s.state.Rooms[id]
		s.state.ActiveRoomID = room.ID
		// TODO: Trigger chat area update
	}
}

// Content returns the sidebar layout.
func (s *Sidebar) Content() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Chats", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewBorder(container.NewPadded(title), nil, nil, nil, s.list)
}

package ui

import (
	"context"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Sidebar is the left panel with room list.
type Sidebar struct {
	state          *state.AppState
	list           *widget.List
	onRoomChanged  func()
	selectionReady bool
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
		if !s.selectionReady {
			return
		}
		if id < 0 || int(id) >= len(s.state.Rooms) {
			return
		}

		room := s.state.Rooms[id]
		s.state.ActiveRoomID = room.ID
		s.state.Messages = nil

		if s.onRoomChanged != nil {
			s.onRoomChanged()
		}

		go func(roomID int64) {
			if s.state.WS != nil {
				_ = s.state.WS.Subscribe(roomID)
			}

			messages, err := s.state.API.GetMessages(context.Background(), s.state.Token, roomID)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, s.state.Window)
				})
				return
			}
			fyne.Do(func() {
				s.state.SetMessages(messages)
			})
		}(room.ID)
	}
}

// EnableSelection allows room clicks after the main screen is fully mounted.
func (s *Sidebar) EnableSelection() {
	s.selectionReady = true
}

// Content returns the sidebar layout.
func (s *Sidebar) Content() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Chats", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	
	contactsBtn := widget.NewButtonWithIcon("", theme.AccountIcon(), func() {
		ShowContactsDialog(s.state)
	})
	
	addGroupBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		ShowCreateGroupDialog(s.state)
	})
	
	header := container.NewBorder(nil, nil, nil, container.NewHBox(addGroupBtn, contactsBtn), container.NewPadded(title))
	
	return container.NewBorder(header, nil, nil, nil, s.list)
}

package ui

import (
	"context"
	"fmt"
	"messenger/client/internal/api"
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
	statusLabel    *widget.Label
	onRoomChanged  func()
	selectionReady bool
}

// NewSidebar creates a new sidebar.
func NewSidebar(s *state.AppState) *Sidebar {
	sb := &Sidebar{state: s}
	sb.setupList()

	s.OnRoomsUpdate = func() {
		sb.list.Refresh()
	}
	s.OnUnreadChange = func() {
		fyne.Do(func() {
			sb.list.Refresh()
		})
	}
	s.OnWSStatusChange = func() {
		fyne.Do(func() {
			sb.refreshStatus()
		})
	}

	return sb
}

func (s *Sidebar) refreshStatus() {
	if s.statusLabel == nil {
		return
	}
	switch s.state.WSStatus {
	case api.ConnectionConnected:
		s.statusLabel.SetText("● онлайн")
	case api.ConnectionReconnecting:
		s.statusLabel.SetText("● переподключение…")
	default:
		s.statusLabel.SetText("● офлайн")
	}
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
			label.SetText(s.roomLabel(room))
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
		s.state.HasMoreMessages = false
		s.state.ClearUnread(room.ID)

		if s.onRoomChanged != nil {
			s.onRoomChanged()
		}

		go func(roomID int64) {
			if s.state.WSManager != nil {
				s.state.WSManager.SetActiveRoom(roomID)
			}

			messages, err := s.state.API.GetMessages(
				context.Background(),
				s.state.Token,
				roomID,
				state.MessagePageSize,
				0,
			)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, s.state.Window)
				})
				return
			}
			fyne.Do(func() {
				s.state.HasMoreMessages = len(messages) >= state.MessagePageSize
				s.state.SetMessages(messages)
			})
		}(room.ID)
	}
}

func (s *Sidebar) roomLabel(room api.Room) string {
	var name string
	if room.Kind == "group" {
		name = "👥 " + room.Name
	} else {
		name = "👤 " + room.PeerEmail
	}
	if unread := s.state.UnreadCount(room.ID); unread > 0 {
		return fmt.Sprintf("%s (%d)", name, unread)
	}
	return name
}

// EnableSelection allows room clicks after the main screen is fully mounted.
func (s *Sidebar) EnableSelection() {
	s.selectionReady = true
}

// OpenRoom selects and opens a room by ID.
func (s *Sidebar) OpenRoom(roomID int64) {
	if !s.selectionReady {
		return
	}
	for i, room := range s.state.Rooms {
		if room.ID == roomID {
			s.list.Select(widget.ListItemID(i))
			return
		}
	}
}

// Content returns the sidebar layout.
func (s *Sidebar) Content() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Chats", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	s.statusLabel = widget.NewLabel("● офлайн")
	s.refreshStatus()

	contactsBtn := widget.NewButtonWithIcon("", theme.AccountIcon(), func() {
		ShowContactsDialog(s.state)
	})

	profileBtn := widget.NewButtonWithIcon("", theme.LoginIcon(), func() {
		ShowProfileDialog(s.state)
	})

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		ShowSettingsDialog(s.state)
	})

	addGroupBtn := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		ShowCreateGroupDialog(s.state)
	})

	headerTop := container.NewBorder(nil, nil, title, s.statusLabel)
	header := container.NewBorder(nil, nil, nil, container.NewHBox(addGroupBtn, contactsBtn, profileBtn, settingsBtn), container.NewPadded(headerTop))

	return container.NewBorder(header, nil, nil, nil, s.list)
}

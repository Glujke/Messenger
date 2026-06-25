package ui

import (
	"context"
	"fmt"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowCreateGroupDialog shows a dialog for creating a new group.
func ShowCreateGroupDialog(s *state.AppState) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Group Name (3-64 chars)")

	// Fetch fresh contacts
	go func() {
		contacts, _ := s.API.ListContacts(context.Background(), s.Token)
		s.SetContacts(contacts)
	}()

	selectedIDs := make(map[int64]bool)
	
	contactsList := widget.NewList(
		func() int { return len(s.Contacts) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewCheck("", nil), widget.NewLabel("Contact"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			contact := s.Contacts[id]
			box := obj.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			label := box.Objects[1].(*widget.Label)
			
			label.SetText(contact.Username)
			check.Checked = selectedIDs[contact.ID]
			check.OnChanged = func(checked bool) {
				if checked {
					selectedIDs[contact.ID] = true
				} else {
					delete(selectedIDs, contact.ID)
				}
			}
		},
	)

	form := container.NewBorder(
		container.NewVBox(widget.NewLabel("Group Name:"), nameEntry, widget.NewSeparator(), widget.NewLabel("Select Members:")),
		nil, nil, nil,
		contactsList,
	)

	d := dialog.NewCustomConfirm("Create Group", "Create", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		
		name := nameEntry.Text
		if len(name) < 3 {
			dialog.ShowError(fmt.Errorf("group name too short"), s.Window)
			return
		}

		userIDs := make([]int64, 0, len(selectedIDs))
		for id := range selectedIDs {
			userIDs = append(userIDs, id)
		}

		go func() {
			_, err := s.API.CreateGroup(context.Background(), s.Token, name, userIDs)
			if err != nil {
				dialog.ShowError(err, s.Window)
				return
			}
			
			// Refresh rooms
			rooms, _ := s.API.GetRooms(context.Background(), s.Token)
			s.SetRooms(rooms)
		}()
	}, s.Window)

	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}

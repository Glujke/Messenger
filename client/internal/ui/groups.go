package ui

import (
	"context"
	"fmt"
	"messenger/client/internal/api"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowCreateGroupDialog shows a dialog for creating a new group.
func ShowCreateGroupDialog(s *state.AppState) {
	go func() {
		contacts, err := s.API.ListContacts(context.Background(), s.Token)
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, s.Window)
				return
			}
			showCreateGroupDialog(s, contacts)
		})
	}()
}

func showCreateGroupDialog(s *state.AppState, contacts []api.Contact) {
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Group Name (3-64 chars)")

	selectedIDs := make(map[int64]bool)

	emptyLabel := widget.NewLabel("Нет контактов")
	emptyLabel.Hide()

	contactsList := widget.NewList(
		func() int { return len(contacts) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewCheck("", nil), widget.NewLabel("Contact"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			contact := contacts[id]
			box := obj.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			label := box.Objects[1].(*widget.Label)

			label.SetText(contact.Username + " (" + contact.Email + ")")
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

	if len(contacts) == 0 {
		emptyLabel.Show()
		contactsList.Hide()
	}

	membersBox := container.NewStack(emptyLabel, contactsList)

	form := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Group Name:"),
			nameEntry,
			widget.NewSeparator(),
			widget.NewLabel("Select Members:"),
		),
		nil, nil, nil,
		membersBox,
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
				fyne.Do(func() {
					dialog.ShowError(err, s.Window)
				})
				return
			}

			rooms, err := s.API.GetRooms(context.Background(), s.Token)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, s.Window)
				})
				return
			}
			fyne.Do(func() {
				s.SetRooms(rooms)
			})
		}()
	}, s.Window)

	d.Resize(fyne.NewSize(400, 500))
	d.Show()
}

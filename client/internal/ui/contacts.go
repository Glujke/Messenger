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

// ContactsWindow manages the contacts UI.
type ContactsWindow struct {
	state *state.AppState
}

// ShowContactsDialog shows a dialog for managing contacts.
func ShowContactsDialog(s *state.AppState) {
	cw := &ContactsWindow{state: s}

	content := cw.makeContent()
	d := dialog.NewCustom("Contacts", "Close", content, s.Window)
	d.Resize(fyne.NewSize(500, 400))

	cw.refresh()
	d.Show()
}

func (cw *ContactsWindow) makeContent() fyne.CanvasObject {
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email or Username")
	inviteBtn := widget.NewButtonWithIcon("Invite", theme.ContentAddIcon(), func() {
		if emailEntry.Text == "" {
			return
		}
		go func() {
			err := cw.state.API.InviteContact(context.Background(), cw.state.Token, emailEntry.Text)
			fyne.Do(func() {
				if err != nil {
					dialog.ShowError(err, cw.state.Window)
					return
				}
				emailEntry.SetText("")
				dialog.ShowInformation("Success", "Invitation sent!", cw.state.Window)
				cw.refresh()
			})
		}()
	})
	inviteBox := container.NewBorder(nil, nil, nil, inviteBtn, emailEntry)

	buildRows := func(incoming bool) []api.ContactRequest {
		rows := make([]api.ContactRequest, 0)
		for _, r := range cw.state.ContactRequests {
			isIncoming := r.ToUserID == cw.state.UserID
			if incoming && isIncoming {
				rows = append(rows, r)
			}
			if !incoming && !isIncoming {
				rows = append(rows, r)
			}
		}
		return rows
	}

	incomingRows := []api.ContactRequest{}
	outgoingRows := []api.ContactRequest{}

	incomingList := widget.NewList(
		func() int { return len(incomingRows) },
		func() fyne.CanvasObject {
			row, _, _ := newLabelButtonRow("Accept")
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			req := incomingRows[id]
			box := obj.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			btn := box.Objects[1].(*widget.Button)

			label.SetText(formatRequestLabel(req, true))
			btn.OnTapped = nil
			btn.Enable()
			if req.Status != "pending" {
				btn.Disable()
				btn.SetText("—")
				return
			}
			btn.SetText("Accept")
			btn.OnTapped = func() {
				go func() {
					err := cw.state.API.AcceptContact(context.Background(), cw.state.Token, req.ID)
					fyne.Do(func() {
						if err != nil {
							dialog.ShowError(err, cw.state.Window)
							return
						}
						cw.refresh()
					})
				}()
			}
		},
	)

	outgoingList := widget.NewList(
		func() int { return len(outgoingRows) },
		func() fyne.CanvasObject { return widget.NewLabel("Request") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			req := outgoingRows[id]
			label := obj.(*widget.Label)
			label.SetText(formatRequestLabel(req, false))
		},
	)

	contactsEmpty := widget.NewLabel("Нет контактов")
	contactsEmpty.Hide()

	contactsList := widget.NewList(
		func() int { return len(cw.state.Contacts) },
		func() fyne.CanvasObject {
			row, _, _ := newLabelButtonRow("Написать")
			return row
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			contact := cw.state.Contacts[id]
			box := obj.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			btn := box.Objects[1].(*widget.Button)

			label.SetText(contact.Username + " (" + contact.Email + ")")
			btn.OnTapped = func() {
				openDirectChat(cw.state, contact)
			}
		},
	)

	cw.state.OnContactsUpdate = func() {
		incomingRows = buildRows(true)
		outgoingRows = buildRows(false)
		if len(cw.state.Contacts) == 0 {
			contactsEmpty.Show()
			contactsList.Hide()
		} else {
			contactsEmpty.Hide()
			contactsList.Show()
		}
		incomingList.Refresh()
		outgoingList.Refresh()
		contactsList.Refresh()
	}

	contactsTab := container.NewStack(contactsEmpty, contactsList)

	tabs := container.NewAppTabs(
		container.NewTabItem("Contacts", contactsTab),
		container.NewTabItem("Incoming", incomingList),
		container.NewTabItem("Outgoing", outgoingList),
	)

	return container.NewBorder(
		container.NewVBox(
			widget.NewLabelWithStyle("Add Contact", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			inviteBox,
			widget.NewSeparator(),
		),
		nil, nil, nil,
		tabs,
	)
}

func formatRequestLabel(req api.ContactRequest, incoming bool) string {
	name := req.PeerUsername
	if name == "" {
		name = req.PeerEmail
	}
	if name == "" {
		if incoming {
			name = fmt.Sprint(req.FromUserID)
		} else {
			name = fmt.Sprint(req.ToUserID)
		}
	}
	if incoming {
		return fmt.Sprintf("От %s, status=%s", name, req.Status)
	}
	return fmt.Sprintf("К %s, status=%s", name, req.Status)
}

func (cw *ContactsWindow) refresh() {
	go func() {
		ctx := context.Background()
		contacts, err := cw.state.API.ListContacts(ctx, cw.state.Token)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, cw.state.Window)
			})
			return
		}
		requests, err := cw.state.API.GetContactRequests(ctx, cw.state.Token)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, cw.state.Window)
			})
			return
		}
		fyne.Do(func() {
			cw.state.SetContactsState(contacts, requests)
		})
	}()
}

func openDirectChat(s *state.AppState, contact api.Contact) {
	go func() {
		roomID, err := s.API.CreateDirectRoom(context.Background(), s.Token, contact.ID)
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
			if s.OnOpenRoom != nil {
				s.OnOpenRoom(roomID)
			}
		})
	}()
}

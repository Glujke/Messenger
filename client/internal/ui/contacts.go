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

type requestRow struct {
	Req      api.ContactRequest
	Incoming bool
}

// ShowContactsDialog shows a dialog for managing contacts.
func ShowContactsDialog(s *state.AppState) {
	cw := &ContactsWindow{state: s}
	
	content := cw.makeContent()
	d := dialog.NewCustom("Contacts", "Close", content, s.Window)
	d.Resize(fyne.NewSize(500, 400))
	
	// Initial load
	cw.refresh()
	
	d.Show()
}

func (cw *ContactsWindow) makeContent() fyne.CanvasObject {
	// 1. Invite Section
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email or Username")
	inviteBtn := widget.NewButtonWithIcon("Invite", theme.ContentAddIcon(), func() {
		if emailEntry.Text == "" {
			return
		}
		go func() {
			err := cw.state.API.InviteContact(context.Background(), cw.state.Token, emailEntry.Text)
			if err != nil {
				dialog.ShowError(err, cw.state.Window)
				return
			}
			emailEntry.SetText("")
			dialog.ShowInformation("Success", "Invitation sent!", cw.state.Window)
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

	// 2a. Incoming Requests
	incomingList := widget.NewList(
		func() int { return len(incomingRows) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil, widget.NewButton("Accept", nil), widget.NewLabel("Request"))
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			req := incomingRows[id]
			box := obj.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			btn := box.Objects[1].(*widget.Button)

			label.SetText(
				"From " + fmt.Sprint(req.FromUserID) + " → you, status=" + req.Status,
			)
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
					if err != nil {
						dialog.ShowError(err, cw.state.Window)
						return
					}
					cw.refresh()
				}()
			}
		},
	)

	// 2b. Outgoing Requests
	outgoingList := widget.NewList(
		func() int { return len(outgoingRows) },
		func() fyne.CanvasObject { return widget.NewLabel("Request") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			req := outgoingRows[id]
			label := obj.(*widget.Label)
			label.SetText(
				"You → " + fmt.Sprint(req.ToUserID) + ", status=" + req.Status,
			)
		},
	)

	// 3. Contacts List
	contactsList := widget.NewList(
		func() int { return len(cw.state.Contacts) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Contact Name")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			contact := cw.state.Contacts[id]
			label := obj.(*widget.Label)
			label.SetText(contact.Username + " (" + contact.Email + ")")
		},
	)

	// Update lists when state changes
	cw.state.OnContactsUpdate = func() {
		incomingRows = buildRows(true)
		outgoingRows = buildRows(false)
		incomingList.Refresh()
		outgoingList.Refresh()
		contactsList.Refresh()
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Contacts", contactsList),
		container.NewTabItem("Incoming", incomingList),
		container.NewTabItem("Outgoing", outgoingList),
	)

	return container.NewBorder(container.NewVBox(widget.NewLabelWithStyle("Add Contact", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), inviteBox, widget.NewSeparator()), nil, nil, nil, tabs)
}

func (cw *ContactsWindow) refresh() {
	go func() {
		ctx := context.Background()
		contacts, _ := cw.state.API.ListContacts(ctx, cw.state.Token)
		requests, _ := cw.state.API.GetContactRequests(ctx, cw.state.Token)

		fyne.Do(func() {
			cw.state.SetContacts(contacts)
			cw.state.SetContactRequests(requests)
		})
	}()
}

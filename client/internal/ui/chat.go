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

// ChatArea is the right panel with message history and input.
type ChatArea struct {
	state *state.AppState
	list  *widget.List
	input *widget.Entry
	send  *widget.Button
}

// NewChatArea creates a new chat area.
func NewChatArea(s *state.AppState) *ChatArea {
	ca := &ChatArea{state: s}
	ca.setupList()
	ca.setupInput()

	s.OnMessagesUpdate = func() {
		ca.list.Refresh()
		if len(s.Messages) > 0 {
			ca.list.ScrollToBottom()
		}
	}

	return ca
}

func (ca *ChatArea) setupList() {
	ca.list = widget.NewList(
		func() int {
			return len(ca.state.Messages)
		},
		func() fyne.CanvasObject {
			return container.NewVBox(
				widget.NewLabelWithStyle("Sender", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Message body"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			msg := ca.state.Messages[id]
			box := obj.(*fyne.Container)
			sender := box.Objects[0].(*widget.Label)
			body := box.Objects[1].(*widget.Label)

			sender.SetText(msg.CreatedAt) // TODO: Better sender display
			body.SetText(msg.Body)
			body.Wrapping = fyne.TextWrapWord
		},
	)
}

func (ca *ChatArea) setupInput() {
	ca.input = widget.NewEntry()
	ca.input.SetPlaceHolder("Type a message...")
	ca.input.OnSubmitted = func(text string) {
		ca.sendMessage()
	}

	ca.send = widget.NewButtonWithIcon("", theme.MailSendIcon(), func() {
		ca.sendMessage()
	})
}

func (ca *ChatArea) sendMessage() {
	text := ca.input.Text
	if text == "" || ca.state.ActiveRoomID == 0 {
		return
	}

	ca.input.SetText("")
	ca.input.Disable()
	ca.send.Disable()

	go func() {
		defer ca.input.Enable()
		defer ca.send.Enable()
		defer ca.state.Window.Canvas().Focus(ca.input)

		// Encrypt before sending
		encrypted, err := ca.state.Encrypter.Encrypt(text)
		if err != nil {
			dialog.ShowError(err, ca.state.Window)
			return
		}

		msg, err := ca.state.API.SendMessage(context.Background(), ca.state.Token, ca.state.ActiveRoomID, encrypted)
		if err != nil {
			dialog.ShowError(err, ca.state.Window)
			return
		}

		ca.state.AddMessage(msg)
	}()
}

// Content returns the chat area layout.
func (ca *ChatArea) Content() fyne.CanvasObject {
	if ca.state.ActiveRoomID == 0 {
		return container.NewCenter(widget.NewLabel("Select a chat to start messaging"))
	}

	inputBox := container.NewBorder(nil, nil, nil, ca.send, ca.input)
	return container.NewBorder(nil, inputBox, nil, nil, ca.list)
}

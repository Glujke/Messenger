package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// messageTextRow shows sender, message text and time.
type messageTextRow struct {
	root   fyne.CanvasObject
	sender *widget.Label
	date   *widget.Label
	body   *widget.Label
}

func newMessageTextRow() *messageTextRow {
	sender := widget.NewLabel("")
	sender.TextStyle = fyne.TextStyle{Bold: true}

	body := widget.NewLabel("Message body")
	body.Wrapping = fyne.TextWrapWord

	date := widget.NewLabel("00.00.0000 00:00")
	date.Alignment = fyne.TextAlignTrailing
	date.Wrapping = fyne.TextWrapOff

	dateSlot := container.NewGridWrap(fyne.NewSize(dateColumnWidth, 18), date)

	textRow := container.NewBorder(nil, nil, nil, dateSlot, body)

	return &messageTextRow{
		root: container.NewVBox(
			sender,
			textRow,
		),
		sender: sender,
		date:   date,
		body:   body,
	}
}

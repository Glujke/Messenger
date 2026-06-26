package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// messageTextRow shows message text on the left and time on the right.
type messageTextRow struct {
	root fyne.CanvasObject
	date *widget.Label
	body *widget.Label
}

func newMessageTextRow() *messageTextRow {
	body := widget.NewLabel("Message body")
	body.Wrapping = fyne.TextWrapWord

	date := widget.NewLabel("00.00.0000 00:00")
	date.Alignment = fyne.TextAlignTrailing
	date.Wrapping = fyne.TextWrapOff

	dateSlot := container.NewGridWrap(fyne.NewSize(dateColumnWidth, 18), date)

	return &messageTextRow{
		root: container.NewBorder(nil, nil, nil, dateSlot, body),
		date: date,
		body: body,
	}
}

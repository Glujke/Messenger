package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func newLabelButtonRow(buttonText string) (*fyne.Container, *widget.Label, *widget.Button) {
	label := widget.NewLabel("")
	btn := widget.NewButton(buttonText, nil)
	row := container.NewHBox(label, btn)
	return row, label, btn
}

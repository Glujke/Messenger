package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// messageListItem is a list row with stable widget references for updates.
type messageListItem struct {
	widget.BaseWidget
	textRow  *messageTextRow
	preview  *tappablePreview
	download *widget.Button
	content  *fyne.Container
}

func newMessageListItem() *messageListItem {
	textRow := newMessageTextRow()
	preview := newTappablePreview()
	download := widget.NewButton("Download", nil)

	item := &messageListItem{
		textRow:  textRow,
		preview:  preview,
		download: download,
		content: container.NewVBox(
			textRow.root,
			preview,
			download,
		),
	}
	item.ExtendBaseWidget(item)
	return item
}

func (m *messageListItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(m.content)
}

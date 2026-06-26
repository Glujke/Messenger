package ui

import (
	"errors"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var errEmptyServerURL = errors.New("укажите адрес сервера")

// ShowSettingsDialog opens server URL settings.
func ShowSettingsDialog(s *state.AppState) {
	serverEntry := widget.NewEntry()
	serverEntry.SetText(s.ServerURL)
	serverEntry.SetPlaceHolder("http://itc05:8080")

	content := container.NewVBox(
		widget.NewLabel("Адрес сервера"),
		serverEntry,
		widget.NewSeparator(),
		widget.NewButton("Выйти", func() {
			if s.Token != "" {
				s.Logout()
			}
		}),
	)

	d := dialog.NewCustomConfirm("Настройки", "Сохранить", "Отмена", content, func(save bool) {
		if !save {
			return
		}
		url := serverEntry.Text
		if url == "" {
			dialog.ShowError(errEmptyServerURL, s.Window)
			return
		}
		s.SetServerURL(url)
		dialog.ShowInformation("Настройки", "Адрес сервера сохранён. Изменения применятся при следующем входе.", s.Window)
	}, s.Window)
	d.Resize(fyne.NewSize(420, 200))
	d.Show()
}

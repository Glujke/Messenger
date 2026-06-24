package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"messenger/client/internal/api"
	"messenger/client/internal/config"
)

func main() {
	cfg := config.Default()

	application := app.New()
	window := application.NewWindow("Messenger")
	window.Resize(fyne.NewSize(480, 320))

	serverEntry := widget.NewEntry()
	serverEntry.SetText(cfg.ServerURL)

	statusLabel := widget.NewLabel("Готов к подключению")

	checkButton := widget.NewButton("Проверить API", func() {
		client := api.New(serverEntry.Text)
		body, err := client.Health(context.Background())
		if err != nil {
			statusLabel.SetText("Ошибка: " + err.Error())
			return
		}
		statusLabel.SetText("Ответ: " + body)
	})

	window.SetContent(container.NewVBox(
		widget.NewLabel("Адрес сервера"),
		serverEntry,
		checkButton,
		statusLabel,
	))

	window.ShowAndRun()
}

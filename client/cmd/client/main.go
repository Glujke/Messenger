package main

import (
	"context"
	"messenger/client/internal/state"
	"messenger/client/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

func main() {
	a := app.New()
	w := a.NewWindow("Messenger")
	w.Resize(fyne.NewSize(400, 500))
	w.SetFixedSize(false)
	w.CenterOnScreen()
	// Set minimum size to prevent UI breaking
	w.SetPadded(true)

	// For now, hardcoded local backend URL
	apiURL := "http://localhost:8080"
	s := state.New(a, w, apiURL)

	// Define what happens after login
	s.OnLogin = func() {
		go func() {
			// 1. Load rooms
			rooms, err := s.API.GetRooms(context.Background(), s.Token)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			s.SetRooms(rooms)

			// 2. Switch to main screen
			mainScreen := ui.NewMainScreen(s)
			w.SetContent(mainScreen.Content())
			w.Resize(fyne.NewSize(900, 600))
			w.CenterOnScreen()
		}()
	}

	// Start with login screen
	login := ui.NewLoginScreen(s)
	w.SetContent(login.Content())

	w.ShowAndRun()
}

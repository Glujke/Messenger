package main

import (
	"context"
	"log"
	"messenger/client/internal/api"
	"messenger/client/internal/config"
	"messenger/client/internal/state"
	"messenger/client/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

func main() {
	cfg := config.Default()
	a := app.New()
	w := a.NewWindow("Messenger")
	w.Resize(fyne.NewSize(400, 500))
	w.SetFixedSize(false)
	w.CenterOnScreen()
	// Set minimum size to prevent UI breaking
	w.SetPadded(true)

	s := state.New(a, w, cfg.ServerURL, cfg.EncryptionKey)

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

			// 2. Initialize WebSocket
			ws, err := api.Dial(cfg.ServerURL, s.Token)
			if err != nil {
				log.Printf("WebSocket connection failed: %v", err)
			} else {
				s.WS = ws
				go func() {
					for {
						event, err := ws.ReadEvent()
						if err != nil {
							log.Printf("WS read error: %v", err)
							return
						}
						if event.Type == api.ServerEventNewMessage {
							// Only add if it's for the active room (or we can handle background updates later)
							if event.Message.RoomID == s.ActiveRoomID {
								s.AddMessage(event.Message)
							}
						}
					}
				}()
			}

			// 3. Switch to main screen
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

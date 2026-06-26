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
	w := a.NewWindow(cfg.AppName)
	w.Resize(fyne.NewSize(400, 500))
	w.SetFixedSize(false)
	w.CenterOnScreen()
	// Set minimum size to prevent UI breaking
	w.SetPadded(true)

	s := state.New(a, w, cfg.ServerURL, cfg.EncryptionKey, cfg.AppName)

	// Define what happens after login
	s.OnLogin = func() {
		go func() {
			log.Printf("post-login: fetching rooms")
			rooms, err := s.API.GetRooms(context.Background(), s.Token)
			if err != nil {
				log.Printf("post-login: rooms error: %v", err)
				fyne.Do(func() {
					dialog.ShowError(err, w)
				})
				return
			}
			log.Printf("post-login: got %d rooms", len(rooms))

			log.Printf("post-login: dialing websocket")
			ws, err := api.Dial(cfg.ServerURL, s.Token)
			if err != nil {
				log.Printf("post-login: websocket error: %v", err)
			} else {
				log.Printf("post-login: websocket connected")
			}

			fyne.Do(func() {
				mainScreen := ui.NewMainScreen(s)
				s.Rooms = rooms
				w.SetContent(mainScreen.Content())

				mainScreen.EnableInteraction()
				if s.OnRoomsUpdate != nil {
					s.OnRoomsUpdate()
				}

				w.Resize(fyne.NewSize(900, 600))
				log.Printf("post-login: UI ready")

				if ws != nil {
					s.WS = ws
					go func() {
						for {
							event, err := ws.ReadEvent()
							if err != nil {
								log.Printf("WS read error: %v", err)
								return
							}
							if event.Type == api.ServerEventNewMessage {
								if event.Message.RoomID == s.ActiveRoomID {
									state.RunOnUI(func() {
										s.AddMessage(event.Message)
									})
								}
							}
						}
					}()
				}
			})
		}()
	}

	// Start with login screen
	login := ui.NewLoginScreen(s)
	w.SetContent(login.Content())

	w.ShowAndRun()
}

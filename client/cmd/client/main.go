package main

import (
	"context"
	"log"
	"messenger/client/internal/api"
	"messenger/client/internal/config"
	"messenger/client/internal/state"
	"messenger/client/internal/storage"
	"messenger/client/internal/ui"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func main() {
	cfg := config.Default()
	a := app.New()
	w := a.NewWindow(cfg.AppName)
	w.Resize(fyne.NewSize(400, 500))
	w.SetFixedSize(false)
	w.CenterOnScreen()
	w.SetPadded(true)

	prefs := storage.NewPrefs(a)
	serverURL := prefs.ServerURL()
	if serverURL == "" {
		serverURL = cfg.ServerURL
	}

	s := state.New(a, w, serverURL, cfg.EncryptionKey, cfg.AppName, prefs)

	showLogin := func() {
		login := ui.NewLoginScreen(s)
		w.SetContent(login.Content())
		w.Resize(fyne.NewSize(400, 500))
	}

	s.OnLogout = func() {
		fyne.Do(func() {
			showLogin()
		})
	}

	s.OnLogin = func() {
		go func() {
			log.Printf("post-login: fetching profile")
			profile, err := s.API.GetMe(context.Background(), s.Token)
			if err != nil {
				log.Printf("post-login: profile error: %v", err)
				s.Prefs.ClearAuthSession()
				fyne.Do(func() {
					dialog.ShowError(err, w)
					showLogin()
				})
				return
			}
			s.Email = profile.Email
			s.Username = profile.Username

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
			ws, err := api.Dial(s.ServerURL, s.Token)
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

	if prefs.RememberMe() && prefs.AuthToken() != "" {
		w.SetContent(widget.NewLabel("Вход..."))
		go func() {
			token := prefs.AuthToken()
			_, err := s.API.GetMe(context.Background(), token)
			if err != nil {
				log.Printf("auto-login failed: %v", err)
				prefs.ClearAuthSession()
				fyne.Do(showLogin)
				return
			}
			s.SetToken(token)
		}()
	} else {
		showLogin()
	}

	w.ShowAndRun()
}

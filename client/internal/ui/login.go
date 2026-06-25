package ui

import (
	"context"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// LoginScreen is the authentication form.
type LoginScreen struct {
	state *state.AppState
}

// NewLoginScreen creates a new login screen.
func NewLoginScreen(s *state.AppState) *LoginScreen {
	return &LoginScreen{state: s}
}

// Content returns the login screen UI layout.
func (s *LoginScreen) Content() fyne.CanvasObject {
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	var loginBtn *widget.Button
	loginBtn = widget.NewButton("Login", func() {
		loginBtn.Disable()
		loginBtn.SetText("Logging in...")
		
		go func() {
			defer func() {
				loginBtn.Enable()
				loginBtn.SetText("Login")
			}()

			token, err := s.state.API.Login(context.Background(), emailEntry.Text, passwordEntry.Text)
			if err != nil {
				dialog.ShowError(err, s.state.Window)
				return
			}
			s.state.SetToken(token)
		}()
	})
	loginBtn.Importance = widget.HighImportance

	registerBtn := widget.NewButton("Create Account", func() {
		// Switch to register screen (logic will be in main/state)
		s.showRegister()
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Welcome to Messenger", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		emailEntry,
		passwordEntry,
		loginBtn,
		widget.NewSeparator(),
		registerBtn,
	)

	return container.NewCenter(container.NewPadded(form))
}

func (s *LoginScreen) showRegister() {
	reg := NewRegisterScreen(s.state)
	s.state.Window.SetContent(reg.Content())
}

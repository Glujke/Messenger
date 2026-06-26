package ui

import (
	"context"
	"log"
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

	rememberCheck := widget.NewCheck("Запомнить меня", nil)

	settingsBtn := widget.NewButton("Настройки", func() {
		ShowSettingsDialog(s.state)
	})

	var loginBtn *widget.Button
	loginBtn = widget.NewButton("Login", func() {
		loginBtn.Disable()
		loginBtn.SetText("Logging in...")

		remember := rememberCheck.Checked
		go func() {
			token, err := s.state.API.Login(context.Background(), emailEntry.Text, passwordEntry.Text)
			if err != nil {
				log.Printf("login: api error: %v", err)
				fyne.Do(func() {
					dialog.ShowError(err, s.state.Window)
					loginBtn.Enable()
					loginBtn.SetText("Login")
				})
				return
			}
			log.Printf("login: success, starting post-login flow")
			s.state.SaveAuthSession(token, remember)
			s.state.SetToken(token)
		}()
	})
	loginBtn.Importance = widget.HighImportance

	registerBtn := widget.NewButton("Create Account", func() {
		s.showRegister()
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Welcome to Messenger", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		emailEntry,
		passwordEntry,
		rememberCheck,
		loginBtn,
		widget.NewSeparator(),
		registerBtn,
		settingsBtn,
	)

	return container.NewCenter(container.NewPadded(form))
}

func (s *LoginScreen) showRegister() {
	reg := NewRegisterScreen(s.state)
	s.state.Window.SetContent(reg.Content())
}

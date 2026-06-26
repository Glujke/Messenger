package ui

import (
	"context"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// RegisterScreen is the account creation form.
type RegisterScreen struct {
	state *state.AppState
}

// NewRegisterScreen creates a new register screen.
func NewRegisterScreen(s *state.AppState) *RegisterScreen {
	return &RegisterScreen{state: s}
}

// Content returns the register screen UI layout.
func (s *RegisterScreen) Content() fyne.CanvasObject {
	emailEntry := widget.NewEntry()
	emailEntry.SetPlaceHolder("Email")

	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Username")

	passwordEntry := widget.NewPasswordEntry()
	passwordEntry.SetPlaceHolder("Password")

	var registerBtn *widget.Button
	registerBtn = widget.NewButton("Register", func() {
		registerBtn.Disable()
		registerBtn.SetText("Registering...")

		go func() {
			err := s.state.API.Register(context.Background(), emailEntry.Text, usernameEntry.Text, passwordEntry.Text)
			if err != nil {
				fyne.Do(func() {
					dialog.ShowError(err, s.state.Window)
					registerBtn.Enable()
					registerBtn.SetText("Register")
				})
				return
			}
			fyne.Do(func() {
				dialog.ShowInformation("Success", "Account created! Please login.", s.state.Window)
				s.showLogin()
			})
		}()
	})
	registerBtn.Importance = widget.HighImportance

	backBtn := widget.NewButton("Back to Login", func() {
		s.showLogin()
	})

	form := container.NewVBox(
		widget.NewLabelWithStyle("Create Account", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		emailEntry,
		usernameEntry,
		passwordEntry,
		registerBtn,
		widget.NewSeparator(),
		backBtn,
	)

	return container.NewCenter(container.NewPadded(form))
}

func (s *RegisterScreen) showLogin() {
	login := NewLoginScreen(s.state)
	s.state.Window.SetContent(login.Content())
}

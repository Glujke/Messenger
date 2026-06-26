package ui

import (
	"context"
	"errors"
	"messenger/client/internal/state"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowProfileDialog opens the user profile editor.
func ShowProfileDialog(s *state.AppState) {
	emailEntry := widget.NewEntry()
	emailEntry.Disable()
	emailEntry.SetText(s.Email)

	usernameEntry := widget.NewEntry()
	usernameEntry.SetText(s.Username)

	oldPasswordEntry := widget.NewPasswordEntry()
	oldPasswordEntry.SetPlaceHolder("Текущий пароль")

	newPasswordEntry := widget.NewPasswordEntry()
	newPasswordEntry.SetPlaceHolder("Новый пароль")

	confirmPasswordEntry := widget.NewPasswordEntry()
	confirmPasswordEntry.SetPlaceHolder("Повтор нового пароля")

	var saveUsernameBtn *widget.Button
	saveUsernameBtn = widget.NewButton("Сохранить имя", func() {
		username := usernameEntry.Text
		if username == "" {
			return
		}
		saveUsernameBtn.Disable()
		go func() {
			profile, err := s.API.UpdateUsername(context.Background(), s.Token, username)
			fyne.Do(func() {
				saveUsernameBtn.Enable()
				if err != nil {
					dialog.ShowError(err, s.Window)
					return
				}
				s.Username = profile.Username
				usernameEntry.SetText(profile.Username)
				dialog.ShowInformation("Профиль", "Имя пользователя обновлено", s.Window)
			})
		}()
	})

	var changePasswordBtn *widget.Button
	changePasswordBtn = widget.NewButton("Сменить пароль", func() {
		if newPasswordEntry.Text != confirmPasswordEntry.Text {
			dialog.ShowError(errors.New("новые пароли не совпадают"), s.Window)
			return
		}
		changePasswordBtn.Disable()
		go func() {
			err := s.API.ChangePassword(context.Background(), s.Token, oldPasswordEntry.Text, newPasswordEntry.Text)
			fyne.Do(func() {
				changePasswordBtn.Enable()
				if err != nil {
					dialog.ShowError(err, s.Window)
					return
				}
				oldPasswordEntry.SetText("")
				newPasswordEntry.SetText("")
				confirmPasswordEntry.SetText("")
				dialog.ShowInformation("Профиль", "Пароль изменён", s.Window)
			})
		}()
	})

	logoutBtn := widget.NewButton("Выйти", func() {
		s.Logout()
	})

	content := container.NewVBox(
		widget.NewLabelWithStyle("Профиль", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Email"),
		emailEntry,
		widget.NewLabel("Имя пользователя"),
		usernameEntry,
		saveUsernameBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Смена пароля", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		oldPasswordEntry,
		newPasswordEntry,
		confirmPasswordEntry,
		changePasswordBtn,
		widget.NewSeparator(),
		logoutBtn,
	)

	d := dialog.NewCustom("Профиль", "Закрыть", content, s.Window)
	d.Resize(fyne.NewSize(420, 480))
	d.Show()

	go func() {
		profile, err := s.API.GetMe(context.Background(), s.Token)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, s.Window)
			})
			return
		}
		fyne.Do(func() {
			s.Email = profile.Email
			s.Username = profile.Username
			emailEntry.SetText(profile.Email)
			usernameEntry.SetText(profile.Username)
		})
	}()
}

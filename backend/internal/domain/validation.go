package domain

import (
	"errors"
	"strings"
)

const minPasswordLength = 8

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
)

// ValidateEmail checks that email is non-empty and looks like an email address.
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" || len(email) > 254 {
		return ErrInvalidEmail
	}

	at := strings.Index(email, "@")
	if at <= 0 || at == len(email)-1 {
		return ErrInvalidEmail
	}

	if strings.Contains(email[at+1:], "@") {
		return ErrInvalidEmail
	}

	return nil
}

// ValidatePassword checks minimum password length.
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return ErrInvalidPassword
	}
	return nil
}

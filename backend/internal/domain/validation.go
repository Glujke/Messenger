package domain

import (
	"errors"
	"strings"
)

const minPasswordLength = 8

var (
	ErrInvalidEmail    = errors.New("invalid email")
	ErrInvalidPassword = errors.New("password must be at least 8 characters")
	ErrInvalidUsername = errors.New("username must be 3-32 characters and contain only letters, numbers, and underscores")

	ErrAlreadyFriends  = errors.New("already friends")
	ErrRequestPending  = errors.New("contact request already pending")
	ErrRequestNotFound = errors.New("contact request not found")
	ErrNotContact      = errors.New("user is not in your contacts")
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

// ValidateUsername checks that username is 3-32 characters and contains only alphanumeric characters or underscores.
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < 3 || len(username) > 32 {
		return ErrInvalidUsername
	}

	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return ErrInvalidUsername
		}
	}

	return nil
}

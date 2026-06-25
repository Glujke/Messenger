package domain

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  error
	}{
		{name: "valid", email: "user@example.com", want: nil},
		{name: "empty", email: "", want: ErrInvalidEmail},
		{name: "no at", email: "userexample.com", want: ErrInvalidEmail},
		{name: "no domain", email: "user@", want: ErrInvalidEmail},
		{name: "no local part", email: "@example.com", want: ErrInvalidEmail},
		{name: "double at", email: "a@b@c.com", want: ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err != tt.want {
				t.Fatalf("ValidateEmail(%q) = %v, want %v", tt.email, err, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     error
	}{
		{name: "valid", password: "secret123", want: nil},
		{name: "too short", password: "short", want: ErrInvalidPassword},
		{name: "empty", password: "", want: ErrInvalidPassword},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != tt.want {
				t.Fatalf("ValidatePassword(%q) = %v, want %v", tt.password, err, tt.want)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		want     error
	}{
		{name: "valid", username: "user123", want: nil},
		{name: "valid with underscore", username: "user_name", want: nil},
		{name: "too short", username: "us", want: ErrInvalidUsername},
		{name: "too long", username: "abcdefghijklmnopqrstuvwxyz123456789", want: ErrInvalidUsername},
		{name: "invalid characters", username: "user-name", want: ErrInvalidUsername},
		{name: "empty", username: "", want: ErrInvalidUsername},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if err != tt.want {
				t.Fatalf("ValidateUsername(%q) = %v, want %v", tt.username, err, tt.want)
			}
		})
	}
}

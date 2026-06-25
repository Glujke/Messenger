package service

import (
	"errors"
	"testing"
	"time"
)

func TestParseToken_Valid(t *testing.T) {
	const secret = "test-secret"

	token, err := IssueToken(42, "user@example.com", "user", secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 42 {
		t.Fatalf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("Email = %q, want user@example.com", claims.Email)
	}
	if claims.Username != "user" {
		t.Fatalf("Username = %q, want user", claims.Username)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, err := IssueToken(1, "user@example.com", "user", "secret-a", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseToken(token, "secret-b")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseToken() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestParseToken_Expired(t *testing.T) {
	token, err := IssueToken(1, "user@example.com", "user", "secret", -time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseToken(token, "secret")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseToken() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestParseToken_Malformed(t *testing.T) {
	_, err := ParseToken("not-a-jwt", "secret")
	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("ParseToken() error = %v, want %v", err, ErrInvalidToken)
	}
}

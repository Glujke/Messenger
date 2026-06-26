package service

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

var ErrWrongPassword = errors.New("invalid current password")

// ProfileService handles profile updates for authenticated users.
type ProfileService struct {
	users repository.UserStore
}

// NewProfileService creates a profile service.
func NewProfileService(users repository.UserStore) *ProfileService {
	return &ProfileService{users: users}
}

// ProfileResult is returned after a successful profile update.
type ProfileResult struct {
	ID       int64
	Email    string
	Username string
}

// UpdateUsername changes the username for the given user.
func (s *ProfileService) UpdateUsername(ctx context.Context, userID int64, username string) (ProfileResult, error) {
	if err := domain.ValidateUsername(username); err != nil {
		return ProfileResult{}, err
	}

	record, err := s.users.UpdateUsername(ctx, userID, strings.ToLower(strings.TrimSpace(username)))
	if err != nil {
		return ProfileResult{}, err
	}

	return ProfileResult{
		ID:       record.ID,
		Email:    record.Email,
		Username: record.Username,
	}, nil
}

// ChangePassword verifies the current password and sets a new one.
func (s *ProfileService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	if err := domain.ValidatePassword(newPassword); err != nil {
		return err
	}

	record, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrWrongPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.users.UpdatePasswordHash(ctx, userID, string(hash))
}

package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotVerified        = errors.New("account not verified")
)

// AuthService handles registration and login use cases.
type AuthService struct {
	users     repository.UserStore
	jwtSecret string
	jwtTTL    time.Duration
}

// NewAuthService creates an authentication service.
func NewAuthService(users repository.UserStore, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		users:     users,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

// RegisterResult is returned after successful registration.
type RegisterResult struct {
	ID       int64
	Email    string
	Username string
}

// Register creates a new unverified user account.
func (s *AuthService) Register(ctx context.Context, email, username, password string) (RegisterResult, error) {
	if err := domain.ValidateEmail(email); err != nil {
		return RegisterResult{}, err
	}
	if err := domain.ValidateUsername(username); err != nil {
		return RegisterResult{}, err
	}
	if err := domain.ValidatePassword(password); err != nil {
		return RegisterResult{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResult{}, err
	}

	record, err := s.users.CreateUser(ctx, strings.ToLower(strings.TrimSpace(email)), strings.ToLower(strings.TrimSpace(username)), string(hash))
	if err != nil {
		return RegisterResult{}, err
	}

	return RegisterResult{
		ID:       record.ID,
		Email:    record.Email,
		Username: record.Username,
	}, nil
}

// Login authenticates a verified user and returns a JWT.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if err := domain.ValidateEmail(email); err != nil {
		return "", ErrInvalidCredentials
	}
	if err := domain.ValidatePassword(password); err != nil {
		return "", ErrInvalidCredentials
	}

	record, err := s.users.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, repository.ErrNotFound) {
		return "", ErrInvalidCredentials
	}
	if err != nil {
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(record.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	if !record.Verified {
		return "", ErrNotVerified
	}

	return IssueToken(record.ID, record.Email, record.Username, s.jwtSecret, s.jwtTTL)
}

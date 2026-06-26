package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

type mockUserStore struct {
	createFn func(ctx context.Context, email, username, passwordHash string) (repository.UserRecord, error)
	findFn   func(ctx context.Context, email string) (repository.UserRecord, error)
}

func (m *mockUserStore) CreateUser(ctx context.Context, email, username, passwordHash string) (repository.UserRecord, error) {
	return m.createFn(ctx, email, username, passwordHash)
}

func (m *mockUserStore) FindByEmail(ctx context.Context, email string) (repository.UserRecord, error) {
	return m.findFn(ctx, email)
}

func (m *mockUserStore) FindByUsername(ctx context.Context, username string) (repository.UserRecord, error) {
	return repository.UserRecord{}, repository.ErrNotFound
}

func (m *mockUserStore) FindByID(context.Context, int64) (repository.UserRecord, error) {
	return repository.UserRecord{}, repository.ErrNotFound
}

func (m *mockUserStore) UpdateUsername(context.Context, int64, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockUserStore) UpdatePasswordHash(context.Context, int64, string) error {
	panic("not implemented")
}

func TestAuthService_Register(t *testing.T) {
	store := &mockUserStore{
		createFn: func(_ context.Context, email, username, passwordHash string) (repository.UserRecord, error) {
			if email != "user@example.com" {
				t.Fatalf("email = %q, want user@example.com", email)
			}
			if username != "user" {
				t.Fatalf("username = %q, want user", username)
			}
			if passwordHash == "" {
				t.Fatal("passwordHash is empty")
			}
			return repository.UserRecord{ID: 1, Email: email, Username: username}, nil
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	result, err := svc.Register(context.Background(), "user@example.com", "user", "secret123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.ID != 1 || result.Email != "user@example.com" || result.Username != "user" {
		t.Fatalf("result = %+v, want id=1 email=user@example.com username=user", result)
	}
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	svc := NewAuthService(&mockUserStore{}, "secret", time.Hour)
	_, err := svc.Register(context.Background(), "bad-email", "user", "secret123")
	if !errors.Is(err, domain.ErrInvalidEmail) {
		t.Fatalf("Register() error = %v, want %v", err, domain.ErrInvalidEmail)
	}
}

func TestAuthService_Register_InvalidUsername(t *testing.T) {
	svc := NewAuthService(&mockUserStore{}, "secret", time.Hour)
	_, err := svc.Register(context.Background(), "user@example.com", "u", "secret123")
	if !errors.Is(err, domain.ErrInvalidUsername) {
		t.Fatalf("Register() error = %v, want %v", err, domain.ErrInvalidUsername)
	}
}

func TestAuthService_Register_EmailTaken(t *testing.T) {
	store := &mockUserStore{
		createFn: func(context.Context, string, string, string) (repository.UserRecord, error) {
			return repository.UserRecord{}, repository.ErrEmailTaken
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	_, err := svc.Register(context.Background(), "user@example.com", "user", "secret123")
	if !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("Register() error = %v, want %v", err, repository.ErrEmailTaken)
	}
}

func TestAuthService_Register_UsernameTaken(t *testing.T) {
	store := &mockUserStore{
		createFn: func(context.Context, string, string, string) (repository.UserRecord, error) {
			return repository.UserRecord{}, repository.ErrUsernameTaken
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	_, err := svc.Register(context.Background(), "user@example.com", "user", "secret123")
	if !errors.Is(err, repository.ErrUsernameTaken) {
		t.Fatalf("Register() error = %v, want %v", err, repository.ErrUsernameTaken)
	}
}

func TestAuthService_Login(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	store := &mockUserStore{
		findFn: func(_ context.Context, email string) (repository.UserRecord, error) {
			return repository.UserRecord{
				ID:           1,
				Email:        email,
				Username:     "user",
				PasswordHash: string(hash),
				Verified:     true,
			}, nil
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	token, err := svc.Login(context.Background(), "user@example.com", "secret123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}
}

func TestAuthService_Login_NotVerified(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	store := &mockUserStore{
		findFn: func(_ context.Context, email string) (repository.UserRecord, error) {
			return repository.UserRecord{
				ID:           1,
				Email:        email,
				Username:     "user",
				PasswordHash: string(hash),
				Verified:     false,
			}, nil
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	_, err = svc.Login(context.Background(), "user@example.com", "secret123")
	if !errors.Is(err, ErrNotVerified) {
		t.Fatalf("Login() error = %v, want %v", err, ErrNotVerified)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	store := &mockUserStore{
		findFn: func(_ context.Context, email string) (repository.UserRecord, error) {
			return repository.UserRecord{
				ID:           1,
				Email:        email,
				Username:     "user",
				PasswordHash: string(hash),
				Verified:     true,
			}, nil
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	_, err = svc.Login(context.Background(), "user@example.com", "wrongpass")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	store := &mockUserStore{
		findFn: func(context.Context, string) (repository.UserRecord, error) {
			return repository.UserRecord{}, repository.ErrNotFound
		},
	}

	svc := NewAuthService(store, "secret", time.Hour)
	_, err := svc.Login(context.Background(), "user@example.com", "secret123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want %v", err, ErrInvalidCredentials)
	}
}

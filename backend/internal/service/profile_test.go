package service

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

type mockProfileUserStore struct {
	findByIDFn        func(ctx context.Context, id int64) (repository.UserRecord, error)
	updateUsernameFn  func(ctx context.Context, id int64, username string) (repository.UserRecord, error)
	updatePasswordFn  func(ctx context.Context, id int64, passwordHash string) error
}

func (m *mockProfileUserStore) CreateUser(context.Context, string, string, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockProfileUserStore) FindByEmail(context.Context, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockProfileUserStore) FindByUsername(context.Context, string) (repository.UserRecord, error) {
	panic("not implemented")
}

func (m *mockProfileUserStore) FindByID(ctx context.Context, id int64) (repository.UserRecord, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockProfileUserStore) UpdateUsername(ctx context.Context, id int64, username string) (repository.UserRecord, error) {
	return m.updateUsernameFn(ctx, id, username)
}

func (m *mockProfileUserStore) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error {
	return m.updatePasswordFn(ctx, id, passwordHash)
}

func TestProfileService_UpdateUsername_Success(t *testing.T) {
	svc := NewProfileService(&mockProfileUserStore{
		updateUsernameFn: func(_ context.Context, id int64, username string) (repository.UserRecord, error) {
			if id != 1 {
				t.Fatalf("id = %d, want 1", id)
			}
			if username != "newname" {
				t.Fatalf("username = %q, want newname", username)
			}
			return repository.UserRecord{ID: 1, Email: "a@b.com", Username: username}, nil
		},
	})

	result, err := svc.UpdateUsername(context.Background(), 1, "newname")
	if err != nil {
		t.Fatalf("UpdateUsername() error = %v", err)
	}
	if result.Username != "newname" {
		t.Fatalf("result = %+v", result)
	}
}

func TestProfileService_UpdateUsername_Invalid(t *testing.T) {
	svc := NewProfileService(&mockProfileUserStore{})
	_, err := svc.UpdateUsername(context.Background(), 1, "ab")
	if !errors.Is(err, domain.ErrInvalidUsername) {
		t.Fatalf("UpdateUsername() error = %v, want %v", err, domain.ErrInvalidUsername)
	}
}

func TestProfileService_ChangePassword_Success(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("oldsecret"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}

	var updatedHash string
	svc := NewProfileService(&mockProfileUserStore{
		findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
			return repository.UserRecord{ID: id, PasswordHash: string(hash)}, nil
		},
		updatePasswordFn: func(_ context.Context, id int64, passwordHash string) error {
			if id != 1 {
				t.Fatalf("id = %d, want 1", id)
			}
			updatedHash = passwordHash
			return nil
		},
	})

	if err := svc.ChangePassword(context.Background(), 1, "oldsecret", "newsecret1"); err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(updatedHash), []byte("newsecret1")); err != nil {
		t.Fatalf("updated hash does not match new password: %v", err)
	}
}

func TestProfileService_ChangePassword_WrongOld(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("oldsecret"), bcrypt.DefaultCost)
	svc := NewProfileService(&mockProfileUserStore{
		findByIDFn: func(_ context.Context, id int64) (repository.UserRecord, error) {
			return repository.UserRecord{ID: id, PasswordHash: string(hash)}, nil
		},
	})

	err := svc.ChangePassword(context.Background(), 1, "wrong", "newsecret1")
	if !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("ChangePassword() error = %v, want %v", err, ErrWrongPassword)
	}
}

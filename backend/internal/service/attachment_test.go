package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

type mockAttachmentStore struct {
	createFn func(ctx context.Context, record repository.AttachmentRecord) (repository.AttachmentRecord, error)
	findFn   func(ctx context.Context, id int64) (repository.AttachmentRecord, error)
	usedFn   func(ctx context.Context, attachmentID int64) (bool, error)
}

func (m *mockAttachmentStore) CreateAttachment(ctx context.Context, record repository.AttachmentRecord) (repository.AttachmentRecord, error) {
	return m.createFn(ctx, record)
}

func (m *mockAttachmentStore) FindAttachment(ctx context.Context, id int64) (repository.AttachmentRecord, error) {
	return m.findFn(ctx, id)
}

func (m *mockAttachmentStore) IsAttachmentUsed(ctx context.Context, attachmentID int64) (bool, error) {
	if m.usedFn == nil {
		return false, nil
	}
	return m.usedFn(ctx, attachmentID)
}

type mockFileStore struct {
	saveFn func(reader io.Reader, originalName string) (string, int64, error)
	openFn func(storageKey string) (io.ReadCloser, error)
}

func (m *mockFileStore) Save(reader io.Reader, originalName string) (string, int64, error) {
	return m.saveFn(reader, originalName)
}

func (m *mockFileStore) Open(storageKey string) (io.ReadCloser, error) {
	return m.openFn(storageKey)
}

func TestAttachmentService_Upload(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return true, nil
		},
	}
	attachments := &mockAttachmentStore{
		createFn: func(_ context.Context, record repository.AttachmentRecord) (repository.AttachmentRecord, error) {
			record.ID = 7
			return record, nil
		},
	}
	files := &mockFileStore{
		saveFn: func(reader io.Reader, _ string) (string, int64, error) {
			data, _ := io.ReadAll(reader)
			return "stored.png", int64(len(data)), nil
		},
	}

	svc := NewAttachmentService(rooms, attachments, files, domain.MaxUploadBytes)
	result, err := svc.Upload(context.Background(), 1, 2, "photo.png", "image/png", bytes.NewBufferString("data"), 4)
	if err != nil {
		t.Fatalf("Upload() error = %v", err)
	}
	if result.ID != 7 || result.MessageType != "image" {
		t.Fatalf("result = %+v", result)
	}
}

func TestAttachmentService_Upload_NotMember(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return false, nil
		},
	}
	svc := NewAttachmentService(rooms, &mockAttachmentStore{}, &mockFileStore{}, domain.MaxUploadBytes)
	_, err := svc.Upload(context.Background(), 1, 2, "photo.png", "image/png", bytes.NewReader(nil), 0)
	if !errors.Is(err, ErrNotRoomMember) {
		t.Fatalf("Upload() error = %v", err)
	}
}

func TestAttachmentService_Open(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return true, nil
		},
	}
	attachments := &mockAttachmentStore{
		findFn: func(context.Context, int64) (repository.AttachmentRecord, error) {
			return repository.AttachmentRecord{
				ID:          7,
				RoomID:      1,
				Filename:    "photo.png",
				ContentType: "image/png",
				SizeBytes:   4,
				StorageKey:  "stored.png",
			}, nil
		},
	}
	files := &mockFileStore{
		openFn: func(string) (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewBufferString("data")), nil
		},
	}

	svc := NewAttachmentService(rooms, attachments, files, domain.MaxUploadBytes)
	result, err := svc.Open(context.Background(), 7, 2)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if result.Filename != "photo.png" {
		t.Fatalf("result = %+v", result)
	}
}

func TestAttachmentService_Open_NotMember(t *testing.T) {
	rooms := &mockRoomStoreWithMember{
		isMemberFn: func(context.Context, int64, int64) (bool, error) {
			return false, nil
		},
	}
	attachments := &mockAttachmentStore{
		findFn: func(context.Context, int64) (repository.AttachmentRecord, error) {
			return repository.AttachmentRecord{RoomID: 1, StorageKey: "x"}, nil
		},
	}
	svc := NewAttachmentService(rooms, attachments, &mockFileStore{}, domain.MaxUploadBytes)
	_, err := svc.Open(context.Background(), 7, 2)
	if !errors.Is(err, ErrNotRoomMember) {
		t.Fatalf("Open() error = %v", err)
	}
}

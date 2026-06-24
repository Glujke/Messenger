package service

import (
	"context"
	"errors"
	"io"
	"mime"
	"strings"

	"messenger/backend/internal/domain"
	"messenger/backend/internal/repository"
)

// FileStore saves and opens uploaded files.
type FileStore interface {
	Save(reader io.Reader, originalName string) (storageKey string, size int64, err error)
	Open(storageKey string) (io.ReadCloser, error)
}

// AttachmentService handles file uploads and downloads.
type AttachmentService struct {
	rooms          repository.RoomStore
	attachments    repository.AttachmentStore
	files          FileStore
	maxUploadBytes int64
}

// NewAttachmentService creates an attachment service.
func NewAttachmentService(
	rooms repository.RoomStore,
	attachments repository.AttachmentStore,
	files FileStore,
	maxUploadBytes int64,
) *AttachmentService {
	return &AttachmentService{
		rooms:          rooms,
		attachments:    attachments,
		files:          files,
		maxUploadBytes: maxUploadBytes,
	}
}

// UploadResult is returned after a successful upload.
type UploadResult struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
	MessageType string `json:"message_type"`
}

// OpenResult is returned for attachment downloads.
type OpenResult struct {
	Filename    string
	ContentType string
	SizeBytes   int64
	Reader      io.ReadCloser
}

// Upload stores a file uploaded by a room member.
func (s *AttachmentService) Upload(
	ctx context.Context,
	roomID, uploaderID int64,
	filename, contentType string,
	reader io.Reader,
	declaredSize int64,
) (UploadResult, error) {
	if declaredSize > s.maxUploadBytes {
		return UploadResult{}, domain.ErrUploadTooLarge
	}

	isMember, err := s.rooms.IsRoomMember(ctx, roomID, uploaderID)
	if err != nil {
		return UploadResult{}, err
	}
	if !isMember {
		return UploadResult{}, ErrNotRoomMember
	}

	filename = domain.SanitizeFilename(filename)
	contentType = normalizeContentType(filename, contentType)

	limited := io.LimitReader(reader, s.maxUploadBytes+1)
	storageKey, written, err := s.files.Save(limited, filename)
	if err != nil {
		return UploadResult{}, err
	}
	if err := domain.ValidateUploadSize(written); err != nil {
		return UploadResult{}, err
	}

	messageType, err := domain.MessageTypeForContentType(contentType)
	if err != nil {
		return UploadResult{}, err
	}

	record, err := s.attachments.CreateAttachment(ctx, repository.AttachmentRecord{
		RoomID:      roomID,
		UploaderID:  uploaderID,
		Filename:    filename,
		ContentType: contentType,
		SizeBytes:   written,
		StorageKey:  storageKey,
	})
	if err != nil {
		return UploadResult{}, err
	}

	return UploadResult{
		ID:          record.ID,
		Filename:    record.Filename,
		ContentType: record.ContentType,
		SizeBytes:   record.SizeBytes,
		MessageType: string(messageType),
	}, nil
}

// Open returns a file for a room member.
func (s *AttachmentService) Open(ctx context.Context, attachmentID, userID int64) (OpenResult, error) {
	record, err := s.attachments.FindAttachment(ctx, attachmentID)
	if errors.Is(err, repository.ErrNotFound) {
		return OpenResult{}, ErrAttachmentNotFound
	}
	if err != nil {
		return OpenResult{}, err
	}

	isMember, err := s.rooms.IsRoomMember(ctx, record.RoomID, userID)
	if err != nil {
		return OpenResult{}, err
	}
	if !isMember {
		return OpenResult{}, ErrNotRoomMember
	}

	reader, err := s.files.Open(record.StorageKey)
	if err != nil {
		return OpenResult{}, err
	}

	return OpenResult{
		Filename:    record.Filename,
		ContentType: record.ContentType,
		SizeBytes:   record.SizeBytes,
		Reader:      reader,
	}, nil
}

func normalizeContentType(filename, contentType string) string {
	contentType = strings.TrimSpace(contentType)
	if contentType != "" && contentType != "application/octet-stream" {
		return contentType
	}
	if detected := mime.TypeByExtension(domain.FileExtension(filename)); detected != "" {
		return detected
	}
	return "application/octet-stream"
}

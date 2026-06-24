package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

const MaxUploadBytes = 20 << 20 // 20 MB

var (
	ErrUploadTooLarge   = errors.New("file exceeds 20 MB limit")
	ErrUploadEmpty      = errors.New("file is empty")
	ErrInvalidFileType  = errors.New("unsupported file type")
	ErrInvalidExtension = errors.New("invalid file extension")
)

// ValidateUploadSize checks file size limits.
func ValidateUploadSize(size int64) error {
	if size <= 0 {
		return ErrUploadEmpty
	}
	if size > MaxUploadBytes {
		return ErrUploadTooLarge
	}
	return nil
}

// MessageTypeForContentType maps a MIME type to image or file message type.
func MessageTypeForContentType(contentType string) (MessageType, error) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	if contentType == "" {
		return MessageTypeFile, nil
	}
	if strings.HasPrefix(contentType, "image/") {
		return MessageTypeImage, nil
	}
	return MessageTypeFile, nil
}

// ValidateCaption checks optional attachment caption length.
func ValidateCaption(body string) error {
	if len(body) > maxTextBodyLength {
		return ErrMessageTooLong
	}
	return nil
}

// SanitizeFilename returns a safe basename for storage metadata.
func SanitizeFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	if name == "" || name == "." || name == ".." {
		return "upload"
	}
	return name
}

// FileExtension returns a lowercase extension including the dot.
func FileExtension(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return ""
	}
	return ext
}

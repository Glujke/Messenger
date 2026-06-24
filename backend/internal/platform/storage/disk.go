package storage

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"messenger/backend/internal/domain"
)

// DiskStore saves uploaded files on the local filesystem.
type DiskStore struct {
	baseDir string
}

// NewDiskStore creates a disk-backed file store.
func NewDiskStore(baseDir string) (*DiskStore, error) {
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		return nil, err
	}
	return &DiskStore{baseDir: baseDir}, nil
}

// Save writes content to a new random file and returns its storage key.
func (s *DiskStore) Save(reader io.Reader, originalName string) (string, int64, error) {
	storageKey, err := newStorageKey(domain.FileExtension(originalName))
	if err != nil {
		return "", 0, err
	}

	path := s.pathForKey(storageKey)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", 0, err
	}

	written, err := io.Copy(file, reader)
	closeErr := file.Close()
	if err != nil {
		_ = os.Remove(path)
		return "", 0, err
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return "", 0, closeErr
	}

	return storageKey, written, nil
}

// Open returns a reader for a stored file.
func (s *DiskStore) Open(storageKey string) (io.ReadCloser, error) {
	return os.Open(s.pathForKey(storageKey))
}

func (s *DiskStore) pathForKey(storageKey string) string {
	return filepath.Join(s.baseDir, filepath.Base(storageKey))
}

func newStorageKey(ext string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate storage key: %w", err)
	}
	return hex.EncodeToString(raw[:]) + ext, nil
}

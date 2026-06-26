package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DownloadsDir returns the system Downloads folder/<appName>, creating it when missing.
func DownloadsDir(appName string) (string, error) {
	base, err := userDownloadsDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// UniqueFilePath picks a non-existing path inside dir for the given filename.
func UniqueFilePath(dir, filename string) string {
	path := filepath.Join(dir, filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, 999, ext))
}

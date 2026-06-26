//go:build !windows

package platform

import (
	"os"
	"path/filepath"
)

func userDownloadsDir() (string, error) {
	if dir := os.Getenv("XDG_DOWNLOAD_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Downloads"), nil
}

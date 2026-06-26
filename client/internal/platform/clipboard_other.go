//go:build !windows

package platform

import (
	"os"
	"strings"
)

func clipboardFiles() ([]string, error) {
	return nil, nil
}

func clipboardImageFile() (string, bool, error) {
	return "", false, nil
}

// ClipboardTextPath returns a file path if clipboard text points to an existing file.
func ClipboardTextPath(text string) (string, bool) {
	text = strings.TrimSpace(text)
	text = strings.Trim(text, "\"")
	if text == "" {
		return "", false
	}
	info, err := os.Stat(text)
	if err != nil || info.IsDir() {
		return "", false
	}
	return text, true
}

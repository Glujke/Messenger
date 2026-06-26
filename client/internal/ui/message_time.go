package ui

import (
	"strings"
	"time"
)

const messageTimeLayout = "02.01.2006 15:04"

func formatMessageTime(createdAt string) string {
	createdAt = strings.TrimSpace(createdAt)
	if createdAt == "" {
		return ""
	}

	if t, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
		return t.Local().Format(messageTimeLayout)
	}
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		return t.Local().Format(messageTimeLayout)
	}

	return createdAt
}

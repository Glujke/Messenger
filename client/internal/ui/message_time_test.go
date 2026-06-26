package ui

import "testing"

func TestFormatMessageTime(t *testing.T) {
	got := formatMessageTime("2026-06-25T08:54:23Z")
	if got == "" || got == "2026-06-25T08:54:23Z" {
		t.Fatalf("expected localized datetime without seconds, got %q", got)
	}
	if len(got) < 16 {
		t.Fatalf("expected date and HH:MM, got %q", got)
	}
}

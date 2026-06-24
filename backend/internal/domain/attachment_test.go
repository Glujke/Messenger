package domain

import "testing"

func TestValidateUploadSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want error
	}{
		{name: "valid", size: 1024, want: nil},
		{name: "empty", size: 0, want: ErrUploadEmpty},
		{name: "too large", size: MaxUploadBytes + 1, want: ErrUploadTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUploadSize(tt.size)
			if err != tt.want {
				t.Fatalf("ValidateUploadSize(%d) = %v, want %v", tt.size, err, tt.want)
			}
		})
	}
}

func TestMessageTypeForContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		want        MessageType
	}{
		{name: "png", contentType: "image/png", want: MessageTypeImage},
		{name: "pdf", contentType: "application/pdf", want: MessageTypeFile},
		{name: "empty", contentType: "", want: MessageTypeFile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MessageTypeForContentType(tt.contentType)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	if got := SanitizeFilename(`..\..\evil.exe`); got != "evil.exe" {
		t.Fatalf("SanitizeFilename() = %q", got)
	}
	if got := SanitizeFilename(""); got != "upload" {
		t.Fatalf("SanitizeFilename() = %q", got)
	}
}

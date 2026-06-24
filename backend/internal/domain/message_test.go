package domain

import "testing"

func TestValidateTextBody(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error
	}{
		{name: "valid", body: "hello", want: nil},
		{name: "trimmed valid", body: "  hello  ", want: nil},
		{name: "empty", body: "", want: ErrEmptyMessageBody},
		{name: "spaces", body: "   ", want: ErrEmptyMessageBody},
		{name: "too long", body: string(make([]byte, maxTextBodyLength+1)), want: ErrMessageTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTextBody(tt.body)
			if err != tt.want {
				t.Fatalf("ValidateTextBody() = %v, want %v", err, tt.want)
			}
		})
	}
}

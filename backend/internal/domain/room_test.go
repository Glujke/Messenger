package domain

import "testing"

func TestValidateDirectPeer(t *testing.T) {
	tests := []struct {
		name     string
		callerID int64
		peerID   int64
		want     error
	}{
		{name: "valid", callerID: 1, peerID: 2, want: nil},
		{name: "self chat", callerID: 1, peerID: 1, want: ErrSelfChat},
		{name: "zero peer", callerID: 1, peerID: 0, want: ErrInvalidPeerID},
		{name: "negative peer", callerID: 1, peerID: -1, want: ErrInvalidPeerID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectPeer(tt.callerID, tt.peerID)
			if err != tt.want {
				t.Fatalf("ValidateDirectPeer(%d, %d) = %v, want %v", tt.callerID, tt.peerID, err, tt.want)
			}
		})
	}
}

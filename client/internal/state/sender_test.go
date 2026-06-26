package state

import (
	"testing"

	"messenger/client/internal/api"
)

func TestSenderLabel(t *testing.T) {
	s := &AppState{
		UserID: 1,
		Username: "me",
		Contacts: []api.Contact{{ID: 2, Username: "friend", Email: "f@local"}},
		Rooms: []api.Room{{ID: 10, Kind: "direct", PeerID: 3, PeerEmail: "peer@local"}},
	}

	if got := s.SenderLabel(1); got != "Вы" {
		t.Fatalf("SenderLabel(1) = %q", got)
	}
	if got := s.SenderLabel(2); got != "friend" {
		t.Fatalf("SenderLabel(2) = %q", got)
	}
	if got := s.SenderLabel(3); got != "peer@local" {
		t.Fatalf("SenderLabel(3) = %q", got)
	}
	if got := s.SenderLabel(99); got != "user:99" {
		t.Fatalf("SenderLabel(99) = %q", got)
	}
}

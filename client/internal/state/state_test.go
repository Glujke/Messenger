package state

import (
	"testing"

	"messenger/client/internal/api"
)

func TestPrependMessages_DedupesAndSorts(t *testing.T) {
	s := &AppState{
		Encrypter: api.NewXOREncrypter("test-key"),
		Messages: []api.Message{
			{ID: 10, Body: "b"},
			{ID: 20, Body: "c"},
		},
	}

	s.PrependMessages([]api.Message{
		{ID: 5, Body: "a"},
		{ID: 10, Body: "dup"},
		{ID: 15, Body: "mid"},
	})

	if len(s.Messages) != 4 {
		t.Fatalf("len = %d, want 4", len(s.Messages))
	}
	if s.Messages[0].ID != 5 || s.Messages[3].ID != 20 {
		t.Fatalf("order = %+v", s.Messages)
	}
	if s.Messages[1].Body != "b" {
		t.Fatalf("duplicate overwrote existing message body")
	}
}

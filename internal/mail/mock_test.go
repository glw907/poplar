package mail

import (
	"context"
	"testing"
)

func TestMockBackend(t *testing.T) {
	b := NewMockBackend()

	t.Run("connect succeeds", func(t *testing.T) {
		if err := b.Connect(context.Background()); err != nil {
			t.Fatalf("Connect: %v", err)
		}
	})

	t.Run("list folders returns expected data", func(t *testing.T) {
		folders, err := b.ListFolders()
		if err != nil {
			t.Fatalf("ListFolders: %v", err)
		}
		if len(folders) == 0 {
			t.Fatal("expected at least one folder")
		}
		if folders[0].Name != "Inbox" {
			t.Errorf("first folder = %q, want Inbox", folders[0].Name)
		}
		if folders[0].Role != "inbox" {
			t.Errorf("Inbox role = %q, want inbox", folders[0].Role)
		}
	})

	t.Run("inbox has unread messages", func(t *testing.T) {
		folders, _ := b.ListFolders()
		inbox := folders[0]
		if inbox.Unseen == 0 {
			t.Error("expected Inbox to have unread messages")
		}
		if inbox.Exists == 0 {
			t.Error("expected Inbox to have messages")
		}
	})

	t.Run("fetch headers returns messages", func(t *testing.T) {
		msgs, err := b.FetchHeaders(nil)
		if err != nil {
			t.Fatalf("FetchHeaders: %v", err)
		}
		if len(msgs) == 0 {
			t.Fatal("expected at least one message")
		}
		for i, m := range msgs {
			if m.Subject == "" {
				t.Errorf("message %d has empty subject", i)
			}
			if m.From == "" {
				t.Errorf("message %d has empty from", i)
			}
		}
	})

	t.Run("updates channel is non-nil", func(t *testing.T) {
		ch := b.Updates()
		if ch == nil {
			t.Fatal("Updates() returned nil channel")
		}
	})

	t.Run("disconnect succeeds", func(t *testing.T) {
		if err := b.Disconnect(); err != nil {
			t.Fatalf("Disconnect: %v", err)
		}
	})
}

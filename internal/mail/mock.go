package mail

import (
	"context"
	"fmt"
	"io"
	"strings"
)

// MockBackend implements Backend with hardcoded data.
// Used for prototype development, testing, and demos.
type MockBackend struct {
	folders []Folder
	msgs    []MessageInfo
	updates chan Update
}

// NewMockBackend creates a MockBackend with realistic sample data.
func NewMockBackend() *MockBackend {
	return &MockBackend{
		folders: []Folder{
			{Name: "Inbox", Exists: 10, Unseen: 3, Role: "inbox"},
			{Name: "Drafts", Exists: 2, Unseen: 0, Role: "drafts"},
			{Name: "Sent", Exists: 145, Unseen: 0, Role: "sent"},
			{Name: "Archive", Exists: 1893, Unseen: 0, Role: "archive"},
			{Name: "Spam", Exists: 12, Unseen: 12, Role: "junk"},
			{Name: "Trash", Exists: 5, Unseen: 0, Role: "trash"},
			{Name: "Notifications", Exists: 47, Unseen: 0, Role: ""},
			{Name: "Remind", Exists: 8, Unseen: 0, Role: ""},
			{Name: "Lists/golang", Exists: 234, Unseen: 0, Role: ""},
			{Name: "Lists/rust", Exists: 89, Unseen: 0, Role: ""},
		},
		msgs: []MessageInfo{
			{UID: "1", Subject: "Re: Project update for Q2 launch", From: "Alice Johnson", Date: "10:23 AM", Flags: 0},
			{UID: "2", Subject: "Quick question about the API", From: "Bob Smith", Date: "9:45 AM", Flags: 0},
			{UID: "3", Subject: "Lunch tomorrow?", From: "Carol White", Date: "9:12 AM", Flags: 0},
			{UID: "4", Subject: "Meeting notes from yesterday", From: "David Chen", Date: "Yesterday", Flags: FlagSeen},
			{UID: "5", Subject: "Invoice #2847 attached", From: "Billing Dept", Date: "Yesterday", Flags: FlagSeen | FlagFlagged},
			{UID: "6", Subject: "Re: Weekend hiking trip", From: "Emma Wilson", Date: "Yesterday", Flags: FlagSeen | FlagAnswered},
			{UID: "7", Subject: "Your subscription renewal", From: "Acme Cloud", Date: "Apr 8", Flags: FlagSeen},
			{UID: "8", Subject: "Code review: auth refactor PR #42", From: "GitHub", Date: "Apr 8", Flags: FlagSeen},
			{UID: "9", Subject: "New comment on your post", From: "Dev Community", Date: "Apr 7", Flags: FlagSeen},
			{UID: "10", Subject: "Flight confirmation: SFO → SEA", From: "Alaska Airlines", Date: "Apr 7", Flags: FlagSeen | FlagFlagged},
		},
		updates: make(chan Update),
	}
}

func (m *MockBackend) Connect(_ context.Context) error { return nil }
func (m *MockBackend) Disconnect() error               { return nil }

// ListFolders returns the hardcoded folder list.
func (m *MockBackend) ListFolders() ([]Folder, error) {
	return m.folders, nil
}

// OpenFolder is a no-op for the mock backend.
func (m *MockBackend) OpenFolder(_ string) error { return nil }

// FetchHeaders returns the hardcoded message list. The uids parameter is
// ignored — the mock always returns all messages.
func (m *MockBackend) FetchHeaders(_ []UID) ([]MessageInfo, error) {
	return m.msgs, nil
}

// FetchBody returns a placeholder body.
func (m *MockBackend) FetchBody(uid UID) (io.Reader, error) {
	return strings.NewReader(fmt.Sprintf("Mock body for message %s", uid)), nil
}

func (m *MockBackend) Search(_ SearchCriteria) ([]UID, error) { return nil, nil }
func (m *MockBackend) Move(_ []UID, _ string) error           { return nil }
func (m *MockBackend) Copy(_ []UID, _ string) error           { return nil }
func (m *MockBackend) Delete(_ []UID) error                   { return nil }
func (m *MockBackend) Flag(_ []UID, _ Flag, _ bool) error     { return nil }
func (m *MockBackend) MarkRead(_ []UID) error                 { return nil }
func (m *MockBackend) MarkAnswered(_ []UID) error             { return nil }

func (m *MockBackend) Send(_ string, _ []string, _ io.Reader) error {
	return nil
}

// Updates returns the update channel. The mock backend never sends updates.
func (m *MockBackend) Updates() <-chan Update {
	return m.updates
}

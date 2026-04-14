package mail

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// MockBackend implements Backend with hardcoded data.
// Used for prototype development, testing, and demos.
type MockBackend struct {
	name    string
	folders []Folder
	msgs    []MessageInfo
	updates chan Update
}

// NewMockBackend creates a MockBackend with realistic sample data.
func NewMockBackend() *MockBackend {
	// Mock timestamps live in time.Local so the hour/minute values
	// below render verbatim for an interactive demo, regardless of
	// the developer's timezone. A CI/golden-file setup that needs
	// reproducible output across timezones should pin a fixed
	// location here instead.
	at := func(month time.Month, day, hour, min int) time.Time {
		return time.Date(2026, month, day, hour, min, 0, 0, time.Local)
	}
	return &MockBackend{
		name: "geoff@907.life",
		folders: []Folder{
			{Name: "Inbox", Exists: 14, Unseen: 4, Role: "inbox"},
			{Name: "Drafts", Exists: 2, Unseen: 0, Role: "drafts"},
			{Name: "Sent", Exists: 145, Unseen: 0, Role: "sent"},
			{Name: "Archive", Exists: 1893, Unseen: 0, Role: "archive"},
			{Name: "Junk", Exists: 12, Unseen: 12, Role: ""},
			{Name: "Trash", Exists: 5, Unseen: 0, Role: "trash"},
			{Name: "Notifications", Exists: 47, Unseen: 0, Role: ""},
			{Name: "Remind", Exists: 8, Unseen: 0, Role: ""},
			{Name: "Lists/golang", Exists: 234, Unseen: 0, Role: ""},
			{Name: "Lists/rust", Exists: 89, Unseen: 0, Role: ""},
		},
		msgs: []MessageInfo{
			// Flat single-message threads: ThreadID == UID, no InReplyTo.
			// Only SentAt is set; the UI formats the display string via
			// formatRelativeDate at render time.
			{UID: "1", ThreadID: "1", Subject: "Re: Project update for Q2 launch", From: "Alice Johnson", SentAt: at(time.April, 13, 10, 23), Flags: 0},
			{UID: "2", ThreadID: "2", Subject: "Quick question about the API", From: "Bob Smith", SentAt: at(time.April, 13, 9, 45), Flags: 0},
			{UID: "3", ThreadID: "3", Subject: "Lunch tomorrow?", From: "Carol White", SentAt: at(time.April, 13, 9, 12), Flags: 0},
			{UID: "4", ThreadID: "4", Subject: "Meeting notes from yesterday", From: "David Chen", SentAt: at(time.April, 12, 15, 47), Flags: FlagSeen},
			{UID: "5", ThreadID: "5", Subject: "Invoice #2847 attached", From: "Billing Dept", SentAt: at(time.April, 12, 11, 32), Flags: FlagSeen | FlagFlagged},
			{UID: "6", ThreadID: "6", Subject: "Re: Weekend hiking trip", From: "Emma Wilson", SentAt: at(time.April, 12, 8, 15), Flags: FlagSeen | FlagAnswered},
			{UID: "7", ThreadID: "7", Subject: "Your subscription renewal", From: "Acme Cloud", SentAt: at(time.April, 8, 16, 22), Flags: FlagSeen},
			{UID: "8", ThreadID: "8", Subject: "Code review: auth refactor PR #42", From: "GitHub", SentAt: at(time.April, 8, 9, 30), Flags: FlagSeen},
			{UID: "9", ThreadID: "9", Subject: "New comment on your post", From: "Dev Community", SentAt: at(time.April, 7, 15, 45), Flags: FlagSeen},
			{UID: "10", ThreadID: "10", Subject: "Flight confirmation: SFO → SEA", From: "Alaska Airlines", SentAt: at(time.April, 7, 10, 15), Flags: FlagSeen | FlagFlagged},

			// Threaded conversation T1: branching shape (root + linear chain + sibling).
			// Exercises the full ├─ │ └─ prefix vocabulary. First child unread so a
			// folded thread can still carry "contains unread" status.
			{UID: "20", ThreadID: "T1", InReplyTo: "", Subject: "Server migration plan", From: "Frank Lee", SentAt: at(time.April, 5, 9, 0), Flags: FlagSeen | FlagAnswered},
			{UID: "21", ThreadID: "T1", InReplyTo: "20", Subject: "Re: Server migration plan", From: "Grace Kim", SentAt: at(time.April, 5, 11, 30), Flags: 0},
			{UID: "22", ThreadID: "T1", InReplyTo: "21", Subject: "Re: Server migration plan", From: "Frank Lee", SentAt: at(time.April, 5, 14, 15), Flags: FlagSeen},
			{UID: "23", ThreadID: "T1", InReplyTo: "20", Subject: "Re: Server migration plan", From: "Henry Park", SentAt: at(time.April, 5, 16, 45), Flags: FlagSeen},
		},
		updates: make(chan Update),
	}
}

func (m *MockBackend) AccountName() string              { return m.name }
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

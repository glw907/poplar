package mail

import (
	"context"
	"io"
)

// SearchCriteria defines message search parameters.
type SearchCriteria struct {
	Header map[string][]string
	Body   []string
	Text   []string
}

// Backend is the interface that mail protocol adapters implement.
// Each method blocks until the operation completes. Bubbletea's
// tea.Cmd model handles async naturally by running blocking calls
// in commands that return messages on completion.
type Backend interface {
	AccountName() string

	Connect(ctx context.Context) error
	Disconnect() error

	ListFolders() ([]Folder, error)
	OpenFolder(name string) error

	FetchHeaders(uids []UID) ([]MessageInfo, error)
	FetchBody(uid UID) (io.Reader, error)

	Search(criteria SearchCriteria) ([]UID, error)

	Move(uids []UID, dest string) error
	Copy(uids []UID, dest string) error
	Delete(uids []UID) error
	Flag(uids []UID, flag Flag, set bool) error
	MarkRead(uids []UID) error
	MarkAnswered(uids []UID) error

	Send(from string, rcpts []string, body io.Reader) error

	Updates() <-chan Update
}

// MessageInfo holds message header information for list display.
type MessageInfo struct {
	UID     UID
	Subject string
	From    string
	Date    string
	Flags   Flag
	Size    uint32
}

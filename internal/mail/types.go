// Package mail defines poplar's mail backend interface and types.
package mail

// UID is a message unique identifier.
type UID string

// Flag represents email message flags.
type Flag uint32

const (
	FlagSeen Flag = 1 << iota
	FlagRecent
	FlagAnswered
	FlagForwarded
	FlagDeleted
	FlagFlagged
	FlagDraft
)

// Folder represents a mail folder with summary counts.
type Folder struct {
	Name   string
	Exists int
	Unseen int
	Role   string
}

// UpdateType classifies asynchronous backend updates.
type UpdateType int

const (
	UpdateNewMail UpdateType = iota
	UpdateFlagsChanged
	UpdateExpunge
	UpdateFolderInfo
)

// Update represents an asynchronous update from the backend.
type Update struct {
	Type   UpdateType
	Folder string
	UIDs   []UID
}

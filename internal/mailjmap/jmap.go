// Package mailjmap implements mail.Backend for JMAP servers
// (Fastmail) by calling git.sr.ht/~rockorager/go-jmap directly.
// All RPC methods are synchronous. A single goroutine owned by
// Connect/Disconnect reads JMAP push StateChange events and emits
// mail.Update values onto the channel returned by Updates().
package mailjmap

import (
	"context"
	"errors"
	"io"
	"sync"

	"git.sr.ht/~rockorager/go-jmap"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/glw907/poplar/internal/config"
	"github.com/glw907/poplar/internal/mail"
)

// Backend is one JMAP account. Construct with New, drive lifecycle
// with Connect/Disconnect.
type Backend struct {
	cfg config.AccountConfig

	mu      sync.Mutex
	client  *jmap.Client
	session *jmap.Session
	current string
	folders map[string]folderEntry
	blobIDs map[mail.UID]string
	states  map[string]string

	bodies  *lru.Cache[string, []byte]
	updates chan mail.Update

	pushCancel context.CancelFunc
	pushDone   chan struct{}
}

type folderEntry struct {
	id     string // JMAP mailbox id
	folder mail.Folder
}

// New constructs an unconnected Backend for cfg. cfg.Source is the
// JMAP session URL (e.g. https://api.fastmail.com/jmap/session);
// cfg.Password (post env-var substitution) supplies the bearer token.
func New(cfg config.AccountConfig) *Backend {
	return &Backend{
		cfg:     cfg,
		folders: make(map[string]folderEntry),
		blobIDs: make(map[mail.UID]string),
		states:  make(map[string]string),
	}
}

// AccountName satisfies mail.Backend.
func (b *Backend) AccountName() string { return b.cfg.Name }

// Updates satisfies mail.Backend. Returns a nil channel before
// Connect succeeds.
func (b *Backend) Updates() <-chan mail.Update { return b.updates }

// Connect satisfies mail.Backend.
func (b *Backend) Connect(_ context.Context) error {
	return errors.New("not implemented")
}

// Disconnect satisfies mail.Backend.
func (b *Backend) Disconnect() error {
	return errors.New("not implemented")
}

// ListFolders satisfies mail.Backend.
func (b *Backend) ListFolders() ([]mail.Folder, error) {
	return nil, errors.New("not implemented")
}

// OpenFolder satisfies mail.Backend.
func (b *Backend) OpenFolder(_ string) error {
	return errors.New("not implemented")
}

// QueryFolder satisfies mail.Backend.
func (b *Backend) QueryFolder(_ string, _, _ int) ([]mail.UID, int, error) {
	return nil, 0, errors.New("not implemented")
}

// FetchHeaders satisfies mail.Backend.
func (b *Backend) FetchHeaders(_ []mail.UID) ([]mail.MessageInfo, error) {
	return nil, errors.New("not implemented")
}

// FetchBody satisfies mail.Backend.
func (b *Backend) FetchBody(_ mail.UID) (io.Reader, error) {
	return nil, errors.New("not implemented")
}

// Search satisfies mail.Backend.
func (b *Backend) Search(_ mail.SearchCriteria) ([]mail.UID, error) {
	return nil, errors.New("not implemented")
}

// Move satisfies mail.Backend.
func (b *Backend) Move(_ []mail.UID, _ string) error {
	return errors.New("not implemented")
}

// Copy satisfies mail.Backend.
func (b *Backend) Copy(_ []mail.UID, _ string) error {
	return errors.New("not implemented")
}

// Delete satisfies mail.Backend.
func (b *Backend) Delete(_ []mail.UID) error {
	return errors.New("not implemented")
}

// Flag satisfies mail.Backend.
func (b *Backend) Flag(_ []mail.UID, _ mail.Flag, _ bool) error {
	return errors.New("not implemented")
}

// MarkRead satisfies mail.Backend.
func (b *Backend) MarkRead(_ []mail.UID) error {
	return errors.New("not implemented")
}

// MarkAnswered satisfies mail.Backend.
func (b *Backend) MarkAnswered(_ []mail.UID) error {
	return errors.New("not implemented")
}

// Send satisfies mail.Backend.
func (b *Backend) Send(_ string, _ []string, _ io.Reader) error {
	return errors.New("not implemented")
}

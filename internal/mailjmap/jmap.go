// Package mailjmap implements mail.Backend for JMAP servers
// (Fastmail) by calling git.sr.ht/~rockorager/go-jmap directly.
// All RPC methods are synchronous. A single goroutine owned by
// Connect/Disconnect reads JMAP push StateChange events and emits
// mail.Update values onto the channel returned by Updates().
package mailjmap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"git.sr.ht/~rockorager/go-jmap"
	jmapmail "git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	lru "github.com/hashicorp/golang-lru/v2"

	"github.com/glw907/poplar/internal/config"
	"github.com/glw907/poplar/internal/mail"
)

// jmapClient is the subset of *jmap.Client poplar uses. The real
// *jmap.Client satisfies it; tests substitute a fake.
type jmapClient interface {
	Do(req *jmap.Request) (*jmap.Response, error)
}

// Backend is one JMAP account. Construct with New, drive lifecycle
// with Connect/Disconnect.
type Backend struct {
	cfg config.AccountConfig

	mu      sync.Mutex
	client  jmapClient
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

// NewWithClient is for tests. It bypasses the network handshake and
// installs a pre-built client that already satisfies the session
// contract. The caller is responsible for populating b.session if any
// method under test reads PrimaryAccounts — assign it directly:
//
//	b := NewWithClient(cfg, fake)
//	b.session = &jmap.Session{...}
func NewWithClient(cfg config.AccountConfig, c jmapClient) *Backend {
	b := New(cfg)
	b.client = c
	cache, _ := lru.New[string, []byte](bodyCacheSize)
	b.bodies = cache
	b.updates = make(chan mail.Update, updatesBuffer)
	return b
}

// AccountName satisfies mail.Backend.
func (b *Backend) AccountName() string { return b.cfg.Name }

// Updates satisfies mail.Backend. Returns a nil channel before
// Connect succeeds.
func (b *Backend) Updates() <-chan mail.Update { return b.updates }

const (
	bodyCacheSize = 64
	updatesBuffer = 64
)

// Connect satisfies mail.Backend. It authenticates against the JMAP
// session endpoint, populates the folder map, and initialises the
// body cache and updates channel.
func (b *Backend) Connect(_ context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	cli := &jmap.Client{
		SessionEndpoint: b.cfg.Source,
	}
	cli.WithAccessToken(b.cfg.Password)
	if err := cli.Authenticate(); err != nil {
		return fmt.Errorf("connect: authenticate: %w", err)
	}
	b.client = cli
	b.session = cli.Session

	if err := b.refreshFoldersLocked(); err != nil {
		return fmt.Errorf("connect: list folders: %w", err)
	}

	cache, err := lru.New[string, []byte](bodyCacheSize)
	if err != nil {
		return fmt.Errorf("connect: init body cache: %w", err)
	}
	b.bodies = cache
	b.updates = make(chan mail.Update, updatesBuffer)

	// Push loop wired in Task 13.

	return nil
}

// refreshFoldersLocked issues Mailbox/get, populates b.folders keyed
// by canonical poplar name, and captures the state string into
// b.states["Mailbox"]. Caller must hold b.mu.
func (b *Backend) refreshFoldersLocked() error {
	accountID := b.session.PrimaryAccounts[jmapmail.URI]

	req := &jmap.Request{}
	req.Invoke(&mailbox.Get{Account: accountID})

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("mailbox/get: %w", err)
	}

	for _, inv := range resp.Responses {
		gr, ok := inv.Args.(*mailbox.GetResponse)
		if !ok {
			continue
		}
		b.states["Mailbox"] = gr.State

		// Build raw mail.Folder slice to run through the classifier.
		raw := make([]mail.Folder, 0, len(gr.List))
		for _, mbox := range gr.List {
			raw = append(raw, mail.Folder{
				Name:   mbox.Name,
				Exists: int(mbox.TotalEmails),
				Unseen: int(mbox.UnreadEmails),
				Role:   string(mbox.Role),
			})
		}

		classified := mail.Classify(raw)
		for i, cf := range classified {
			key := cf.DisplayName
			b.folders[key] = folderEntry{
				id:     string(gr.List[i].ID),
				folder: cf.Folder,
			}
		}
		break
	}
	return nil
}

// Disconnect satisfies mail.Backend. It cancels the push loop (Task
// 13) and tears down all session state.
func (b *Backend) Disconnect() error {
	b.mu.Lock()
	cancel := b.pushCancel
	done := b.pushDone
	b.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.updates != nil {
		close(b.updates)
		b.updates = nil
	}
	b.client = nil
	b.session = nil
	b.current = ""
	b.folders = make(map[string]folderEntry)
	b.blobIDs = make(map[mail.UID]string)
	b.states = make(map[string]string)
	b.bodies = nil
	return nil
}

// ListFolders satisfies mail.Backend. It returns the cached folder
// map, refreshing from the server if the cache is empty.
func (b *Backend) ListFolders() ([]mail.Folder, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.folders) == 0 {
		if err := b.refreshFoldersLocked(); err != nil {
			return nil, fmt.Errorf("list folders: %w", err)
		}
	}
	out := make([]mail.Folder, 0, len(b.folders))
	for _, e := range b.folders {
		out = append(out, e.folder)
	}
	return out, nil
}

// OpenFolder satisfies mail.Backend. It sets the current folder to
// name; returns an error if name is not in the cached folder map.
func (b *Backend) OpenFolder(name string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.folders[name]; !ok {
		return fmt.Errorf("open folder: unknown folder %q", name)
	}
	b.current = name
	return nil
}

// QueryFolder satisfies mail.Backend. It issues Email/query against
// the named folder and returns (UIDs, total, error). offset and limit
// map directly to JMAP Position and Limit. Results are sorted by
// receivedAt descending (newest first).
func (b *Backend) QueryFolder(name string, offset, limit int) ([]mail.UID, int, error) {
	b.mu.Lock()
	entry, ok := b.folders[name]
	accountID := b.session.PrimaryAccounts[jmapmail.URI]
	b.mu.Unlock()
	if !ok {
		return nil, 0, fmt.Errorf("query folder: unknown folder %q", name)
	}

	req := &jmap.Request{}
	req.Invoke(&email.Query{
		Account: accountID,
		Filter: &email.FilterCondition{
			InMailbox: jmap.ID(entry.id),
		},
		Sort: []*email.SortComparator{
			{Property: "receivedAt", IsAscending: false},
		},
		Position:       int64(offset),
		Limit:          uint64(limit),
		CalculateTotal: true,
	})

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("query folder: %w", err)
	}

	for _, inv := range resp.Responses {
		qr, ok := inv.Args.(*email.QueryResponse)
		if !ok {
			continue
		}
		uids := make([]mail.UID, len(qr.IDs))
		for i, id := range qr.IDs {
			uids[i] = mail.UID(id)
		}
		return uids, int(qr.Total), nil
	}
	return nil, 0, fmt.Errorf("query folder: no Email/query response")
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

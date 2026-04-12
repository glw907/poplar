# Poplar Pass 2: Backend Adapter + Connect

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Define the `mail.Backend` interface, write a JMAP adapter wrapping the forked worker, parse account config from TOML, and connect to Fastmail to list folders.

**Architecture:** The `internal/mail/` package defines poplar-native types and a `Backend` interface. A JMAP adapter in that package wraps the forked aerc JMAP worker, bridging its async message-passing model to synchronous method calls via a pump goroutine. Account config is parsed from `~/.config/poplar/accounts.toml` (TOML) with credential command resolution.

**Tech Stack:** Go 1.25, BurntSushi/toml, forked aerc JMAP worker, cobra CLI

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Create | `internal/mail/types.go` | Poplar-native types: Folder, Flag, UID, Update |
| Create | `internal/mail/backend.go` | Backend interface definition |
| Create | `internal/mail/jmap.go` | JMAP adapter wrapping forked worker |
| Create | `internal/mail/jmap_test.go` | Type translation unit tests |
| Create | `internal/poplar/accounts.go` | TOML config parser + credential resolution |
| Create | `internal/poplar/accounts_test.go` | Config parsing unit tests |
| Modify | `cmd/poplar/root.go` | Wire config → adapter → connect → list → print |

---

### Task 1: Poplar-Native Types

**Files:**
- Create: `internal/mail/types.go`

- [ ] **Step 1: Create types.go with Folder, Flag, UID, and Update types**

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/glw907/Projects/beautiful-aerc && go vet ./internal/mail/`
Expected: no output (clean)

- [ ] **Step 3: Commit**

```bash
git add internal/mail/types.go
git commit -m "Add poplar-native mail types for backend interface

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 2: Backend Interface

**Files:**
- Create: `internal/mail/backend.go`

The interface covers all operations from the design spec. Only Connect, Disconnect, ListFolders, and Updates are implemented in Pass 2; the rest are defined so future passes fill them in.

- [ ] **Step 1: Create backend.go with the full Backend interface**

```go
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
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /home/glw907/Projects/beautiful-aerc && go vet ./internal/mail/`
Expected: no output (clean)

- [ ] **Step 3: Commit**

```bash
git add internal/mail/backend.go
git commit -m "Define mail.Backend interface for poplar adapters

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 3: Account Config Parser

**Files:**
- Create: `internal/poplar/accounts.go`
- Create: `internal/poplar/accounts_test.go`

Reads `~/.config/poplar/accounts.toml`, runs credential commands, and produces `[]AccountConfig`.

- [ ] **Step 1: Write the failing test for TOML parsing**

File: `internal/poplar/accounts_test.go`

```go
package poplar

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAccounts(t *testing.T) {
	tests := []struct {
		name    string
		toml    string
		wantN   int
		wantErr string
	}{
		{
			name: "single jmap account",
			toml: `[[account]]
name = "Fastmail"
backend = "jmap"
source = "jmap+oauthbearer://geoff@907.life@api.fastmail.com/.well-known/jmap"
credential-cmd = "echo test-token"
copy-to = "Sent"
folders-sort = ["Inbox", "Sent", "Archive"]
params = {cache-state = "true", cache-blobs = "true"}
`,
			wantN: 1,
		},
		{
			name: "multiple accounts",
			toml: `[[account]]
name = "Work"
backend = "jmap"
source = "jmap://user@work.com@jmap.work.com"
credential-cmd = "echo work-pass"

[[account]]
name = "Personal"
backend = "imap"
source = "imaps://user@personal.com@imap.personal.com:993"
credential-cmd = "echo personal-pass"
`,
			wantN: 2,
		},
		{
			name:    "missing name",
			toml:    "[[account]]\nbackend = \"jmap\"\nsource = \"jmap://x@y\"\n",
			wantErr: "account 0: name is required",
		},
		{
			name:    "missing source",
			toml:    "[[account]]\nname = \"Test\"\nbackend = \"jmap\"\n",
			wantErr: "account \"Test\": source is required",
		},
		{
			name:    "empty file",
			toml:    "",
			wantErr: "no accounts defined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "accounts.toml")
			if err := os.WriteFile(path, []byte(tt.toml), 0644); err != nil {
				t.Fatal(err)
			}
			accounts, err := ParseAccounts(path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(accounts) != tt.wantN {
				t.Fatalf("expected %d accounts, got %d", tt.wantN, len(accounts))
			}
		})
	}
}

func TestParseAccountsCredentialInjection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts.toml")
	toml := `[[account]]
name = "Test"
backend = "jmap"
source = "jmap+oauthbearer://user@example.com@api.example.com/.well-known/jmap"
credential-cmd = "echo secret-token"
`
	if err := os.WriteFile(path, []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}
	accounts, err := ParseAccounts(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	// Source URL should now contain the credential
	if !strings.Contains(accounts[0].Source, "secret-token") {
		t.Errorf("expected source to contain credential, got %q", accounts[0].Source)
	}
}

func TestParseAccountsFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts.toml")
	toml := `[[account]]
name = "Fastmail"
backend = "jmap"
source = "jmap://user@fm.com@api.fm.com"
credential-cmd = "echo pass"
copy-to = "Sent"
folders-sort = ["Inbox", "Sent"]
from = "Test User <test@fm.com>"
params = {cache-state = "true"}
`
	if err := os.WriteFile(path, []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}
	accounts, err := ParseAccounts(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := accounts[0]
	if a.Name != "Fastmail" {
		t.Errorf("name = %q, want %q", a.Name, "Fastmail")
	}
	if a.Backend != "jmap" {
		t.Errorf("backend = %q, want %q", a.Backend, "jmap")
	}
	if len(a.CopyTo) != 1 || a.CopyTo[0] != "Sent" {
		t.Errorf("copy-to = %v, want [Sent]", a.CopyTo)
	}
	if len(a.Folders) != 2 || a.Folders[0] != "Inbox" {
		t.Errorf("folders = %v, want [Inbox Sent]", a.Folders)
	}
	if a.Params["cache-state"] != "true" {
		t.Errorf("params[cache-state] = %q, want %q", a.Params["cache-state"], "true")
	}
	if a.From == nil || a.From.Address != "test@fm.com" {
		t.Errorf("from = %v, want test@fm.com", a.From)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/beautiful-aerc && go test ./internal/poplar/ -run TestParseAccounts -v`
Expected: FAIL — `ParseAccounts` not defined

- [ ] **Step 3: Implement the config parser**

File: `internal/poplar/accounts.go`

```go
package poplar

import (
	"fmt"
	"net/mail"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

type configFile struct {
	Account []accountEntry `toml:"account"`
}

type accountEntry struct {
	Name            string            `toml:"name"`
	Backend         string            `toml:"backend"`
	Source          string            `toml:"source"`
	CredentialCmd   string            `toml:"credential-cmd"`
	CopyTo          string            `toml:"copy-to"`
	FoldersSort     []string          `toml:"folders-sort"`
	FoldersExclude  []string          `toml:"folders-exclude"`
	From            string            `toml:"from"`
	Outgoing        string            `toml:"outgoing"`
	OutgoingCredCmd string            `toml:"outgoing-credential-cmd"`
	Params          map[string]string `toml:"params"`
}

// ParseAccounts reads a poplar accounts.toml file and returns
// configured accounts with credentials resolved.
func ParseAccounts(path string) ([]AccountConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading accounts config: %w", err)
	}

	var cf configFile
	if err := toml.Unmarshal(data, &cf); err != nil {
		return nil, fmt.Errorf("parsing accounts config: %w", err)
	}

	if len(cf.Account) == 0 {
		return nil, fmt.Errorf("no accounts defined")
	}

	var accounts []AccountConfig
	for i, entry := range cf.Account {
		acct, err := entry.toAccountConfig(i)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *acct)
	}
	return accounts, nil
}

func (e *accountEntry) toAccountConfig(index int) (*AccountConfig, error) {
	if e.Name == "" {
		return nil, fmt.Errorf("account %d: name is required", index)
	}
	if e.Source == "" {
		return nil, fmt.Errorf("account %q: source is required", e.Name)
	}

	source := e.Source
	if e.CredentialCmd != "" {
		cred, err := runCredentialCmd(e.CredentialCmd)
		if err != nil {
			return nil, fmt.Errorf("account %q: credential command: %w", e.Name, err)
		}
		source, err = injectCredential(source, cred)
		if err != nil {
			return nil, fmt.Errorf("account %q: injecting credential: %w", e.Name, err)
		}
	}

	acct := &AccountConfig{
		Name:            e.Name,
		Backend:         e.Backend,
		Source:          source,
		Folders:         e.FoldersSort,
		FoldersExclude:  e.FoldersExclude,
		Params:          e.Params,
		Outgoing:        e.Outgoing,
		OutgoingCredCmd: e.OutgoingCredCmd,
	}

	if e.CopyTo != "" {
		acct.CopyTo = []string{e.CopyTo}
	}

	if e.From != "" {
		addr, err := mail.ParseAddress(e.From)
		if err != nil {
			return nil, fmt.Errorf("account %q: parsing from address: %w", e.Name, err)
		}
		acct.From = addr
	}

	return acct, nil
}

func runCredentialCmd(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", fmt.Errorf("running %q: %w", cmd, err)
	}
	return strings.TrimSpace(string(out)), nil
}

func injectCredential(source, credential string) (string, error) {
	u, err := url.Parse(source)
	if err != nil {
		return "", fmt.Errorf("parsing source URL: %w", err)
	}
	username := ""
	if u.User != nil {
		username = u.User.Username()
	}
	u.User = url.UserPassword(username, credential)
	return u.String(), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/glw907/Projects/beautiful-aerc && go test ./internal/poplar/ -run TestParseAccounts -v`
Expected: PASS (all three test functions)

- [ ] **Step 5: Commit**

```bash
git add internal/poplar/accounts.go internal/poplar/accounts_test.go
git commit -m "Add TOML account config parser with credential resolution

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 4: JMAP Adapter

**Files:**
- Create: `internal/mail/jmap.go`
- Create: `internal/mail/jmap_test.go`

Wraps the forked JMAP worker's async message-passing with synchronous method calls. A pump goroutine reads from `types.WorkerMessages` and dispatches callbacks.

- [ ] **Step 1: Write the failing test for folder translation**

File: `internal/mail/jmap_test.go`

```go
package mail

import (
	"testing"

	"github.com/glw907/beautiful-aerc/internal/aercfork/models"
)

func TestTranslateFolder(t *testing.T) {
	tests := []struct {
		name string
		dir  *models.Directory
		want Folder
	}{
		{
			name: "inbox with unread",
			dir: &models.Directory{
				Name:   "Inbox",
				Exists: 42,
				Unseen: 5,
				Role:   models.InboxRole,
			},
			want: Folder{
				Name:   "Inbox",
				Exists: 42,
				Unseen: 5,
				Role:   "inbox",
			},
		},
		{
			name: "sent with no role",
			dir: &models.Directory{
				Name:   "Sent",
				Exists: 100,
				Unseen: 0,
			},
			want: Folder{
				Name:   "Sent",
				Exists: 100,
				Unseen: 0,
				Role:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translateFolder(tt.dir)
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Exists != tt.want.Exists {
				t.Errorf("Exists = %d, want %d", got.Exists, tt.want.Exists)
			}
			if got.Unseen != tt.want.Unseen {
				t.Errorf("Unseen = %d, want %d", got.Unseen, tt.want.Unseen)
			}
			if got.Role != tt.want.Role {
				t.Errorf("Role = %q, want %q", got.Role, tt.want.Role)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/glw907/Projects/beautiful-aerc && go test ./internal/mail/ -run TestTranslateFolder -v`
Expected: FAIL — `translateFolder` not defined

- [ ] **Step 3: Implement the JMAP adapter**

File: `internal/mail/jmap.go`

```go
package mail

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/glw907/beautiful-aerc/internal/aercfork/models"
	"github.com/glw907/beautiful-aerc/internal/aercfork/worker"
	"github.com/glw907/beautiful-aerc/internal/aercfork/worker/types"
	"github.com/glw907/beautiful-aerc/internal/poplar"
)

// JMAPAdapter wraps the forked aerc JMAP worker behind the Backend
// interface, bridging async message-passing to synchronous calls.
type JMAPAdapter struct {
	config  *poplar.AccountConfig
	w       *types.Worker
	updates chan Update
	done    chan struct{}
	mu      sync.Mutex
}

// NewJMAPAdapter creates a JMAP backend adapter for the given account.
func NewJMAPAdapter(config *poplar.AccountConfig) (*JMAPAdapter, error) {
	w, err := worker.NewWorker(config.Source, config.Name)
	if err != nil {
		return nil, fmt.Errorf("creating worker: %w", err)
	}
	return &JMAPAdapter{
		config:  config,
		w:       w,
		updates: make(chan Update, 50),
		done:    make(chan struct{}),
	}, nil
}

// Connect configures and connects the JMAP worker to the server.
func (a *JMAPAdapter) Connect(ctx context.Context) error {
	go a.w.Backend.Run()
	go a.pump()

	if err := a.doAction(&types.Configure{Config: a.config}); err != nil {
		return fmt.Errorf("configuring worker: %w", err)
	}
	if err := a.doAction(&types.Connect{}); err != nil {
		return fmt.Errorf("connecting: %w", err)
	}
	return nil
}

// Disconnect stops the message pump.
func (a *JMAPAdapter) Disconnect() error {
	close(a.done)
	return a.doAction(&types.Disconnect{})
}

// ListFolders returns all mail folders from the server.
func (a *JMAPAdapter) ListFolders() ([]Folder, error) {
	var folders []Folder
	err := a.doCollect(&types.ListDirectories{}, func(msg types.WorkerMessage) {
		if d, ok := msg.(*types.Directory); ok {
			folders = append(folders, translateFolder(d.Dir))
		}
	})
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}
	return folders, nil
}

// OpenFolder selects a folder as the current working folder.
func (a *JMAPAdapter) OpenFolder(name string) error {
	return a.doAction(&types.OpenDirectory{Directory: name})
}

// FetchHeaders retrieves header info for the given message UIDs.
func (a *JMAPAdapter) FetchHeaders(uids []UID) ([]MessageInfo, error) {
	return nil, fmt.Errorf("not implemented")
}

// FetchBody retrieves the full body of a single message.
func (a *JMAPAdapter) FetchBody(uid UID) (io.Reader, error) {
	return nil, fmt.Errorf("not implemented")
}

// Search finds messages matching the given criteria.
func (a *JMAPAdapter) Search(criteria SearchCriteria) ([]UID, error) {
	return nil, fmt.Errorf("not implemented")
}

// Move moves messages to the destination folder.
func (a *JMAPAdapter) Move(uids []UID, dest string) error {
	return fmt.Errorf("not implemented")
}

// Copy copies messages to the destination folder.
func (a *JMAPAdapter) Copy(uids []UID, dest string) error {
	return fmt.Errorf("not implemented")
}

// Delete moves messages to trash.
func (a *JMAPAdapter) Delete(uids []UID) error {
	return fmt.Errorf("not implemented")
}

// Flag sets or clears a flag on messages.
func (a *JMAPAdapter) Flag(uids []UID, flag Flag, set bool) error {
	return fmt.Errorf("not implemented")
}

// MarkRead marks messages as read.
func (a *JMAPAdapter) MarkRead(uids []UID) error {
	return fmt.Errorf("not implemented")
}

// MarkAnswered marks messages as answered.
func (a *JMAPAdapter) MarkAnswered(uids []UID) error {
	return fmt.Errorf("not implemented")
}

// Send sends a message.
func (a *JMAPAdapter) Send(from string, rcpts []string, body io.Reader) error {
	return fmt.Errorf("not implemented")
}

// Updates returns a channel of asynchronous backend updates.
func (a *JMAPAdapter) Updates() <-chan Update {
	return a.updates
}

// pump reads worker response messages and dispatches callbacks.
// Runs in its own goroutine, started by Connect.
func (a *JMAPAdapter) pump() {
	for {
		select {
		case <-a.done:
			return
		case msg := <-types.WorkerMessages:
			a.w.ProcessMessage(msg)
		}
	}
}

// doAction sends an action to the worker and blocks until Done or Error.
func (a *JMAPAdapter) doAction(msg types.WorkerMessage) error {
	ch := make(chan error, 1)
	a.w.PostAction(msg, func(resp types.WorkerMessage) {
		switch r := resp.(type) {
		case *types.Done:
			ch <- nil
		case *types.Error:
			ch <- r.Error
		case *types.ConnError:
			ch <- r.Error
		case *types.Unsupported:
			ch <- fmt.Errorf("unsupported operation")
		}
	})
	return <-ch
}

// doCollect sends an action and calls collect for each intermediate
// response before the final Done/Error.
func (a *JMAPAdapter) doCollect(msg types.WorkerMessage, collect func(types.WorkerMessage)) error {
	ch := make(chan error, 1)
	a.w.PostAction(msg, func(resp types.WorkerMessage) {
		switch r := resp.(type) {
		case *types.Done:
			ch <- nil
		case *types.Error:
			ch <- r.Error
		case *types.ConnError:
			ch <- r.Error
		case *types.Unsupported:
			ch <- fmt.Errorf("unsupported operation")
		default:
			collect(resp)
		}
	})
	return <-ch
}

func translateFolder(d *models.Directory) Folder {
	return Folder{
		Name:   d.Name,
		Exists: d.Exists,
		Unseen: d.Unseen,
		Role:   string(d.Role),
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/glw907/Projects/beautiful-aerc && go test ./internal/mail/ -run TestTranslateFolder -v`
Expected: PASS

- [ ] **Step 5: Verify full build**

Run: `cd /home/glw907/Projects/beautiful-aerc && go vet ./internal/mail/`
Expected: no output (clean)

- [ ] **Step 6: Commit**

```bash
git add internal/mail/jmap.go internal/mail/jmap_test.go
git commit -m "Add JMAP adapter wrapping forked worker with sync interface

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 5: Wire CLI

**Files:**
- Modify: `cmd/poplar/root.go`

Add a RunE that reads config, creates a JMAP adapter, connects, lists folders, and prints them to stdout. This is the Pass 2 gate test.

- [ ] **Step 1: Create the user's config file**

Create `~/.config/poplar/accounts.toml` with the user's Fastmail config:

```toml
[[account]]
name = "Fastmail"
backend = "jmap"
source = "jmap+oauthbearer://geoff@907.life@api.fastmail.com/.well-known/jmap"
credential-cmd = "fastmail-password"
copy-to = "Sent"
folders-sort = ["Inbox", "Notifications", "Drafts", "Sent", "Archive"]
params = {cache-state = "true", cache-blobs = "true"}
```

- [ ] **Step 2: Update root.go with connect + list folders**

File: `cmd/poplar/root.go`

```go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/poplar"
	"github.com/spf13/cobra"

	// Import forked workers for init() side effects (handler registration).
	_ "github.com/glw907/beautiful-aerc/internal/aercfork/worker"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "poplar",
		Short:        "A bubbletea-based terminal email client",
		SilenceUsage: true,
		RunE:         runRoot,
	}
	return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("finding home directory: %w", err)
	}

	configPath := filepath.Join(home, ".config", "poplar", "accounts.toml")
	accounts, err := poplar.ParseAccounts(configPath)
	if err != nil {
		return fmt.Errorf("loading accounts: %w", err)
	}

	acct := &accounts[0]
	adapter, err := mail.NewJMAPAdapter(acct)
	if err != nil {
		return fmt.Errorf("creating adapter: %w", err)
	}

	ctx := context.Background()
	if err := adapter.Connect(ctx); err != nil {
		return fmt.Errorf("connecting: %w", err)
	}

	folders, err := adapter.ListFolders()
	if err != nil {
		return fmt.Errorf("listing folders: %w", err)
	}

	for _, f := range folders {
		role := ""
		if f.Role != "" {
			role = " [" + f.Role + "]"
		}
		fmt.Fprintf(os.Stdout, "%-30s %d messages, %d unread%s\n",
			f.Name, f.Exists, f.Unseen, role)
	}

	return nil
}
```

- [ ] **Step 3: Build and verify compilation**

Run: `cd /home/glw907/Projects/beautiful-aerc && make build`
Expected: all four binaries compile without errors

- [ ] **Step 4: Commit**

```bash
git add cmd/poplar/root.go
git commit -m "Wire poplar CLI to connect and list folders via JMAP adapter

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

### Task 6: Integration Test + Quality Gate

- [ ] **Step 1: Run make check**

Run: `cd /home/glw907/Projects/beautiful-aerc && make check`
Expected: vet clean, all tests pass

- [ ] **Step 2: Create user config directory and file**

Run: `mkdir -p ~/.config/poplar`

Write `~/.config/poplar/accounts.toml` with the Fastmail account config from Task 5 Step 1.

- [ ] **Step 3: Run poplar and verify folder listing**

Run: `cd /home/glw907/Projects/beautiful-aerc && ./poplar`

Expected: prints folder names with message counts and unread counts from the Fastmail account. Output like:
```
Inbox                          42 messages, 5 unread [inbox]
Sent                           1234 messages, 0 unread [sent]
Archive                        5678 messages, 0 unread [archive]
...
```

The program should exit cleanly after printing.

- [ ] **Step 4: Fix any issues found during integration test**

If the connection fails or output is unexpected, debug and fix. Common issues:
- Credential command not found: verify `fastmail-password` is in PATH
- URL parsing: check that `+oauthbearer` suffix is preserved through credential injection
- Folder names: verify the JMAP worker's `MailboxPath()` produces expected names

- [ ] **Step 5: Install binaries**

Run: `cd /home/glw907/Projects/beautiful-aerc && make install`
Expected: all four binaries installed to `~/.local/bin/`

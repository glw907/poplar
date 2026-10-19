# Pass 3 — JMAP direct-on-rockorager backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `internal/mailjmap/` directly against
`git.sr.ht/~rockorager/go-jmap` under a synchronous `mail.Backend`,
vendor minimal auth/keepalive snippets into `internal/mailauth/`,
delete `internal/mailworker/` entirely, and wire the prototype to a
live Fastmail account.

**Architecture:** One synchronous `*Backend` per account in
`internal/mailjmap/`. RPC methods call `*jmap.Client` directly. A
single push goroutine owned by `Connect`/`Disconnect` reads
`StateChange` events from the JMAP push EventSource, fans deltas
through `Email/changes` + `Email/get`, and emits `mail.Update`
values onto the existing `Updates() <-chan mail.Update` channel.
Body fetches go through an in-memory LRU + singleflight collapse.
Folder loads use a 500-message window with `Email/query` offset; the
UI lazy-loads further windows on cursor proximity to the bottom.

**Tech Stack:** Go 1.26.1, `git.sr.ht/~rockorager/go-jmap@v0.5.3`
(already vendored), `github.com/hashicorp/golang-lru/v2` (new dep
this pass), `golang.org/x/sync/singleflight` (transitive), bubbletea
(already vendored), `emersion/go-message` (already vendored, used by
existing content pipeline). Config TOML is already on
`BurntSushi/toml`.

**Spec:** `docs/superpowers/specs/2026-04-25-jmap-direct-backend-design.md`

**Required reading before starting:**
- Invoke `go-conventions` skill before writing any Go file.
- Invoke `elm-conventions` skill before writing any TUI code.
- Read `docs/poplar/invariants.md` once.
- Read the spec once.
- Skim `git.sr.ht/~rockorager/go-jmap` package surface at
  `~/go/pkg/mod/git.sr.ht/~rockorager/go-jmap@v0.5.3/` —
  particularly `mail/`, `core/`, `client.go`, `session.go`,
  `statechange.go`.
- Skim `internal/mail/backend.go`, `internal/mail/types.go`, and
  the existing `internal/mailjmap/jmap.go` (will be rewritten).

**Conventions for this plan:** Each task is one coherent unit of
work that ends with `make check` green and a commit. Tests are
table-driven, use `*testing.T` directly (no testify outside
existing `mailworker` code being deleted). Errors wrap with
`fmt.Errorf("%s: %w", op, err)`. `tea.Cmd` failures return
`ui.ErrorMsg{Op, Err}`. Commits use imperative mood with
Co-Authored-By trailer per repo convention.

---

## Phase 0 — Interface and dependency prep

These changes land before any backend code so the rewrite has
something to compile against. The MockBackend updates also unblock
the UI work in Phase 3.

### Task 1: Add `QueryFolder` to `mail.Backend` and update MockBackend

**Files:**
- Modify: `internal/mail/backend.go`
- Modify: `internal/mail/mock.go`
- Modify: `internal/mail/mock_test.go`
- Modify: `internal/mailjmap/jmap.go` (stub the new method on the soon-to-be-deleted adapter)

- [ ] **Step 1: Add the method to the interface**

In `internal/mail/backend.go` add to the `Backend` interface,
between `OpenFolder` and `FetchHeaders`:

```go
// QueryFolder returns up to limit message UIDs from name starting
// at offset (newest-first), plus the total message count. The
// total enables the UI to show "showing N of M" and to stop
// dispatching load-more once exhausted.
QueryFolder(name string, offset, limit int) (uids []UID, total int, err error)
```

- [ ] **Step 2: Implement on MockBackend**

Append to `internal/mail/mock.go` (alongside the other mock methods):

```go
// QueryFolder slices the hardcoded message list. The mock ignores
// folder name (always returns the same set) and clamps offset/limit
// to the available range.
func (m *MockBackend) QueryFolder(_ string, offset, limit int) ([]UID, int, error) {
    total := len(m.msgs)
    if offset >= total {
        return nil, total, nil
    }
    end := offset + limit
    if end > total {
        end = total
    }
    uids := make([]UID, 0, end-offset)
    for _, msg := range m.msgs[offset:end] {
        uids = append(uids, msg.UID)
    }
    return uids, total, nil
}
```

- [ ] **Step 3: Stub on the existing JMAPAdapter**

In `internal/mailjmap/jmap.go` (which gets rewritten in Phase 2;
this stub is just to keep the build green during Phase 0):

```go
func (a *JMAPAdapter) QueryFolder(_ string, _, _ int) ([]mail.UID, int, error) {
    return nil, 0, fmt.Errorf("not implemented")
}
```

- [ ] **Step 4: Add a test for the mock implementation**

Append to `internal/mail/mock_test.go`:

```go
func TestMockBackend_QueryFolder(t *testing.T) {
    b := NewMockBackend()
    total := len(b.msgs)
    cases := []struct {
        name           string
        offset, limit  int
        wantLen        int
    }{
        {"first window", 0, 5, 5},
        {"past end", total + 10, 5, 0},
        {"clamps end", total - 2, 10, 2},
        {"zero limit", 0, 0, 0},
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            uids, gotTotal, err := b.QueryFolder("Inbox", tc.offset, tc.limit)
            if err != nil {
                t.Fatalf("QueryFolder: %v", err)
            }
            if gotTotal != total {
                t.Errorf("total = %d, want %d", gotTotal, total)
            }
            if len(uids) != tc.wantLen {
                t.Errorf("len(uids) = %d, want %d", len(uids), tc.wantLen)
            }
        })
    }
}
```

- [ ] **Step 5: `make check`**

```
make check
```
Expected: all tests pass.

- [ ] **Step 6: Commit**

```
git add internal/mail internal/mailjmap/jmap.go
git commit -m "Pass 3 prep: add QueryFolder to mail.Backend"
```

---

### Task 2: Add `UpdateConnState` and `ConnState` to `mail.Update`

**Files:**
- Modify: `internal/mail/types.go`

- [ ] **Step 1: Extend the const block and Update struct**

Replace the entire const block + Update struct in
`internal/mail/types.go` with:

```go
// UpdateType classifies asynchronous backend updates.
type UpdateType int

const (
    UpdateNewMail UpdateType = iota
    UpdateFlagsChanged
    UpdateExpunge
    UpdateFolderInfo
    UpdateConnState
)

// ConnState classifies the backend's transport state. Carried on
// Update.ConnState only when Type == UpdateConnState.
type ConnState int

const (
    ConnOffline ConnState = iota
    ConnReconnecting
    ConnConnected
)

// Update represents an asynchronous update from the backend. The
// ConnState field is populated only for UpdateConnState; for every
// other Type it is the zero value (ConnOffline) and ignored.
type Update struct {
    Type      UpdateType
    Folder    string
    UIDs      []UID
    ConnState ConnState
}
```

- [ ] **Step 2: `make check`**

Build only — no behavior change yet. Should compile cleanly because
the new field is additive and the new const is unused.

- [ ] **Step 3: Commit**

```
git add internal/mail/types.go
git commit -m "Pass 3 prep: add UpdateConnState + ConnState to mail.Update"
```

---

### Task 3: Add env-var substitution to `config.ParseAccounts`

**Files:**
- Modify: `internal/config/accounts.go` (or wherever `ParseAccounts` lives — `grep -l ParseAccounts internal/config`)
- Modify: corresponding `_test.go`

The Pass 3 live wiring needs `password = "$FASTMAIL_API_TOKEN"` in
`accounts.toml` to resolve from the environment so secrets do not
land in git.

- [ ] **Step 1: Find the parse path**

```
grep -n "ParseAccounts\|Password" internal/config/*.go
```

- [ ] **Step 2: Add a helper and wire it into the loop that
  populates `AccountConfig` fields**

```go
// resolveEnv replaces a leading "$VAR" with os.Getenv("VAR"). The
// only supported form is the bare $VAR token; anything else is
// returned unchanged so passwords containing a literal "$" still
// work. Empty env returns an error so the user gets a clear
// failure on misconfiguration.
func resolveEnv(s string) (string, error) {
    if !strings.HasPrefix(s, "$") || len(s) < 2 {
        return s, nil
    }
    name := s[1:]
    if !isShellName(name) {
        return s, nil
    }
    val := os.Getenv(name)
    if val == "" {
        return "", fmt.Errorf("env var %s is empty or unset", name)
    }
    return val, nil
}

func isShellName(s string) bool {
    for i, r := range s {
        if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
            continue
        }
        if i > 0 && r >= '0' && r <= '9' {
            continue
        }
        return false
    }
    return s != ""
}
```

Call it from the parse loop right after decoding `acct.Password`:

```go
resolved, err := resolveEnv(acct.Password)
if err != nil {
    return nil, fmt.Errorf("account %q password: %w", acct.Name, err)
}
acct.Password = resolved
```

- [ ] **Step 3: Add a table-driven test**

Cover: literal password unchanged; `$VAR` resolves when set;
unset `$VAR` errors; literal `$` not followed by a name unchanged
(`"$"`, `"$1abc"`).

- [ ] **Step 4: `make check` + commit**

```
git add internal/config
git commit -m "Pass 3 prep: env-var substitution in account passwords"
```

---

### Task 4: Add the `golang-lru/v2` dependency

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: `go get`**

```
go get github.com/hashicorp/golang-lru/v2@latest
go mod tidy
```

- [ ] **Step 2: Verify build still works**

```
make check
```

- [ ] **Step 3: Commit**

```
git add go.mod go.sum
git commit -m "Pass 3 prep: add hashicorp/golang-lru/v2 dependency"
```

---

## Phase 1 — `internal/mailauth/` vendored snippets

Two small, MIT-licensed files vendored from aerc with provenance
comments. Built and tested in isolation so Phase 2 has the auth
helper available.

### Task 5: Vendor `xoauth2.go`

**Files:**
- Create: `internal/mailauth/xoauth2.go`
- Create: `internal/mailauth/xoauth2_test.go`
- Create: `internal/mailauth/README.md`

- [ ] **Step 1: Locate the source in the existing fork**

```
find internal/mailworker -name "xoauth2*"
```

Expected: `internal/mailworker/auth/xoauth2.go` (or similar). Read
it and the surrounding files to understand the dependencies — it
should depend only on `emersion/go-sasl`.

- [ ] **Step 2: Copy into `internal/mailauth/xoauth2.go` with a
  provenance comment**

Top of file:

```go
// Vendored from git.sr.ht/~rjarry/aerc (commit <hash from go.mod or
// the fork commit message>) — auth/xoauth2.go. MIT-licensed.
// Modifications: package renamed to mailauth; <list any other edits
// or "none">. Provides the XOAUTH2 SASL mechanism that
// emersion/go-sasl does not ship.
```

Package declaration:

```go
package mailauth
```

Keep the rest of the file byte-identical where possible.

- [ ] **Step 3: Add a vector test**

Use a known-good XOAUTH2 challenge from the OAuth 2.0 spec
(`user=alice@example.com\x01auth=Bearer FAKE_TOKEN\x01\x01`,
base64-encoded). Build a `Client` via the new constructor, drive
one round of `Next`, assert the bytes match.

```go
func TestXOAuth2_Challenge(t *testing.T) {
    cli := NewXOAuth2Client("alice@example.com", "FAKE_TOKEN")
    mech, ir, err := cli.Start()
    if err != nil {
        t.Fatalf("Start: %v", err)
    }
    if mech != "XOAUTH2" {
        t.Errorf("mech = %q, want XOAUTH2", mech)
    }
    want := "user=alice@example.com\x01auth=Bearer FAKE_TOKEN\x01\x01"
    if string(ir) != want {
        t.Errorf("initial response =\n  %q\nwant\n  %q", ir, want)
    }
}
```

(Adjust the constructor name and function signatures to whatever
the vendored file actually exports — read it before writing the
test.)

- [ ] **Step 4: Add `internal/mailauth/README.md`**

```markdown
# internal/mailauth

Small vendored snippets that fill gaps in the emersion mail stack.

| File | Origin | License | Why |
|---|---|---|---|
| `xoauth2.go` | aerc `auth/xoauth2.go` | MIT | XOAUTH2 SASL mech (`emersion/go-sasl` does not ship one). |
| `keepalive/keepalive.go` | aerc `lib/keepalive.go` (or similar) | MIT | TCP keepalive helper for long-lived IMAP/JMAP connections. |

Each file carries a provenance comment recording the source commit
and any modifications. Update those comments if upstream changes
land.
```

- [ ] **Step 5: `make check` + commit**

```
git add internal/mailauth
git commit -m "Pass 3: vendor XOAUTH2 SASL helper into internal/mailauth"
```

---

### Task 6: Vendor `keepalive/`

**Files:**
- Create: `internal/mailauth/keepalive/keepalive.go`

- [ ] **Step 1: Copy from the existing fork**

```
cp -i internal/mailworker/keepalive/*.go internal/mailauth/keepalive/
```

(Or whichever path the fork actually uses — `find internal/mailworker -name "keepalive*"`.)

- [ ] **Step 2: Adjust package + provenance**

Add provenance comment as in Task 5. Change package declaration to
`package keepalive`. Drop any unused imports.

- [ ] **Step 3: `make check` + commit**

```
git add internal/mailauth/keepalive
git commit -m "Pass 3: vendor TCP keepalive helper into internal/mailauth"
```

---

## Phase 2 — `internal/mailjmap/` direct rewrite

The big one. We rewrite the package from scratch as a thin layer over
`git.sr.ht/~rockorager/go-jmap`. Each task ships a coherent slice
that builds and tests independently.

### Task 7: Wipe and re-skeleton the package

**Files:**
- Replace: `internal/mailjmap/jmap.go`
- Delete: `internal/mailjmap/jmap_test.go` (will be rewritten task by task)
- Create: `internal/mailjmap/jmap_test.go`

- [ ] **Step 1: Delete the old file content and write the skeleton**

```go
// Package mailjmap implements mail.Backend for JMAP servers
// (Fastmail) by calling git.sr.ht/~rockorager/go-jmap directly.
// All RPC methods are synchronous. A single goroutine owned by
// Connect/Disconnect reads JMAP push StateChange events and emits
// mail.Update values onto the channel returned by Updates().
package mailjmap

import (
    "context"
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

    mu       sync.Mutex
    client   *jmap.Client     // built in Connect
    session  *jmap.Session
    current  string           // current folder name
    folders  map[string]folderEntry
    blobIDs  map[mail.UID]string
    states   map[string]string

    bodies   *lru.Cache[string, []byte]
    updates  chan mail.Update

    pushCancel context.CancelFunc
    pushDone   chan struct{}
}

type folderEntry struct {
    id     string // JMAP mailbox id
    folder mail.Folder
}

// New constructs an unconnected Backend for cfg. cfg.Protocol must
// be "jmap"; cfg.Host is the JMAP session URL (e.g.
// https://api.fastmail.com/jmap/session); cfg.Username and
// cfg.Password (post env-var substitution) supply XOAUTH2 bearer
// auth.
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
```

Add stub implementations for the rest of `mail.Backend` returning
`fmt.Errorf("not implemented")` so this task compiles. Subsequent
tasks fill them in.

- [ ] **Step 2: Compile-only check**

```
make build
```

- [ ] **Step 3: Add a constructor smoke test**

```go
// internal/mailjmap/jmap_test.go
package mailjmap

import (
    "testing"

    "github.com/glw907/poplar/internal/config"
)

func TestNew_AccountName(t *testing.T) {
    b := New(config.AccountConfig{Name: "alice@example.com"})
    if got := b.AccountName(); got != "alice@example.com" {
        t.Errorf("AccountName = %q, want %q", got, "alice@example.com")
    }
}
```

- [ ] **Step 4: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: skeleton mailjmap rewrite (no behavior yet)"
```

---

### Task 8: Implement `Connect` and `Disconnect`

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Modify: `internal/mailjmap/jmap_test.go`

The session bootstrap, no push loop yet — that lands in Task 13.

- [ ] **Step 1: Implement Connect**

```go
const (
    bodyCacheSize = 64
    updatesBuffer = 64
)

func (b *Backend) Connect(ctx context.Context) error {
    b.mu.Lock()
    defer b.mu.Unlock()

    cli := &jmap.Client{
        SessionEndpoint: b.cfg.Host,
    }
    cli.WithAccessToken(b.cfg.Password)
    if err := cli.Authenticate(); err != nil {
        return fmt.Errorf("connect: authenticate: %w", err)
    }
    b.client = cli
    b.session = cli.Session

    if err := b.refreshFolders(); err != nil {
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
```

(Adjust the `*jmap.Client` API surface to match the actual library
— `client.go` in the vendored package shows the real method names.
The XOAUTH2 mech is unused here because Fastmail accepts a bearer
token directly; the helper from `internal/mailauth` is needed only
when poplar talks IMAP in Pass 8.)

- [ ] **Step 2: Implement refreshFolders (private helper used by
  Connect and the push loop)**

`refreshFolders` issues `Mailbox/get`, populates `b.folders` keyed
by canonical poplar name (run the names through the existing
`mail.Classify` to normalize "INBOX" → "Inbox" etc.), captures the
state string into `b.states["Mailbox"]`.

- [ ] **Step 3: Implement Disconnect**

```go
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
    b.folders = make(map[string]folderEntry)
    b.blobIDs = make(map[mail.UID]string)
    b.states = make(map[string]string)
    b.bodies = nil
    return nil
}
```

- [ ] **Step 4: Test with a fake Client shim**

The cleanest path: introduce a small `jmapClient` interface inside
the package that `*jmap.Client` already satisfies (one method per
RPC poplar uses), then test against a fake. Defer that
introduction to Task 9 where it carries more weight; for Task 8
just write a smoke test that constructs a Backend with no
networking and calls Disconnect (idempotent on never-connected).

```go
func TestBackend_DisconnectWithoutConnect(t *testing.T) {
    b := New(config.AccountConfig{Name: "alice"})
    if err := b.Disconnect(); err != nil {
        t.Fatalf("Disconnect on never-connected: %v", err)
    }
}
```

- [ ] **Step 5: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: mailjmap Connect/Disconnect lifecycle"
```

---

### Task 9: Introduce the `jmapClient` test seam

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Create: `internal/mailjmap/fake_test.go`

To keep tests fast and offline, define a narrow interface that
`*jmap.Client` satisfies and that a test fake can implement.

- [ ] **Step 1: Define the interface**

In `jmap.go`:

```go
// jmapClient is the subset of *jmap.Client poplar uses. The real
// *jmap.Client satisfies it; tests substitute a fake.
type jmapClient interface {
    Do(req *jmap.Request) (*jmap.Response, error)
}
```

Replace the `client *jmap.Client` field on `Backend` with
`client jmapClient`. `Authenticate` happens before the `Backend`
sees the client, so it stays out of the interface.

A constructor variant for tests:

```go
// NewWithClient is for tests. It bypasses the network handshake
// and installs a pre-built client that already satisfies the
// session contract.
func NewWithClient(cfg config.AccountConfig, c jmapClient) *Backend {
    b := New(cfg)
    b.client = c
    b.bodies, _ = lru.New[string, []byte](bodyCacheSize)
    b.updates = make(chan mail.Update, updatesBuffer)
    return b
}
```

- [ ] **Step 2: Build a `fakeClient` in `fake_test.go`**

```go
type fakeClient struct {
    mu       sync.Mutex
    sent     []*jmap.Request
    respond  func(req *jmap.Request) (*jmap.Response, error)
}

func (f *fakeClient) Do(req *jmap.Request) (*jmap.Response, error) {
    f.mu.Lock()
    f.sent = append(f.sent, req)
    f.mu.Unlock()
    if f.respond == nil {
        return &jmap.Response{}, nil
    }
    return f.respond(req)
}
```

Plus a small builder helper that constructs a `*jmap.Response` with
the given `MethodResponses` slice — the rest of the tasks lean on
this.

- [ ] **Step 3: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: introduce jmapClient seam for offline tests"
```

---

### Task 10: Implement `ListFolders`, `OpenFolder`, `QueryFolder`

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Modify: `internal/mailjmap/jmap_test.go`

- [ ] **Step 1: Implement methods**

```go
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

func (b *Backend) OpenFolder(name string) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    if _, ok := b.folders[name]; !ok {
        return fmt.Errorf("open folder: unknown folder %q", name)
    }
    b.current = name
    return nil
}

func (b *Backend) QueryFolder(name string, offset, limit int) ([]mail.UID, int, error) {
    b.mu.Lock()
    entry, ok := b.folders[name]
    b.mu.Unlock()
    if !ok {
        return nil, 0, fmt.Errorf("query folder: unknown folder %q", name)
    }

    req := &jmap.Request{Using: []string{"urn:ietf:params:jmap:mail"}}
    queryCallID := req.Invoke(&mailcap.EmailQuery{
        AccountID:      b.session.PrimaryAccount("urn:ietf:params:jmap:mail"),
        Filter:         &mailcap.EmailFilterCondition{InMailbox: entry.id},
        Sort:           []*mailcap.Comparator{{Property: "receivedAt", IsAscending: false}},
        Position:       int64(offset),
        Limit:          uint64(limit),
        CalculateTotal: true,
    })

    resp, err := b.client.Do(req)
    if err != nil {
        return nil, 0, fmt.Errorf("query folder: %w", err)
    }
    queryResult, err := decodeEmailQuery(resp, queryCallID)
    if err != nil {
        return nil, 0, fmt.Errorf("query folder: %w", err)
    }

    uids := make([]mail.UID, 0, len(queryResult.IDs))
    for _, id := range queryResult.IDs {
        uids = append(uids, mail.UID(id))
    }
    return uids, int(queryResult.Total), nil
}
```

(`mailcap` is whatever import path rockorager/go-jmap uses for
`mail` capability methods — read `mail/mail.go` in the vendored
package to settle the actual symbol names. The example sketch may
need adjusting; the structure is correct.)

- [ ] **Step 2: Tests**

Drive `QueryFolder` against `fakeClient` with a scripted `Email/query`
response. Assert offset/limit translate, total propagates,
unknown-folder errors surface, IDs round-trip into `mail.UID`.

- [ ] **Step 3: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: mailjmap ListFolders/OpenFolder/QueryFolder"
```

---

### Task 11: Implement `FetchHeaders`

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Modify: `internal/mailjmap/jmap_test.go`

- [ ] **Step 1: Implement**

`FetchHeaders` issues `Email/get` with a minimal `properties` set
covering `id`, `blobId`, `subject`, `from`, `receivedAt`, `keywords`,
`size`, `inReplyTo`, `threadId`. Translate each `*Email` into
`mail.MessageInfo`; capture `blobId` into `b.blobIDs[uid]`.

```go
var headerProperties = []string{
    "id", "blobId", "subject", "from", "receivedAt",
    "keywords", "size", "inReplyTo", "threadId",
}

func (b *Backend) FetchHeaders(uids []mail.UID) ([]mail.MessageInfo, error) {
    if len(uids) == 0 {
        return nil, nil
    }
    ids := make([]jmap.ID, 0, len(uids))
    for _, u := range uids {
        ids = append(ids, jmap.ID(u))
    }
    req := &jmap.Request{Using: []string{"urn:ietf:params:jmap:mail"}}
    callID := req.Invoke(&mailcap.EmailGet{
        AccountID:  b.accountID(),
        IDs:        ids,
        Properties: headerProperties,
    })
    resp, err := b.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch headers: %w", err)
    }
    emails, err := decodeEmailGet(resp, callID)
    if err != nil {
        return nil, fmt.Errorf("fetch headers: %w", err)
    }

    b.mu.Lock()
    out := make([]mail.MessageInfo, 0, len(emails))
    for _, e := range emails {
        b.blobIDs[mail.UID(e.ID)] = string(e.BlobID)
        out = append(out, translateEmail(e))
    }
    b.mu.Unlock()
    return out, nil
}

func translateEmail(e *mailcap.Email) mail.MessageInfo {
    return mail.MessageInfo{
        UID:       mail.UID(e.ID),
        Subject:   e.Subject,
        From:      formatFromList(e.From),
        SentAt:    e.ReceivedAt.Time,
        Flags:     translateKeywords(e.Keywords),
        Size:      uint32(e.Size),
        ThreadID:  mail.UID(e.ThreadID),
        InReplyTo: firstInReplyTo(e.InReplyTo),
    }
}
```

`formatFromList` joins the address list as `"Display Name"` (or the
email address if no display name); `translateKeywords` maps JMAP
keywords (`$seen`, `$flagged`, `$answered`, `$draft`, `$forwarded`)
to `mail.Flag` bits; `firstInReplyTo` returns the first message-id
header value or empty.

- [ ] **Step 2: Tests**

Scripted `Email/get` response with two emails: one fully-flagged
seen+answered, one unread+flagged. Assert flags translate, blobIDs
land in the map, ThreadID populates.

- [ ] **Step 3: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: mailjmap FetchHeaders"
```

---

### Task 12: Implement `FetchBody` with LRU + singleflight

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Modify: `internal/mailjmap/jmap_test.go`

- [ ] **Step 1: Implement**

```go
import "golang.org/x/sync/singleflight"

// In Backend struct:
//   bodyGroup singleflight.Group

func (b *Backend) FetchBody(uid mail.UID) (io.Reader, error) {
    b.mu.Lock()
    blobID, ok := b.blobIDs[uid]
    cache := b.bodies
    b.mu.Unlock()
    if !ok {
        return nil, fmt.Errorf("fetch body: unknown uid %q (call FetchHeaders first)", uid)
    }
    if buf, hit := cache.Get(blobID); hit {
        return bytes.NewReader(buf), nil
    }

    v, err, _ := b.bodyGroup.Do(blobID, func() (any, error) {
        // Re-check after singleflight serializes — another caller may
        // have populated the cache.
        if buf, hit := cache.Get(blobID); hit {
            return buf, nil
        }
        buf, err := b.downloadBlob(blobID)
        if err != nil {
            return nil, err
        }
        cache.Add(blobID, buf)
        return buf, nil
    })
    if err != nil {
        return nil, fmt.Errorf("fetch body: %w", err)
    }
    return bytes.NewReader(v.([]byte)), nil
}

func (b *Backend) downloadBlob(blobID string) ([]byte, error) {
    // Use the JMAP download endpoint from the session. Reads the
    // raw RFC822 body. Authorization header set via the same bearer
    // token used for Connect.
    url := b.session.DownloadURL(b.accountID(), blobID, "blob")
    req, _ := http.NewRequest(http.MethodGet, url, nil)
    req.Header.Set("Authorization", "Bearer "+b.cfg.Password)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("blob download: status %d", resp.StatusCode)
    }
    return io.ReadAll(resp.Body)
}
```

(Adjust `session.DownloadURL` to the actual API in
rockorager/go-jmap — may be `Session.DownloadEndpoint(accountID,
blobID, name, type)` or similar. Check `session.go`.)

- [ ] **Step 2: Tests**

Two cases:
1. Cache miss → calls downloader, populates cache, second call hits
   cache (downloader called once).
2. Singleflight collapse: spawn 10 goroutines calling FetchBody for
   the same UID concurrently against a downloader that records its
   call count; assert exactly one invocation.

Use a `bodyDownloader` interface seam to swap `downloadBlob` for a
fake in tests, similar to the `jmapClient` seam.

- [ ] **Step 3: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: mailjmap FetchBody with LRU + singleflight"
```

---

### Task 13: Implement push loop and `UpdateConnState` emission

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Create: `internal/mailjmap/push.go`
- Modify: `internal/mailjmap/jmap_test.go`

- [ ] **Step 1: Wire push loop start into Connect**

At the end of `Connect`, before `return nil`:

```go
ctx, cancel := context.WithCancel(context.Background())
b.pushCancel = cancel
b.pushDone = make(chan struct{})
go b.pushLoop(ctx)
```

- [ ] **Step 2: Implement `pushLoop` in `push.go`**

```go
package mailjmap

import (
    "context"
    "time"

    "github.com/glw907/poplar/internal/mail"
)

const (
    pushBackoffInitial = 1 * time.Second
    pushBackoffMax     = 30 * time.Second
)

func (b *Backend) pushLoop(ctx context.Context) {
    defer close(b.pushDone)

    backoff := pushBackoffInitial
    for {
        if ctx.Err() != nil {
            return
        }
        err := b.runEventSource(ctx)
        if ctx.Err() != nil {
            return
        }
        if err != nil {
            b.emit(mail.Update{Type: mail.UpdateConnState, ConnState: mail.ConnReconnecting})
            select {
            case <-ctx.Done():
                return
            case <-time.After(backoff):
            }
            backoff *= 2
            if backoff > pushBackoffMax {
                backoff = pushBackoffMax
            }
            continue
        }
        backoff = pushBackoffInitial
    }
}

// runEventSource opens the JMAP push EventSource and blocks
// reading events. On clean shutdown (ctx cancelled) returns nil; on
// transport failure returns the error to the loop for retry.
func (b *Backend) runEventSource(ctx context.Context) error {
    // Use rockorager/go-jmap's push subpackage — see
    // ~/go/pkg/mod/git.sr.ht/~rockorager/go-jmap@v0.5.3/core/ for
    // the EventSource client API. On open success, emit
    // ConnConnected. On each StateChange, dispatch to
    // handleStateChange.
    // ...
}

func (b *Backend) handleStateChange(typ string, newState string) {
    b.mu.Lock()
    old := b.states[typ]
    b.states[typ] = newState
    b.mu.Unlock()
    if old == newState {
        return
    }
    switch typ {
    case "Email":
        b.dispatchEmailChanges()
    case "Mailbox":
        b.dispatchMailboxChanges()
    }
}

// emit sends an update non-blockingly. If the buffer is full, drop
// the message and continue — the consumer is unhealthy. Logging
// goes through fmt.Fprintln(os.Stderr) per project convention.
func (b *Backend) emit(u mail.Update) {
    b.mu.Lock()
    ch := b.updates
    b.mu.Unlock()
    if ch == nil {
        return
    }
    select {
    case ch <- u:
    default:
        fmt.Fprintln(os.Stderr, "mailjmap: dropped update, buffer full")
    }
}
```

- [ ] **Step 3: Implement `dispatchEmailChanges`**

`Email/changes` from the stored cursor → list of created, updated,
destroyed IDs. For each non-empty bucket, emit one `mail.Update`:

- created → `UpdateNewMail` with the created UIDs.
- updated → `UpdateFlagsChanged` with the updated UIDs (run a
  follow-up `Email/get` to refresh `b.blobIDs` and any cached
  headers — the actual UI refresh is out of scope for Pass 3 but
  the blobID needs to stay current for FetchBody).
- destroyed → `UpdateExpunge` with the destroyed UIDs.

`dispatchMailboxChanges` issues `Mailbox/changes` and emits
`UpdateFolderInfo` per affected mailbox name.

- [ ] **Step 4: Tests**

The push loop is tricky to test end-to-end without a fake
EventSource. Cover the deterministic pieces:

1. `handleStateChange` dedup: same state string in twice ⇒ second
   call is a no-op (no outgoing emits).
2. Cursor advance: after a successful `dispatchEmailChanges`, the
   stored state matches the value passed in; on dispatcher error,
   the old state is preserved.
3. `emit` buffer-full drop: fill the channel, call emit, assert no
   blocking and no panic.
4. ConnState emission shape: simulated runEventSource returning
   nil-then-error sequences via a swappable `runEventSource` field
   (test seam) emits ConnReconnecting on each error.

Use a `runEventSourceFunc` field on `Backend` (default points to
the real `runEventSource`) so tests can swap it.

- [ ] **Step 5: Pre-seed states in Connect**

Before spawning the push loop, run `Email/queryChanges` (or
`Email/get` with `properties=[id]` returning the current state
string) and `Mailbox/get` (already done in `refreshFolders`) to
seed `b.states["Email"]` and `b.states["Mailbox"]`. This avoids the
first push event flooding the UI with "everything changed."

- [ ] **Step 6: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: mailjmap push loop + UpdateConnState emission"
```

---

### Task 14: Implement mutating RPC methods

**Files:**
- Modify: `internal/mailjmap/jmap.go`
- Modify: `internal/mailjmap/jmap_test.go`

`Move`, `Copy`, `Delete`, `Flag`, `MarkRead`, `MarkAnswered`, plus a
`Send` stub.

- [ ] **Step 1: Implement**

Each mutator builds an `Email/set` invocation. `MarkRead` /
`MarkAnswered` set `keywords/$seen` / `keywords/$answered` to true
on every UID in one batched call. `Flag` toggles `keywords/$flagged`
based on the `set` arg. `Delete` is a soft delete: move every UID
into the Trash mailbox by patching `mailboxIds` (set Trash id true,
all others false). `Move` patches `mailboxIds`. `Copy` uses
`Email/copy`.

```go
func (b *Backend) MarkRead(uids []mail.UID) error {
    return b.setKeyword(uids, "$seen", true)
}

func (b *Backend) MarkAnswered(uids []mail.UID) error {
    return b.setKeyword(uids, "$answered", true)
}

func (b *Backend) Flag(uids []mail.UID, flag mail.Flag, set bool) error {
    keyword, err := keywordForFlag(flag)
    if err != nil {
        return err
    }
    return b.setKeyword(uids, keyword, set)
}

func (b *Backend) setKeyword(uids []mail.UID, keyword string, set bool) error {
    if len(uids) == 0 {
        return nil
    }
    update := make(map[jmap.ID]any, len(uids))
    for _, u := range uids {
        update[jmap.ID(u)] = map[string]any{
            "keywords/" + keyword: jmapBool(set),
        }
    }
    req := &jmap.Request{Using: []string{"urn:ietf:params:jmap:mail"}}
    callID := req.Invoke(&mailcap.EmailSet{
        AccountID: b.accountID(),
        Update:    update,
    })
    resp, err := b.client.Do(req)
    if err != nil {
        return fmt.Errorf("set keyword %s: %w", keyword, err)
    }
    if err := checkEmailSet(resp, callID); err != nil {
        return fmt.Errorf("set keyword %s: %w", keyword, err)
    }
    return nil
}
```

`jmapBool` returns `true` or removes the key entirely (JMAP
patch semantics). `checkEmailSet` reads `notUpdated` and surfaces
the first error.

- [ ] **Step 2: `Send` stub**

```go
func (b *Backend) Send(_ string, _ []string, _ io.Reader) error {
    return errors.New("send not implemented in pass 3 — see pass 9")
}
```

- [ ] **Step 3: `Search` stub**

For Pass 3 wire `Search` to a basic `Email/query` with a `text`
filter, or leave it returning `nil, nil` — sidebar search filters
in-memory in Pass 2.5b-7, so the backend method is unused. Pick
the simpler "return nil, nil" with a comment until Pass 6's
server-side search lands.

- [ ] **Step 4: Tests**

Table-driven over the keyword setters. Assert the request body
shape (one `Email/set` invocation with the right Update map),
notUpdated error surfacing, empty-uid no-op.

- [ ] **Step 5: `make check` + commit**

```
git add internal/mailjmap
git commit -m "Pass 3: mailjmap mutating RPC methods"
```

---

## Phase 3 — UI plumbing

The backend is now feature-complete. The UI gains lazy-load,
connection-state pumping, and a window counter. The viewer is
unchanged this pass.

### Task 15: `MessageList.AppendMessages` + `IsNearBottom`

**Files:**
- Modify: `internal/ui/msglist.go`
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Add helpers**

```go
// cursorUID returns the UID under the cursor or empty if rows is
// empty. Used as an anchor across rebuild.
func (m *MessageList) cursorUID() mail.UID {
    if len(m.rows) == 0 || m.cursor >= len(m.rows) {
        return ""
    }
    return m.rows[m.cursor].uid
}

// snapToUID positions the cursor on the row whose UID matches uid.
// Falls back to clamp at len(rows)-1 when not found.
func (m *MessageList) snapToUID(uid mail.UID) {
    if uid == "" || len(m.rows) == 0 {
        m.cursor = 0
        return
    }
    for i, r := range m.rows {
        if r.uid == uid {
            m.cursor = i
            return
        }
    }
    m.cursor = len(m.rows) - 1
}

// IsNearBottom reports whether the cursor is within k rows of the
// last row. Used by AccountTab to trigger lazy-load before the
// user runs out of messages.
func (m *MessageList) IsNearBottom(k int) bool {
    return len(m.rows) > 0 && m.cursor >= len(m.rows)-k
}

// AppendMessages adds extra to the message list, re-runs the
// group→sort→flatten pipeline, and restores the cursor by UID.
// Used for lazy-loading the next window of a large folder. Safe
// against duplicate UIDs — the existing rebuild dedups on UID.
func (m *MessageList) AppendMessages(extra []mail.MessageInfo) {
    cursorUID := m.cursorUID()
    m.source = append(m.source, extra...)
    m.now = time.Now()
    m.rebuild()
    m.snapToUID(cursorUID)
}
```

(The exact field names — `m.rows`, `m.cursor`, `m.source`, `m.now`,
`r.uid` — must match the existing `MessageList` struct. Read
`msglist.go` and adjust.)

- [ ] **Step 2: Tests**

```go
func TestMessageList_AppendMessages_PreservesCursor(t *testing.T) {
    m := newMessageListWithMessages(testMessages(0, 50))
    m.cursor = 30
    cursorUID := m.rows[30].uid

    m.AppendMessages(testMessages(50, 100))

    if m.rows[m.cursor].uid != cursorUID {
        t.Errorf("cursor moved off uid %s", cursorUID)
    }
    if len(m.rows) <= 50 {
        t.Errorf("rows not appended: got %d", len(m.rows))
    }
}

func TestMessageList_IsNearBottom(t *testing.T) {
    m := newMessageListWithMessages(testMessages(0, 100))
    m.cursor = 99
    if !m.IsNearBottom(5) {
        t.Error("cursor=99 of 100 should be near bottom (k=5)")
    }
    m.cursor = 50
    if m.IsNearBottom(5) {
        t.Error("cursor=50 of 100 should not be near bottom (k=5)")
    }
}
```

- [ ] **Step 3: `make check` + commit**

```
git add internal/ui/msglist.go internal/ui/msglist_test.go
git commit -m "Pass 3: MessageList AppendMessages + IsNearBottom"
```

---

### Task 16: `AccountTab` pagination state and load-more flow

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/cmds.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Add the per-folder page state**

```go
// folderPage tracks lazy-load state for one folder.
type folderPage struct {
    loaded           int
    total            int
    loadMoreInFlight bool
}

// On AccountTab struct, add:
//   pages map[string]*folderPage
```

Initialize `pages` in `NewAccountTab` to an empty map.

- [ ] **Step 2: Replace `loadFolderCmd` with a Query+Headers chain**

In `cmds.go`:

```go
const initialWindow = 500

// folderQueryDoneMsg carries a Query result; AccountTab follows up
// with FetchHeadersCmd to materialize the headers.
type folderQueryDoneMsg struct {
    name  string
    uids  []mail.UID
    total int
    reset bool // true on initial load, false on append
}

// headersAppliedMsg is the terminal message of an initial folder
// load: the rows are now in the message list.
type headersAppliedMsg struct {
    name string
    msgs []mail.MessageInfo
}

// headersAppendedMsg is the terminal message of a load-more.
type headersAppendedMsg struct {
    name string
    msgs []mail.MessageInfo
}

func openFolderCmd(b mail.Backend, name string) tea.Cmd {
    return func() tea.Msg {
        if err := b.OpenFolder(name); err != nil {
            return ErrorMsg{Op: "open folder", Err: err}
        }
        uids, total, err := b.QueryFolder(name, 0, initialWindow)
        if err != nil {
            return ErrorMsg{Op: "query folder", Err: err}
        }
        return folderQueryDoneMsg{name: name, uids: uids, total: total, reset: true}
    }
}

func loadMoreCmd(b mail.Backend, name string, offset int) tea.Cmd {
    return func() tea.Msg {
        uids, total, err := b.QueryFolder(name, offset, initialWindow)
        if err != nil {
            return ErrorMsg{Op: "load more", Err: err}
        }
        return folderQueryDoneMsg{name: name, uids: uids, total: total, reset: false}
    }
}

func fetchHeadersCmd(b mail.Backend, name string, uids []mail.UID, reset bool) tea.Cmd {
    return func() tea.Msg {
        msgs, err := b.FetchHeaders(uids)
        if err != nil {
            return ErrorMsg{Op: "fetch headers", Err: err}
        }
        if reset {
            return headersAppliedMsg{name: name, msgs: msgs}
        }
        return headersAppendedMsg{name: name, msgs: msgs}
    }
}
```

Delete the old `loadFolderCmd` once nothing references it.

- [ ] **Step 3: Wire AccountTab Update**

```go
case folderQueryDoneMsg:
    page := m.pageFor(msg.name)
    page.total = msg.total
    if !msg.reset {
        page.loadMoreInFlight = true
    }
    return m, fetchHeadersCmd(m.backend, msg.name, msg.uids, msg.reset)

case headersAppliedMsg:
    page := m.pageFor(msg.name)
    page.loaded = len(msg.msgs)
    m.msgList.SetMessages(msg.msgs)
    return m, nil

case headersAppendedMsg:
    page := m.pageFor(msg.name)
    page.loaded += len(msg.msgs)
    page.loadMoreInFlight = false
    m.msgList.AppendMessages(msg.msgs)
    return m, nil
```

After every cursor-moving key handler, call:

```go
if cmd := m.maybeLoadMore(); cmd != nil {
    cmds = append(cmds, cmd)
}
```

with:

```go
const loadMoreTrigger = 20

func (m *AccountTab) maybeLoadMore() tea.Cmd {
    page := m.pageFor(m.currentFolder())
    if page.loadMoreInFlight || page.loaded >= page.total {
        return nil
    }
    if !m.msgList.IsNearBottom(loadMoreTrigger) {
        return nil
    }
    page.loadMoreInFlight = true
    return loadMoreCmd(m.backend, m.currentFolder(), page.loaded)
}

func (m *AccountTab) pageFor(name string) *folderPage {
    if m.pages[name] == nil {
        m.pages[name] = &folderPage{}
    }
    return m.pages[name]
}
```

Replace existing `OpenFolder` Cmd dispatches with `openFolderCmd`.

- [ ] **Step 4: Tests**

Drive AccountTab against MockBackend (which now satisfies
QueryFolder). Assert:

1. Initial open dispatches one fetchHeadersCmd of the right size.
2. Cursor near bottom dispatches loadMoreCmd with correct offset.
3. While `loadMoreInFlight`, near-bottom does **not** redispatch.
4. When `loaded == total`, near-bottom does not dispatch.

The MockBackend has only ~14 messages; either grow it for tests or
craft a synthetic AccountTab fixture with a fake backend that
returns 1500 messages. The synthetic-fixture path is cleaner.

- [ ] **Step 5: `make check` + commit**

```
git add internal/ui
git commit -m "Pass 3: AccountTab lazy-load pagination flow"
```

---

### Task 17: Wire `UpdateConnState` into the App pump

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/cmds.go`
- Modify: `internal/ui/app_test.go`

The App needs a goroutine-style pump translating the backend's
`Updates()` channel into bubbletea Msgs. This is the first
backend→App async path in poplar.

- [ ] **Step 1: Add the pump Cmd**

```go
// backendUpdateMsg wraps a single mail.Update in a tea.Msg.
type backendUpdateMsg struct{ update mail.Update }

// pumpUpdatesCmd waits for one mail.Update, returns it as a
// backendUpdateMsg, and re-arms itself. App's Update loop is
// responsible for calling this Cmd again so the pump stays alive.
func pumpUpdatesCmd(b mail.Backend) tea.Cmd {
    return func() tea.Msg {
        u, ok := <-b.Updates()
        if !ok {
            return backendUpdateMsg{update: mail.Update{Type: mail.UpdateConnState, ConnState: mail.ConnOffline}}
        }
        return backendUpdateMsg{update: u}
    }
}
```

- [ ] **Step 2: Start the pump in `App.Init`**

```go
func (m App) Init() tea.Cmd {
    return tea.Batch(m.acct.Init(), pumpUpdatesCmd(m.backend))
}
```

This requires App to hold a `backend mail.Backend` field — add it,
populated from `NewApp`.

- [ ] **Step 3: Handle backendUpdateMsg in App.Update**

```go
case backendUpdateMsg:
    cmds := []tea.Cmd{pumpUpdatesCmd(m.backend)} // re-arm
    if msg.update.Type == mail.UpdateConnState {
        m.statusBar = m.statusBar.SetConnectionState(translateConnState(msg.update.ConnState))
    }
    // Other Update types delegate to AccountTab in a later pass.
    return m, tea.Batch(cmds...)
```

with:

```go
func translateConnState(s mail.ConnState) ConnectionState {
    switch s {
    case mail.ConnConnected:
        return Connected
    case mail.ConnReconnecting:
        return Reconnecting
    default:
        return Offline
    }
}
```

- [ ] **Step 4: Default to Offline at startup**

In `NewApp`, change the initial `SetConnectionState(Connected)` to
`SetConnectionState(Offline)`. The first `ConnConnected` from the
push loop flips it to green.

- [ ] **Step 5: Tests**

`App.Update` with a synthetic `backendUpdateMsg{Type: ConnState,
ConnState: ConnConnected}` flips the status bar; with
`ConnReconnecting` shows reconnecting; with the channel closing
(simulated by a backendUpdateMsg carrying `ConnOffline` per the
pumpUpdatesCmd contract), shows offline.

- [ ] **Step 6: `make check` + commit**

```
git add internal/ui
git commit -m "Pass 3: pump backend Updates into App, wire ConnState"
```

---

### Task 18: Window counter footer hint

**Files:**
- Modify: `internal/ui/footer.go`
- Modify: `internal/ui/account_tab.go` (or wherever footer hints feed in)

- [ ] **Step 1: Add a low-rank hint**

Footer hints are populated from the active context. Find the
account-context hint table and add an entry whose text comes from
AccountTab state:

```go
{Key: "", Desc: m.windowCounter(), Rank: 8}
```

with:

```go
func (m AccountTab) windowCounter() string {
    page := m.pages[m.currentFolder()]
    if page == nil || page.total == 0 || page.loaded >= page.total {
        return ""
    }
    return fmt.Sprintf("%d/%d", page.loaded, page.total)
}
```

When the counter is empty, the hint row should drop entirely
(empty Desc is treated as no-hint).

- [ ] **Step 2: Test**

Render footer with a fake AccountTab where `loaded < total`; assert
the counter text appears. Render with `loaded == total`; assert no
counter row.

- [ ] **Step 3: `make check` + commit**

```
git add internal/ui
git commit -m "Pass 3: footer window counter hint"
```

---

### Task 19: Loading spinner on folder open

**Files:**
- Modify: `internal/ui/msglist.go` (or `account_tab.go` depending on where the panel is rendered)
- Modify: corresponding `_test.go`

While the initial `openFolderCmd` is running, the account panel
should show a "Loading messages…" placeholder built via
`NewSpinner(theme)` (Pass 2.5b-6 helper).

- [ ] **Step 1: Add a `loading bool` flag on AccountTab (or
  MessageList) plus a Spinner**

Set true on `openFolderCmd` dispatch, false on `headersAppliedMsg`.
Render the spinner placeholder when `loading && len(rows)==0`.

- [ ] **Step 2: Drive the spinner via tea.Tick**

The shared `NewSpinner` produces a `bubbles/spinner.Model`. Its
`Tick` Cmd advances the frame; route that into Update.

- [ ] **Step 3: Tests**

Render with `loading=true, rows=nil` — assert the spinner glyph
appears in the rendered string. Toggle to `loading=false` — assert
it's gone.

- [ ] **Step 4: `make check` + commit**

```
git add internal/ui
git commit -m "Pass 3: spinner placeholder during folder open"
```

---

## Phase 4 — Cleanup and live wiring

### Task 20: Delete `internal/mailworker/`

**Files:**
- Delete: `internal/mailworker/` (entire tree)

- [ ] **Step 1: Confirm nothing imports it**

```
grep -rn "internal/mailworker" --include='*.go' .
```

Expected output: only matches inside `internal/mailworker/` itself
(if any). If anything in `internal/`, `cmd/`, or test files still
references it, fix those before deleting.

- [ ] **Step 2: Delete and tidy**

```
git rm -r internal/mailworker/
go mod tidy
```

`go mod tidy` likely drops a handful of fork-only transitive deps.
Review the diff before committing.

- [ ] **Step 3: `make check`**

```
make check
```

Expected: green. If anything fails, the cause is a stale import
that step 1 missed.

- [ ] **Step 4: Commit**

```
git add -A
git commit -m "Pass 3: delete internal/mailworker (replaced by direct mailjmap)"
```

---

### Task 21: Wire live JMAP in `cmd/poplar`

**Files:**
- Modify: `cmd/poplar/root.go`
- Modify: `cmd/poplar/config_init.go` (already has `openBackendForInit`; reuse the pattern)

- [ ] **Step 1: Add a backend-builder switch**

```go
func openBackend(acct config.AccountConfig) (mail.Backend, error) {
    switch acct.Protocol {
    case "mock", "":
        return mail.NewMockBackend(), nil
    case "jmap":
        return mailjmap.New(acct), nil
    case "imap":
        return nil, fmt.Errorf("imap backend lands in pass 8")
    default:
        return nil, fmt.Errorf("unknown protocol %q for account %q", acct.Protocol, acct.Name)
    }
}
```

Place this in a new file `cmd/poplar/backend.go` if one does not
already exist.

- [ ] **Step 2: Use it in root.go**

Replace the current hardcoded `mail.NewMockBackend()` in `root.go`
with:

```go
accts, err := config.ParseAccounts(/* config path */)
if err != nil {
    return fmt.Errorf("load accounts: %w", err)
}
if len(accts) == 0 {
    return fmt.Errorf("no accounts configured; see docs/poplar/config.md")
}
backend, err := openBackend(accts[0])
if err != nil {
    return fmt.Errorf("open backend: %w", err)
}
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
if err := backend.Connect(ctx); err != nil {
    return fmt.Errorf("connect: %w", err)
}
defer backend.Disconnect()
```

For the v1-trigger pass we keep "first account wins" — multi-account
tabs come in Pass 11.

A `--account` flag falling back to the first configured account is
optional polish; if it lands here, mention in the spec / next
starter prompt.

- [ ] **Step 3: Smoke build**

```
make build
./poplar --help
```

(`./poplar` itself will need a real `accounts.toml` to fully run —
that's the live-integration step.)

- [ ] **Step 4: Commit**

```
git add cmd/poplar
git commit -m "Pass 3: wire live JMAP backend selection in root.go"
```

---

### Task 22: Live-integration verification

**Files:** none (manual run + status update)

- [ ] **Step 1: Configure**

Edit `~/.config/poplar/accounts.toml`:

```toml
[[account]]
name = "geoff@907.life"
protocol = "jmap"
host = "https://api.fastmail.com/jmap/session"
username = "geoff@907.life"
password = "$FASTMAIL_API_TOKEN"
```

Confirm `$FASTMAIL_API_TOKEN` is exported (it lives in
`~/.local/secrets`; `source ~/.bashrc` if needed).

- [ ] **Step 2: Install and run**

```
make install
poplar
```

- [ ] **Step 3: Verify the acceptance checklist**

- Status bar shows `●` within ~5 seconds of launch.
- Inbox loads with up to 500 messages.
- Scrolling to within 20 of the bottom triggers a load-more; the
  footer shows `500/2,xxx`-style progression and the counter
  updates.
- Opening a message renders the body via the existing viewer and
  marks it read.
- Sending an email from another client (phone or web) surfaces a
  new row within seconds (push event).
- Disconnecting the network flips the indicator to `◐`; reconnect
  flips back to `●`.
- Errors (wrong token, unreachable host) surface in the banner
  with the expected `Op` verb-phrase.

- [ ] **Step 4: Capture any regressions**

Anything that fails goes back through the relevant task — do not
mark Pass 3 done with known regressions.

---

## Phase 5 — Pass-end consolidation

### Task 23: ADRs, invariants, archive, ship

**Files:**
- Create: `docs/poplar/decisions/0077-jmap-direct-implementation.md` (and any further ADRs for individual decisions made during implementation that diverge from the spec)
- Modify: `docs/poplar/invariants.md` (in place)
- Modify: `docs/poplar/STATUS.md`
- Move: this plan and the spec to `docs/superpowers/archive/`

Run the `poplar-pass` skill's pass-end consolidation ritual:

- [ ] **Step 1: Run `/simplify`**
- [ ] **Step 2: Write ADR(s) for design decisions made during the
  pass that aren't already covered by ADR-0075 (e.g. the
  `pumpUpdatesCmd` re-arm pattern, the `jmapClient` test seam, the
  500-message initial window if it's a load-bearing detail).**
- [ ] **Step 3: Update `docs/poplar/invariants.md` in place — add
  facts about the direct-on-rockorager implementation, body LRU,
  push goroutine, ConnState pumping, lazy-load. Remove obsolete
  `mailworker/` references. Update the decision-index table.**
- [ ] **Step 4: Update `STATUS.md` — mark Pass 3 done, write the
  Pass 6 (or whichever is next) starter prompt.**
- [ ] **Step 5: `git mv` plan + spec into
  `docs/superpowers/archive/`.**
- [ ] **Step 6: `make check` green; commit; push; `make install`.**

---

## Self-review

**Spec coverage walkthrough:**

| Spec section | Tasks |
|---|---|
| `mail.Backend` additions (QueryFolder, ConnState, UpdateConnState) | 1, 2 |
| `internal/mailauth/` (xoauth2, keepalive) | 5, 6 |
| `internal/mailjmap/` (rewrite + push + LRU + singleflight) | 7–14 |
| `internal/mailworker/` deletion | 20 |
| `MessageList.AppendMessages` + `IsNearBottom` | 15 |
| `AccountTab` lazy-load with `pages`, in-flight guard, footer counter | 16, 18 |
| App-level `connState` pump | 17 |
| Loading spinner | 19 |
| Live wiring + acceptance | 21, 22 |
| Pass-end ritual | 23 |

**Placeholder check:** No "TBD" / "implement later" / "appropriate
error handling" instances. Where a library symbol's exact name
isn't pinned (e.g. the rockorager/go-jmap mail capability path),
the task explicitly directs the reader to the vendored module
source. That counts as a known unknown the engineer resolves at
implementation time, not a placeholder.

**Type consistency:** `folderPage`, `folderEntry`, `jmapClient`,
`pumpUpdatesCmd`, `backendUpdateMsg`, `headersAppliedMsg`,
`headersAppendedMsg`, `folderQueryDoneMsg`, `openFolderCmd`,
`loadMoreCmd`, `fetchHeadersCmd`, `maybeLoadMore`, `pageFor`,
`windowCounter` are introduced once and reused consistently.
`mail.UpdateConnState` / `mail.ConnState` / `ConnOffline` /
`ConnReconnecting` / `ConnConnected` match the const-block ordering
defined in Task 2. `ConnectionState` (UI side) /
`Connected/Offline/Reconnecting` (UI side) match the existing
`status_bar.go` enum.

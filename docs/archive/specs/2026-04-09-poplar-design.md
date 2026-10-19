# Poplar: Bubbletea Email Client

**Date:** 2026-04-09
**Status:** Design approved

## Summary

Poplar is a bubbletea-based terminal email client that replaces aerc's
UI while reusing its mail worker code. It lives in the beautiful-aerc
monorepo, shares the existing filter/theme/compose infrastructure, and
uses forked aerc worker code for IMAP and JMAP backends. Target
providers: Gmail (IMAP) and Fastmail (JMAP).

## Motivation

- Nicer-looking TUI than aerc's tcell-based interface
- Bubbletea is more accessible to contributors than aerc's custom UI
  framework
- Full control over the UI layer while leveraging battle-tested mail
  protocol code
- External-editor-only compose (nvim-mail) simplifies the design

## Architecture

### Approach: Clean Fork with Adapter Layer

Fork aerc's `worker/`, `models/`, and required `lib/` packages into
the monorepo. Define a `mail.Backend` interface that wraps the forked
worker message-passing system. The bubbletea UI talks only to the
adapter interface, never to aerc/worker types directly.

Aerc is MIT-licensed. Keep the copyright notice in forked files.

Track aerc upstream as a git remote for cherry-picking protocol-level
bug fixes and improvements.

### Extraction Changes

The forked worker code requires these modifications:

- **`config.AccountConfig`**: Replace aerc's 30-field struct with
  poplar's own account config struct. Workers only need: Source URL,
  params map, credential command, headers include/exclude, identity
  fields.
- **`types.WorkerMessages` global channel**: Replace with a per-worker
  channel injected via constructor.
- **`config.Ui` reference in maildir threading**: Not relevant for
  v1 (IMAP + JMAP only). If maildir is added later, replace with
  parameters passed at call time.

### Monorepo Structure

```
beautiful-aerc/
  cmd/mailrender/        existing
  cmd/fastmail-cli/      existing
  cmd/tidytext/          existing
  cmd/poplar/            new binary
  internal/
    filter/              existing — used by mailrender + poplar viewer
    theme/               existing — used by everyone
    compose/             existing — used by mailrender + poplar compose
    tidy/                existing
    jmap/                existing (fastmail-cli's JMAP client)
    header/              existing
    rules/               existing
    mail/                new: Backend interface + adapters
    worker/              new: forked aerc workers (IMAP, JMAP)
    models/              new: forked aerc models
    ui/                  new: bubbletea components
  .config/aerc/          existing config + themes
```

Poplar calls `internal/filter` directly as a library for message
rendering — no subprocess overhead. Same for compose reflow and
theme loading.

The Makefile produces four binaries: mailrender, fastmail-cli,
tidytext, poplar. `make install` puts all four in `~/.local/bin/`.

Long-term: mailrender either becomes a thin CLI wrapper around the
filter library or is retired entirely once poplar is the daily driver.

## Mail Backend Interface

```go
// internal/mail/backend.go

type Backend interface {
    Connect(ctx context.Context) error
    Disconnect() error

    ListFolders() ([]Folder, error)
    OpenFolder(name string) error

    FetchHeaders(uids []UID) ([]Message, error)
    FetchBody(uid UID) (*FullMessage, error)

    Search(criteria SearchCriteria) ([]UID, error)

    Move(uids []UID, dest string) error
    Copy(uids []UID, dest string) error
    Delete(uids []UID) error
    Flag(uids []UID, flag Flag, set bool) error
    MarkRead(uids []UID) error
    MarkAnswered(uids []UID) error

    Send(msg io.Reader) error

    Updates() <-chan Update
}
```

- Synchronous where possible — bubbletea's `tea.Cmd` handles async
  naturally (blocking call in a command, result comes back as a
  message).
- `Updates()` channel for IDLE/JMAP push notifications (new mail,
  flag changes, expunges).
- Poplar-native types (`Message`, `Folder`, `UID`, `Flag`) defined
  in `internal/mail/message.go`. The adapter translates between these
  and forked aerc `models.*` types.

## UI Architecture

Bubbletea Elm architecture: root model owns child components, each
with their own `Update`/`View` cycle.

### Root Model (`ui/app.go`)

- Tab bar across the top
- Routes keypresses to active tab
- Holds account map, subscribes to `account.Updates()`

### Tab Types

| Tab | Components | Description |
|-----|-----------|-------------|
| Folder view | sidebar + msglist | Persistent tab per account |
| Viewer | viewer | Message body, opens on selection |
| Compose | (external) | Spawns editor via `tea.ExecProcess` |

### Tab Interface

```go
type Tab interface {
    tea.Model
    Title() string
    Closeable() bool  // false for account folder views
}
```

### Components

- **`sidebar.go`** — folder list with unread counts and Nerd Font
  icons. Keyboard navigation.
- **`msglist.go`** — message list with columns (flags, sender,
  subject, date). Thread prefixes with box-drawing characters.
  Sort, filter, scroll, selection.
- **`viewer.go`** — renders message body via `internal/filter`
  (html or plain). Scrollable viewport. Header block via
  `internal/filter` headers.
- **`statusbar.go`** — account/folder info, status messages.
- **`commandinput.go`** — ex-style `:` command mode.

### Styling

Lipgloss styles derived from theme TOML color slots. Same theme
file drives both message rendering (via filter package ANSI tokens)
and UI chrome (via lipgloss). Switching themes changes everything.

```go
type Styles struct {
    TabActive     lipgloss.Style  // accent_secondary on bg_base
    TabInactive   lipgloss.Style  // fg_dim on bg_elevated
    SidebarFolder lipgloss.Style  // fg_base
    SidebarUnread lipgloss.Style  // accent_tertiary, bold
    MsgUnread     lipgloss.Style  // accent_tertiary
    MsgRead       lipgloss.Style  // fg_dim
    MsgFlagged    lipgloss.Style  // color_warning
    Selected      lipgloss.Style  // bg_selection
    StatusBar     lipgloss.Style  // fg_bright on bg_border
    Error         lipgloss.Style  // color_error
}
```

## Configuration

Poplar has its own config at `~/.config/poplar/`:

| File | Purpose |
|------|---------|
| `accounts.toml` | Account definitions |
| `poplar.toml` | UI settings (columns, threading, sort, editor) |
| `keybindings.toml` | Context-sensitive keybindings |

Theme files are read from `~/.config/aerc/themes/` (configurable)
during coexistence. Same TOML format, same 16 slots + tokens.

### Account Config Example

```toml
[[account]]
name = "Fastmail"
backend = "jmap"
source = "jmap://geoff@907.life"
credential-cmd = "fastmail-password"
folders-sort = ["Inbox", "Notifications", "Drafts", "Sent", "Archive"]
copy-to = "Sent"

[[account]]
name = "Gmail"
backend = "imap"
source = "imaps://user@gmail.com@imap.gmail.com:993"
credential-cmd = "gmail-password"
smtp = "smtps://user@gmail.com@smtp.gmail.com:465"
smtp-credential-cmd = "gmail-password"
```

## Compose and Send

External-editor only. Flow:

1. User hits reply/compose keybinding
2. Poplar writes headers + quoted body to temp file (using
   `internal/compose` for reflow/formatting)
3. `tea.ExecProcess` launches nvim-mail — bubbletea suspends
4. Editor exits 0: parse file, show review prompt (y/n/e)
5. On `y`: add format-flowed markers, convert to HTML multipart
   via `internal/filter.ToHTML`, send via `backend.Send()`
6. Copy to Sent folder if configured
7. Editor exits non-zero: abort compose

Send method is backend-dependent:
- JMAP (Fastmail): `EmailSubmission/set` via the JMAP worker
- IMAP (Gmail): separate SMTP connection, configured per account

## Integration with Existing Tools

**Beautiful-aerc provides (unchanged during coexistence):**
- `internal/filter/` — message rendering (called as library)
- `internal/theme/` — theme loading and ANSI resolution
- `internal/compose/` — compose buffer normalization
- `fastmail-cli` binary — rules, masked email, folders
- `tidytext` binary — prose tidying
- Theme TOML files
- nvim-mail compose editor

**Poplar builds:**
- Forked mail workers with adapter interface
- Bubbletea UI
- Config system
- Compose flow (temp file, editor, review, send)
- Keybinding system

**Integration points:**
- Filter/theme/compose code used as library (same Go module)
- fastmail-cli invoked for rules/masked email via keybindings
- nvim-mail works as-is via `tea.ExecProcess`
- Themes shared from `~/.config/aerc/themes/`

## Implementation Passes

Each pass is scoped for one Claude session, ending with something
testable. Every pass finishes with the pass-end checklist:

1. `/simplify` — code quality review
2. Update `docs/architecture.md` — add design decisions made
3. Update `docs/STATUS.md` — mark pass done, update current state,
   set next starter prompt
4. Commit all changes
5. `git push`

Start a session with "next pass" or "continue development" to pick
up where the last pass left off (reads STATUS.md for context).

### Pass 1 — Scaffold + Fork
- Create `cmd/poplar/main.go` with minimal cobra root command
- Fork aerc's `worker/types/`, `models/`, `worker/jmap/`,
  `worker/imap/`, `worker/middleware/`, required `lib/` pieces
  into `internal/worker/`, `internal/models/`
- Strip `config.AccountConfig`, replace with poplar config struct
- Replace global `WorkerMessages` channel with injected channel
- Update Makefile to build poplar as fourth binary
- **Gate:** `make build` compiles all four binaries

### Pass 2 — Backend Adapter + Connect
- Define `mail.Backend` interface in `internal/mail/`
- Define poplar-native types (`Message`, `Folder`, `UID`, `Flag`)
- Write JMAP adapter wrapping the forked worker
- Write account config parser (`~/.config/poplar/accounts.toml`)
- `cmd/poplar/` connects using config, calls `ListFolders`, prints
  to stdout
- **Gate:** `poplar` connects to Fastmail, prints folder list, exits

### Pass 3 — Bubbletea Shell
- Add bubbletea + lipgloss + bubbles dependencies
- Root model (`ui/app.go`) with tab bar
- Tab interface with `Title()` and `Closeable()`
- Sidebar component showing live folder list from backend
- Theme-to-lipgloss bridge (`ui/styles.go`) reading theme TOML
- Folder selection updates placeholder message area
- **Gate:** interactive TUI, navigate folders with j/k, themed

### Pass 4 — Message List
- Fetch headers for selected folder via backend adapter
- Message list component with columns (flags, sender, subject, date)
- Threading with box-drawing prefixes
- Scroll, selection, unread/read styling from theme
- Nerd Font status icons matching aerc setup
- **Gate:** browse Inbox contents with threading visible

### Pass 5 — Message Viewer
- Open selected message in a new tab (Enter key)
- Fetch body via backend adapter
- Render via `internal/filter` (html/plain) as library call
- Render headers via `internal/filter` headers
- Scrollable viewport (bubbles viewport component)
- Close tab returns to message list
- **Gate:** read email messages with full styled rendering

### Pass 6 — Triage Actions
- Delete (move to Trash), archive (move to Archive)
- Flag, mark read/unread
- Status bar component with action feedback messages
- d = delete + next, D = delete + close viewer
- a = archive + next, A = archive + close viewer
- **Gate:** triage email without leaving poplar

### Pass 7 — Command Mode + Search
- Status bar command input (`:` enters command mode)
- Parse and execute: `:move <folder>`, `:copy <folder>`,
  `:delete`, `:search <term>`, `:quit`
- Search results displayed in message list
- Escape exits command mode
- **Gate:** all command-driven operations work

### Pass 8 — Gmail IMAP
- Write IMAP adapter wrapping the forked IMAP worker
- Gmail middleware (X-GM-EXT-1) for search/threading
- Add Gmail account to test config
- Verify folder list, message fetch, threading, actions
- **Gate:** Gmail account working alongside Fastmail

### Pass 9 — Compose + Send
- New compose: write empty headers to temp file, launch nvim-mail
  via `tea.ExecProcess`
- Reply/reply-all/forward: generate headers + quoted body via
  `internal/compose`, same editor flow
- Review prompt after editor exit (y = send, n = abort, e = re-edit)
- Multipart conversion via `internal/filter.ToHTML`
- Send via `backend.Send()` (JMAP submit / SMTP)
- Copy to Sent folder if configured
- Non-zero editor exit = abort
- **Gate:** full compose, reply, and send loop working

### Pass 10 — Keybindings + Config
- Keybinding map parser (`~/.config/poplar/keybindings.toml`)
- Context-sensitive scopes (global, msglist, viewer, command)
- `poplar.toml` parser for UI settings (columns, sort, threading,
  editor, theme-dir)
- Wire config into all components
- **Gate:** keybindings and UI settings fully customizable

### Pass 11 — Polish for Daily Use
- Multi-account support (tab per account, account switcher)
- fastmail-cli integration keybindings (ff/fs/ft, md)
- Theme switching (read theme-dir from config)
- Connection error handling, automatic reconnection
- IDLE/JMAP push working (new mail appears without manual refresh)
- **Gate:** replace aerc as daily driver

### v2 Backlog
- Internalize mailrender as library (retire standalone binary)
- Move themes to `~/.config/poplar/themes/`
- Built-in link picker (bubbletea list component)
- Desktop notifications on new mail
- Contact completion (khard in command mode)
- Plugin system for custom actions
- Maildir/notmuch backends if demand exists

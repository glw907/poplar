# Poplar Chrome Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the 3-row tab bar with a 1-row top frame line, redesign the status bar as the bottom frame edge with combined status indicator, add toast notifications on the top-right, update keybindings, and add account label to the sidebar.

**Architecture:** Three-sided frame (top `──┬──╮`, right `│`, bottom status bar `──┴──╯`) with open left edge. Tab bar and Tab interface removed. App holds AccountTab directly. TopLine renders the top frame edge with toast overlay. StatusBar becomes the bottom frame edge with combined status indicator. Footer sits outside the frame with grouped keybindings separated by dim `┊`.

**Tech Stack:** Go, bubbletea, lipgloss, bubbles/help, bubbles/spinner

**Spec:** `docs/superpowers/specs/2026-04-11-poplar-chrome-redesign-design.md`
**Keybindings:** `docs/poplar/keybindings.md`

---

## File Structure

| Action | File | Responsibility |
|--------|------|----------------|
| Delete | `internal/ui/tab_bar.go` | Removed — no more tab bar |
| Delete | `internal/ui/tab_bar_test.go` | Removed |
| Delete | `internal/ui/tab.go` | Removed — Tab interface no longer needed |
| Create | `internal/ui/top_line.go` | Top frame line `──┬──╮` with toast overlay |
| Create | `internal/ui/top_line_test.go` | Tests for top line rendering and toast |
| Modify | `internal/ui/status_bar.go` | Bottom frame edge with `┴`, `─╯`, combined indicator |
| Modify | `internal/ui/status_bar_test.go` | Updated tests |
| Modify | `internal/ui/footer.go` | Grouped keybindings with `┊` separator, 1-space padding |
| Modify | `internal/ui/footer_test.go` | Updated tests |
| Modify | `internal/ui/keys.go` | New key groups: folder jump, viewer, navigation |
| Modify | `internal/ui/styles.go` | New styles: TopLine, Toast types, StatusIndicator |
| Modify | `internal/ui/account_tab.go` | Account name in sidebar, remove Tab interface compliance |
| Modify | `internal/ui/account_tab_test.go` | Updated tests |
| Modify | `internal/ui/app.go` | Hold AccountTab directly, wire TopLine, new layout |
| Modify | `internal/ui/app_test.go` | Updated tests |
| Modify | `internal/mail/backend.go` | Add `AccountName() string` method |
| Modify | `internal/mail/mock.go` | Implement `AccountName()` |
| Modify | `cmd/poplar/root.go` | Remove appModel Tab wrapping, simplify |

---

### Task 1: Add AccountName to Backend Interface

**Files:**
- Modify: `internal/mail/backend.go`
- Modify: `internal/mail/mock.go`

- [ ] **Step 1: Add AccountName to Backend interface**

```go
// In internal/mail/backend.go, add to the Backend interface:
AccountName() string
```

- [ ] **Step 2: Implement AccountName on MockBackend**

In `internal/mail/mock.go`, add a `name` field to `MockBackend` and set it in `NewMockBackend`:

```go
type MockBackend struct {
	name    string
	folders []Folder
	msgs    []MessageInfo
	updates chan Update
}

func NewMockBackend() *MockBackend {
	return &MockBackend{
		name: "geoff@907.life",
		// ... rest unchanged
	}
}

func (m *MockBackend) AccountName() string { return m.name }
```

- [ ] **Step 3: Verify build**

Run: `go build ./...`
Expected: compiles clean

- [ ] **Step 4: Commit**

```bash
git add internal/mail/backend.go internal/mail/mock.go
git commit -m "Add AccountName to Backend interface"
```

---

### Task 2: Delete Tab Bar and Tab Interface

**Files:**
- Delete: `internal/ui/tab_bar.go`
- Delete: `internal/ui/tab_bar_test.go`
- Delete: `internal/ui/tab.go`
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Remove Tab interface methods from AccountTab**

In `internal/ui/account_tab.go`, remove these methods: `Title()`, `Icon()`, `Closeable()`. Remove the `icon` and `folder` fields from the struct since the sidebar handles folder display. Keep `focused`, `styles`, `backend`, `width`, `height`.

```go
type AccountTab struct {
	styles  Styles
	backend mail.Backend
	focused Panel
	width   int
	height  int
}

func NewAccountTab(styles Styles, backend mail.Backend) AccountTab {
	return AccountTab{
		styles:  styles,
		backend: backend,
		focused: SidebarPanel,
	}
}
```

- [ ] **Step 2: Remove the Update method that returns tea.Model**

AccountTab no longer needs to satisfy `tea.Model` or the `Tab` interface. Remove the `Update(msg tea.Msg) (tea.Model, tea.Cmd)` method. Rename `updateTab` to `Update` and have it return `(AccountTab, tea.Cmd)`:

```go
func (m AccountTab) Update(msg tea.Msg) (AccountTab, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.Type == tea.KeyTab {
			if m.focused == SidebarPanel {
				m.focused = MsgListPanel
			} else {
				m.focused = SidebarPanel
			}
		}
	}
	return m, nil
}
```

- [ ] **Step 3: Move maxInt to account_tab.go**

`maxInt` is defined in `tab_bar.go` but used by StatusBar and App. Move it to `account_tab.go` (next to the existing `minInt`):

```go
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
```

- [ ] **Step 4: Delete tab files**

```bash
rm internal/ui/tab_bar.go internal/ui/tab_bar_test.go internal/ui/tab.go
```

- [ ] **Step 5: Update account_tab_test.go**

Remove any test that references `Title()`, `Icon()`, `Closeable()`, or casts to `Tab` interface. Update `Update` call sites to use the new signature `(AccountTab, tea.Cmd)`:

```go
func TestAccountTab(t *testing.T) {
	styles := NewStyles(theme.Nord)
	backend := mail.NewMockBackend()

	t.Run("focus cycling", func(t *testing.T) {
		tab := NewAccountTab(styles, backend)
		tab.width = 80
		tab.height = 20
		tab, _ = tab.Update(tea.KeyMsg{Type: tea.KeyTab})
		if tab.focused != MsgListPanel {
			t.Errorf("focused = %d, want MsgListPanel", tab.focused)
		}
		tab, _ = tab.Update(tea.KeyMsg{Type: tea.KeyTab})
		if tab.focused != SidebarPanel {
			t.Errorf("focused = %d, want SidebarPanel", tab.focused)
		}
	})

	t.Run("window size", func(t *testing.T) {
		tab := NewAccountTab(styles, backend)
		tab, _ = tab.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		if tab.width != 120 || tab.height != 40 {
			t.Errorf("size = %dx%d, want 120x40", tab.width, tab.height)
		}
	})
}
```

- [ ] **Step 6: Verify tests pass**

Run: `go test ./internal/ui/ -v`
Expected: PASS (some tests in app_test.go may fail — that's expected, fixed in Task 6)

- [ ] **Step 7: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git add -u internal/ui/tab_bar.go internal/ui/tab_bar_test.go internal/ui/tab.go
git commit -m "Remove tab bar and Tab interface, simplify AccountTab"
```

---

### Task 3: Create TopLine Component

**Files:**
- Create: `internal/ui/top_line.go`
- Create: `internal/ui/top_line_test.go`
- Modify: `internal/ui/styles.go`

- [ ] **Step 1: Add TopLine styles**

In `internal/ui/styles.go`, add to the `Styles` struct and `NewStyles`:

```go
// In Styles struct, add:
TopLine     lipgloss.Style
ToastText   lipgloss.Style

// In NewStyles, add:
TopLine: lipgloss.NewStyle().
    Foreground(t.BgBorder),
ToastText: lipgloss.NewStyle().
    Foreground(t.ColorSuccess),
```

- [ ] **Step 2: Write failing test for TopLine**

Create `internal/ui/top_line_test.go`:

```go
package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func TestTopLineView(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("basic frame line", func(t *testing.T) {
		tl := NewTopLine(styles)
		result := stripANSI(tl.View(80, 30))
		if !strings.HasSuffix(strings.TrimRight(result, " "), "╮") {
			t.Errorf("top line missing ╮ at right edge: %q", result)
		}
		if !strings.Contains(result, "┬") {
			t.Errorf("top line missing ┬ divider junction: %q", result)
		}
	})

	t.Run("divider at correct position", func(t *testing.T) {
		tl := NewTopLine(styles)
		result := stripANSI(tl.View(80, 30))
		runes := []rune(result)
		if runes[30] != '┬' {
			t.Errorf("expected ┬ at position 30, got %c", runes[30])
		}
	})

	t.Run("width matches terminal", func(t *testing.T) {
		tl := NewTopLine(styles)
		result := tl.View(80, 30)
		if lipgloss.Width(result) != 80 {
			t.Errorf("width = %d, want 80", lipgloss.Width(result))
		}
	})

	t.Run("no divider when dividerCol is 0", func(t *testing.T) {
		tl := NewTopLine(styles)
		result := stripANSI(tl.View(80, 0))
		if strings.Contains(result, "┬") {
			t.Errorf("should not have ┬ when dividerCol is 0: %q", result)
		}
	})

	t.Run("toast overlays right side", func(t *testing.T) {
		tl := NewTopLine(styles)
		tl.SetToast("✓ 3 archived")
		result := stripANSI(tl.View(80, 30))
		if !strings.Contains(result, "✓ 3 archived") {
			t.Errorf("toast not visible: %q", result)
		}
		if !strings.HasSuffix(strings.TrimRight(result, " "), "╮") {
			t.Errorf("╮ missing after toast: %q", result)
		}
	})

	t.Run("toast clears", func(t *testing.T) {
		tl := NewTopLine(styles)
		tl.SetToast("✓ done")
		tl.ClearToast()
		result := stripANSI(tl.View(80, 30))
		if strings.Contains(result, "done") {
			t.Error("toast should be cleared")
		}
	})
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/ui/ -run TestTopLine -v`
Expected: FAIL — `NewTopLine` not defined

- [ ] **Step 4: Implement TopLine**

Create `internal/ui/top_line.go`:

```go
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TopLine renders the top frame edge: ──┬──╮ with optional toast overlay.
type TopLine struct {
	styles Styles
	toast  string
}

// NewTopLine creates a TopLine with the given styles.
func NewTopLine(styles Styles) TopLine {
	return TopLine{styles: styles}
}

// SetToast sets a toast message to overlay on the right side.
func (tl *TopLine) SetToast(msg string) {
	tl.toast = msg
}

// ClearToast removes the toast message.
func (tl *TopLine) ClearToast() {
	tl.toast = ""
}

// View renders the top line at the given width. dividerCol is the
// column position of the panel divider (0 to skip the junction).
func (tl TopLine) View(width, dividerCol int) string {
	style := tl.styles.TopLine

	// Build the right portion: " toast ─╮" or just "─╮"
	rightEnd := "─╮"
	var toastPart string
	if tl.toast != "" {
		toastPart = " " + tl.toast + " "
	}
	rightEndWidth := lipgloss.Width(rightEnd)
	toastWidth := lipgloss.Width(toastPart)

	// Fill the line with ─, placing ┬ at dividerCol
	fillWidth := width - rightEndWidth - toastWidth
	if fillWidth < 1 {
		fillWidth = 1
	}

	var buf strings.Builder
	for i := 0; i < fillWidth; i++ {
		if dividerCol > 0 && i == dividerCol {
			buf.WriteRune('┬')
		} else {
			buf.WriteRune('─')
		}
	}

	line := buf.String()
	if tl.toast != "" {
		return style.Render(line) +
			tl.styles.ToastText.Render(toastPart) +
			style.Render(rightEnd)
	}
	return style.Render(line + rightEnd)
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/ui/ -run TestTopLine -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ui/top_line.go internal/ui/top_line_test.go internal/ui/styles.go
git commit -m "Add TopLine component: top frame edge with toast overlay"
```

---

### Task 4: Redesign StatusBar as Bottom Frame Edge

**Files:**
- Modify: `internal/ui/status_bar.go`
- Modify: `internal/ui/status_bar_test.go`
- Modify: `internal/ui/styles.go`

- [ ] **Step 1: Write failing tests for new StatusBar**

Replace `internal/ui/status_bar_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestStatusBarView(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("renders counts and connection", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(true)
		result := stripANSI(sb.View(80, 30))
		if !strings.Contains(result, "10 messages") {
			t.Error("missing message count")
		}
		if !strings.Contains(result, "3 unread") {
			t.Error("missing unread count")
		}
		if !strings.Contains(result, "● connected") {
			t.Error("missing connection indicator")
		}
	})

	t.Run("no folder name", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(true)
		result := stripANSI(sb.View(80, 30))
		if strings.Contains(result, "Inbox") {
			t.Error("status bar should not show folder name")
		}
	})

	t.Run("ends with ─╯", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(true)
		result := stripANSI(sb.View(80, 30))
		trimmed := strings.TrimRight(result, " ")
		if !strings.HasSuffix(trimmed, "─╯") {
			t.Errorf("status bar should end with ─╯: %q", trimmed)
		}
	})

	t.Run("has ┴ at divider position", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(true)
		result := stripANSI(sb.View(80, 30))
		runes := []rune(result)
		if len(runes) > 30 && runes[30] != '┴' {
			t.Errorf("expected ┴ at position 30, got %c", runes[30])
		}
	})

	t.Run("offline state", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(false)
		result := stripANSI(sb.View(80, 30))
		if !strings.Contains(result, "○ offline") {
			t.Error("missing offline indicator")
		}
	})

	t.Run("no divider when dividerCol is 0", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(true)
		result := stripANSI(sb.View(80, 0))
		if strings.Contains(result, "┴") {
			t.Error("should not have ┴ when dividerCol is 0")
		}
	})

	t.Run("width matches terminal", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetCounts(10, 3)
		sb.SetConnected(true)
		result := sb.View(80, 30)
		if lipgloss.Width(result) != 80 {
			t.Errorf("width = %d, want 80", lipgloss.Width(result))
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -run TestStatusBar -v`
Expected: FAIL — `SetCounts` not defined, wrong `View` signature

- [ ] **Step 3: Rewrite StatusBar**

Replace `internal/ui/status_bar.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ConnectionState represents the mail connection status.
type ConnectionState int

const (
	Connected ConnectionState = iota
	Offline
	Reconnecting
)

// StatusBar renders the bottom frame edge with combined status indicator.
type StatusBar struct {
	styles    Styles
	total     int
	unread    int
	connState ConnectionState
}

// NewStatusBar creates a StatusBar with the given styles.
func NewStatusBar(styles Styles) StatusBar {
	return StatusBar{
		styles:    styles,
		connState: Connected,
	}
}

// SetCounts updates the message and unread counts.
func (sb *StatusBar) SetCounts(total, unread int) {
	sb.total = total
	sb.unread = unread
}

// SetConnected sets the connection state to connected or offline.
func (sb *StatusBar) SetConnected(connected bool) {
	if connected {
		sb.connState = Connected
	} else {
		sb.connState = Offline
	}
}

// SetConnectionState sets the connection state directly.
func (sb *StatusBar) SetConnectionState(state ConnectionState) {
	sb.connState = state
}

// View renders the status bar at the given width. dividerCol is the
// column position of the panel divider (0 to skip the junction).
func (sb StatusBar) View(width, dividerCol int) string {
	// Build right portion: " 10 messages · 3 unread · ● connected ─╯"
	counts := fmt.Sprintf("%d messages", sb.total)
	if sb.unread > 0 {
		counts += fmt.Sprintf(" · %d unread", sb.unread)
	}

	var connIcon, connText string
	var connStyle lipgloss.Style
	switch sb.connState {
	case Connected:
		connIcon = "●"
		connText = "connected"
		connStyle = sb.styles.StatusConnected
	case Offline:
		connIcon = "○"
		connText = "offline"
		connStyle = sb.styles.StatusOffline
	case Reconnecting:
		connIcon = "◐"
		connText = "reconnecting"
		connStyle = sb.styles.StatusReconnect
	}

	// Plain text for width calculation
	rightText := " " + counts + " · " + connIcon + " " + connText + " ─╯"
	rightWidth := lipgloss.Width(rightText)

	// Build left fill with ┴ at dividerCol
	fillWidth := maxInt(0, width-rightWidth)
	var buf strings.Builder
	for i := 0; i < fillWidth; i++ {
		if dividerCol > 0 && i == dividerCol {
			buf.WriteRune('┴')
		} else {
			buf.WriteRune('─')
		}
	}

	// Render with styles
	framePart := sb.styles.StatusBar.Render(buf.String())
	countsPart := sb.styles.StatusBar.Render(" " + counts + " · ")
	connPart := connStyle.Render(connIcon)
	textPart := sb.styles.StatusBar.Render(" " + connText + " ")
	endPart := sb.styles.TopLine.Render("─╯")

	return framePart + countsPart + connPart + textPart + endPart
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -run TestStatusBar -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ui/status_bar.go internal/ui/status_bar_test.go
git commit -m "Redesign StatusBar as bottom frame edge with combined indicator"
```

---

### Task 5: Update Footer with Grouped Keybindings

**Files:**
- Modify: `internal/ui/keys.go`
- Modify: `internal/ui/footer.go`
- Modify: `internal/ui/footer_test.go`

- [ ] **Step 1: Write failing tests for grouped footer**

Replace `internal/ui/footer_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestFooterView(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("message list context has group separator", func(t *testing.T) {
		f := NewFooter(styles)
		f.SetContext(MsgListContext)
		result := stripANSI(f.View(120))
		if !strings.Contains(result, "┊") {
			t.Error("missing group separator ┊")
		}
	})

	t.Run("message list has triage group", func(t *testing.T) {
		f := NewFooter(styles)
		f.SetContext(MsgListContext)
		result := stripANSI(f.View(120))
		if !strings.Contains(result, "d:del") {
			t.Error("missing d:del")
		}
		if !strings.Contains(result, "a:archive") {
			t.Error("missing a:archive")
		}
	})

	t.Run("message list has reply group", func(t *testing.T) {
		f := NewFooter(styles)
		f.SetContext(MsgListContext)
		result := stripANSI(f.View(120))
		if !strings.Contains(result, "r:reply") {
			t.Error("missing r:reply")
		}
		if !strings.Contains(result, "c:compose") {
			t.Error("missing c:compose")
		}
	})

	t.Run("sidebar context has folder jumps", func(t *testing.T) {
		f := NewFooter(styles)
		f.SetContext(SidebarContext)
		result := stripANSI(f.View(120))
		if !strings.Contains(result, "I:inbox") {
			t.Error("missing I:inbox folder jump")
		}
		if !strings.Contains(result, "D:drafts") {
			t.Error("missing D:drafts folder jump")
		}
	})

	t.Run("starts with 1-space padding", func(t *testing.T) {
		f := NewFooter(styles)
		f.SetContext(MsgListContext)
		result := stripANSI(f.View(120))
		if !strings.HasPrefix(result, " ") {
			t.Error("footer should start with 1-space padding")
		}
	})

	t.Run("sidebar does not show delete", func(t *testing.T) {
		f := NewFooter(styles)
		f.SetContext(SidebarContext)
		result := stripANSI(f.View(120))
		if strings.Contains(result, "d:del") {
			t.Error("sidebar should not show delete hint")
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -run TestFooter -v`
Expected: FAIL — grouped separator not present

- [ ] **Step 3: Rewrite keys.go with logical groups**

Replace `internal/ui/keys.go`:

```go
package ui

import "github.com/charmbracelet/bubbles/key"

// keyGroup is a slice of bindings that belong together visually.
type keyGroup []key.Binding

// GlobalKeys are handled by the root App model.
type GlobalKeys struct {
	Help key.Binding
	Cmd  key.Binding
	Quit key.Binding
}

// NewGlobalKeys returns the default global key bindings.
func NewGlobalKeys() GlobalKeys {
	return GlobalKeys{
		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Cmd:  key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// FolderJumpKeys jump to canonical folders.
type FolderJumpKeys struct {
	Inbox   key.Binding
	Drafts  key.Binding
	Sent    key.Binding
	Archive key.Binding
	Spam    key.Binding
	Trash   key.Binding
}

// NewFolderJumpKeys returns the default folder jump bindings.
func NewFolderJumpKeys() FolderJumpKeys {
	return FolderJumpKeys{
		Inbox:   key.NewBinding(key.WithKeys("I"), key.WithHelp("I", "inbox")),
		Drafts:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "drafts")),
		Sent:    key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sent")),
		Archive: key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "archive")),
		Spam:    key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "spam")),
		Trash:   key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "trash")),
	}
}

// MsgListKeys groups for the message list footer.
type MsgListKeys struct {
	triage  keyGroup
	reply   keyGroup
	app     keyGroup
}

// NewMsgListKeys returns grouped message list bindings.
func NewMsgListKeys() MsgListKeys {
	return MsgListKeys{
		triage: keyGroup{
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "del")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
			key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "star")),
		},
		reply: keyGroup{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
			key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "all")),
			key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fwd")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compose")),
		},
		app: keyGroup{
			key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
			key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
			key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
		},
	}
}

// Groups returns the key groups for footer rendering.
func (k MsgListKeys) Groups() []keyGroup {
	return []keyGroup{k.triage, k.reply, k.app}
}

// SidebarKeys groups for the sidebar footer.
type SidebarKeys struct {
	action keyGroup
	folder keyGroup
	app    keyGroup
}

// NewSidebarKeys returns grouped sidebar bindings.
func NewSidebarKeys() SidebarKeys {
	fj := NewFolderJumpKeys()
	return SidebarKeys{
		action: keyGroup{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
			key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compose")),
		},
		folder: keyGroup{fj.Inbox, fj.Drafts, fj.Sent, fj.Archive},
		app: keyGroup{
			key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
			key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
		},
	}
}

// Groups returns the key groups for footer rendering.
func (k SidebarKeys) Groups() []keyGroup {
	return []keyGroup{k.action, k.folder, k.app}
}

// ViewerKeys groups for the viewer footer.
type ViewerKeys struct {
	triage keyGroup
	reply  keyGroup
	viewer keyGroup
}

// NewViewerKeys returns grouped viewer bindings.
func NewViewerKeys() ViewerKeys {
	return ViewerKeys{
		triage: keyGroup{
			key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "del")),
			key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
			key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "star")),
		},
		reply: keyGroup{
			key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
			key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "all")),
			key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fwd")),
		},
		viewer: keyGroup{
			key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "links")),
			key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "close")),
			key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		},
	}
}

// Groups returns the key groups for footer rendering.
func (k ViewerKeys) Groups() []keyGroup {
	return []keyGroup{k.triage, k.reply, k.viewer}
}
```

- [ ] **Step 4: Rewrite footer.go with group rendering**

Replace `internal/ui/footer.go`:

```go
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FooterContext identifies which keybinding set to display.
type FooterContext int

const (
	MsgListContext FooterContext = iota
	SidebarContext
	ViewerContext
)

// Footer renders context-appropriate keybinding hints with group separators.
type Footer struct {
	styles      Styles
	context     FooterContext
	msgKeys     MsgListKeys
	sidebarKeys SidebarKeys
	viewerKeys  ViewerKeys
}

// NewFooter creates a Footer with the given styles.
func NewFooter(styles Styles) Footer {
	return Footer{
		styles:      styles,
		context:     MsgListContext,
		msgKeys:     NewMsgListKeys(),
		sidebarKeys: NewSidebarKeys(),
		viewerKeys:  NewViewerKeys(),
	}
}

// SetContext switches the displayed keybinding set.
func (f *Footer) SetContext(ctx FooterContext) {
	f.context = ctx
}

// View renders the footer at the given width.
func (f Footer) View(width int) string {
	var groups []keyGroup
	switch f.context {
	case SidebarContext:
		groups = f.sidebarKeys.Groups()
	case ViewerContext:
		groups = f.viewerKeys.Groups()
	default:
		groups = f.msgKeys.Groups()
	}

	sep := " " + f.styles.FooterSep.Render("┊") + "  "

	var parts []string
	for _, g := range groups {
		var bindings []string
		for _, b := range g {
			k := f.styles.FooterKey.Render(b.Help().Key)
			d := f.styles.FooterHint.Render(":" + b.Help().Desc)
			bindings = append(bindings, k+d)
		}
		parts = append(parts, strings.Join(bindings, "  "))
	}

	line := " " + strings.Join(parts, sep)
	return lipgloss.NewStyle().Width(width).Render(line)
}
```

- [ ] **Step 5: Add FooterSep style**

In `internal/ui/styles.go`, add to `Styles` struct and `NewStyles`:

```go
// In Styles struct, add:
FooterSep lipgloss.Style

// In NewStyles, add:
FooterSep: lipgloss.NewStyle().
    Foreground(t.FgDim),
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./internal/ui/ -run TestFooter -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/ui/keys.go internal/ui/footer.go internal/ui/footer_test.go internal/ui/styles.go
git commit -m "Rewrite footer with grouped keybindings and ┊ separator"
```

---

### Task 6: Rewire App to New Layout

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/app_test.go`
- Modify: `cmd/poplar/root.go`

- [ ] **Step 1: Write failing tests for new App layout**

Replace `internal/ui/app_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestApp(t *testing.T) {
	backend := mail.NewMockBackend()

	t.Run("initial state", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		if app.width != 0 {
			t.Errorf("initial width = %d, want 0", app.width)
		}
	})

	t.Run("quit on q", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app.width = 80
		app.height = 24
		_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Fatal("expected quit command")
		}
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Errorf("expected QuitMsg, got %T", msg)
		}
	})

	t.Run("window size stored", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		if app.width != 120 || app.height != 40 {
			t.Errorf("size = %dx%d, want 120x40", app.width, app.height)
		}
	})

	t.Run("view has top line with ╮", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := app.View()
		plain := stripANSI(view)
		lines := strings.Split(plain, "\n")
		if len(lines) < 1 {
			t.Fatal("no lines rendered")
		}
		trimmed := strings.TrimRight(lines[0], " ")
		if !strings.HasSuffix(trimmed, "╮") {
			t.Errorf("first line should end with ╮: %q", trimmed)
		}
	})

	t.Run("view has status bar with ╯", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := app.View()
		plain := stripANSI(view)
		lines := strings.Split(plain, "\n")
		// Status bar is second-to-last line
		statusLine := strings.TrimRight(lines[len(lines)-2], " ")
		if !strings.HasSuffix(statusLine, "╯") {
			t.Errorf("status bar should end with ╯: %q", statusLine)
		}
	})

	t.Run("view has right border │", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := app.View()
		plain := stripANSI(view)
		lines := strings.Split(plain, "\n")
		// Content lines (between top line and status bar) should have │ at right
		for i := 1; i < len(lines)-2; i++ {
			runes := []rune(lines[i])
			if len(runes) == 0 {
				continue
			}
			last := runes[len(runes)-1]
			if last != '│' {
				t.Errorf("line %d: last char = %c, want │: %q", i+1, last, lines[i])
			}
		}
	})

	t.Run("no tab bar", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := app.View()
		plain := stripANSI(view)
		if strings.Contains(plain, "╭") {
			t.Error("should not contain ╭ (tab bar removed)")
		}
	})

	t.Run("content height is height minus 3 chrome rows", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app.width = 80
		app.height = 24
		if app.contentHeight() != 21 {
			t.Errorf("contentHeight = %d, want 21", app.contentHeight())
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -run TestApp -v`
Expected: FAIL — old App structure

- [ ] **Step 3: Rewrite App**

Replace `internal/ui/app.go`:

```go
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

// App is the root bubbletea model for poplar.
type App struct {
	account   AccountTab
	styles    Styles
	topLine   TopLine
	statusBar StatusBar
	footer    Footer
	keys      GlobalKeys
	width     int
	height    int
}

// NewApp creates the root model with a single account view.
func NewApp(t *theme.CompiledTheme, backend mail.Backend) App {
	styles := NewStyles(t)
	acct := NewAccountTab(styles, backend)

	sb := NewStatusBar(styles)
	folders, _ := backend.ListFolders()
	if len(folders) > 0 {
		inbox := folders[0]
		sb.SetCounts(inbox.Exists, inbox.Unseen)
	}
	sb.SetConnected(true)

	return App{
		account:   acct,
		styles:    styles,
		topLine:   NewTopLine(styles),
		statusBar: sb,
		footer:    NewFooter(styles),
		keys:      NewGlobalKeys(),
	}
}

// Init returns no initial command.
func (m App) Init() tea.Cmd { return nil }

// Update handles global keys and delegates to the account tab.
func (m App) Update(msg tea.Msg) (App, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		contentHeight := m.contentHeight()
		tabMsg := tea.WindowSizeMsg{Width: m.width - 1, Height: contentHeight}
		var cmd tea.Cmd
		m.account, cmd = m.account.Update(tabMsg)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			// Stubbed for help popover
			return m, nil
		case ":":
			// Stubbed for command mode
			return m, nil
		}
	}

	// Delegate to account tab
	var cmd tea.Cmd
	m.account, cmd = m.account.Update(msg)
	m.updateFooterContext()

	return m, cmd
}

// View composes the full-screen layout.
func (m App) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	dividerCol := sidebarWidth
	topLine := m.topLine.View(m.width, dividerCol)

	// Render content with right border
	rawContent := m.account.View()
	rightBorder := m.styles.FrameBorder.Render("│")
	contentLines := strings.Split(rawContent, "\n")
	for i, line := range contentLines {
		pad := maxInt(0, m.width-1-lipgloss.Width(line))
		contentLines[i] = line + strings.Repeat(" ", pad) + rightBorder
	}
	content := strings.Join(contentLines, "\n")

	status := m.statusBar.View(m.width, dividerCol)
	foot := m.footer.View(m.width)

	return lipgloss.JoinVertical(lipgloss.Left,
		topLine,
		content,
		status,
		foot,
	)
}

// contentHeight returns the height available for the content area.
// Chrome: top line (1) + status bar (1) + footer (1) = 3 rows.
func (m App) contentHeight() int {
	chrome := 3
	h := m.height - chrome
	if h < 1 {
		return 1
	}
	return h
}

// updateFooterContext switches the footer based on the account tab's focus.
func (m *App) updateFooterContext() {
	if m.account.focused == SidebarPanel {
		m.footer.SetContext(SidebarContext)
	} else {
		m.footer.SetContext(MsgListContext)
	}
}
```

- [ ] **Step 4: Simplify root.go**

Replace `cmd/poplar/root.go`:

```go
package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/theme"
	"github.com/glw907/beautiful-aerc/internal/ui"
	"github.com/spf13/cobra"
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

// appModel wraps ui.App to satisfy tea.Model (returns tea.Model, not App).
type appModel struct {
	app ui.App
}

func (m appModel) Init() tea.Cmd { return m.app.Init() }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	app, cmd := m.app.Update(msg)
	m.app = app
	return m, cmd
}

func (m appModel) View() string { return m.app.View() }

func runRoot(_ *cobra.Command, _ []string) error {
	backend := mail.NewMockBackend()
	app := ui.NewApp(theme.Nord, backend)

	p := tea.NewProgram(appModel{app: app}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running poplar: %w", err)
	}
	return nil
}
```

- [ ] **Step 5: Run all tests**

Run: `go test ./internal/ui/ -v`
Expected: PASS

- [ ] **Step 6: Build and launch to verify visually**

```bash
go build -o /tmp/poplar ./cmd/poplar/ && kitty --title "Poplar" -e /tmp/poplar &
```

Verify: top line with `──┬──╮`, right `│` border on content rows, status bar ending with `─╯`, footer below with `┊` separators.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go cmd/poplar/root.go
git commit -m "Rewire App: drop tabs, TopLine + StatusBar frame, grouped footer"
```

---

### Task 7: Add Account Name to Sidebar

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Write failing test**

Add to `internal/ui/account_tab_test.go`:

```go
t.Run("view shows account name", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	tab.width = 80
	tab.height = 20
	tab, _ = tab.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	view := stripANSI(tab.View())
	if !strings.Contains(view, "geoff@907.life") {
		t.Error("sidebar should show account name")
	}
})
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/ui/ -run TestAccountTab/view_shows_account -v`
Expected: FAIL

- [ ] **Step 3: Update AccountTab to render account name**

In `internal/ui/account_tab.go`, update `View()` to prepend the account name and blank line to the sidebar panel:

```go
func (m AccountTab) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	sw := minInt(sidebarWidth, m.width/2)
	mw := m.width - sw - 1 // -1 for divider

	// Sidebar: account name + blank line + placeholder
	acctName := m.styles.Dim.Render(
		lipgloss.NewStyle().Width(sw).Render(" " + m.backend.AccountName()),
	)
	blankLine := strings.Repeat(" ", sw)

	sidebarBody := renderPlaceholder("Sidebar", sw, m.height-2, m.focused == SidebarPanel, m.styles)
	sidebar := acctName + "\n" + blankLine + "\n" + sidebarBody

	divider := renderDivider(m.height, m.styles)
	msglist := renderPlaceholder("Message List", mw, m.height, m.focused == MsgListPanel, m.styles)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, divider, msglist)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/ui/ -run TestAccountTab -v`
Expected: PASS

- [ ] **Step 5: Build and launch to verify**

```bash
go build -o /tmp/poplar ./cmd/poplar/ && kitty --title "Poplar" -e /tmp/poplar &
```

Verify: account name appears at top of sidebar in dim text, blank line below it.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "Show account name at top of sidebar"
```

---

### Task 8: Visual Verification and Polish

**Files:**
- All UI files (read-only inspection)

- [ ] **Step 1: Build and launch at 120x40**

```bash
go build -o /tmp/poplar ./cmd/poplar/
tmux kill-session -t poplar 2>/dev/null
tmux new-session -d -s poplar -x 120 -y 40
tmux send-keys -t poplar '/tmp/poplar' Enter
sleep 1
tmux capture-pane -t poplar -p
```

Verify against spec wireframe:
- Top line: `──┬──╮` spanning full width, `┬` at column 30
- Content: `│` right border on every row, `│` divider at column 30
- Status bar: `──┴── counts · ● connected ─╯`
- Footer: grouped with `┊` separators, 1-space left padding
- Account name at top of sidebar in dim text

- [ ] **Step 2: Check junction alignment**

The `┬` on the top line, the `│` divider on content rows, and the `┴` on the status bar must all be at the same column (30). Inspect the captured output.

- [ ] **Step 3: Check width consistency**

Every line should be exactly 120 characters wide (or the terminal width). Check with:

```bash
tmux capture-pane -t poplar -p | sed 's/\x1b\[[0-9;]*m//g' | while IFS= read -r line; do echo "$(echo -n "$line" | wc -m)"; done | sort -u
```

- [ ] **Step 4: Run make check**

```bash
make check
```

Expected: vet + test pass

- [ ] **Step 5: Fix any issues found, commit**

If alignment or width issues are found, fix them and re-verify. Commit the fixes.

- [ ] **Step 6: make install**

```bash
make install
```

---

### Task 9: Update Documentation

**Files:**
- Modify: `docs/poplar/STATUS.md`
- Modify: `docs/poplar/architecture.md`
- Modify: `docs/poplar/wireframes.md`

- [ ] **Step 1: Update architecture.md**

Add these decisions to the Key Decisions section:

```markdown
### Drop tabs in favor of sidebar
**Decision:** Remove the tab bar entirely. The sidebar (always
visible) shows folder context. Opening a message renders in the
right panel, not a new tab. `q` returns to the message list.
**Rationale:** With the sidebar always visible, the tab bar
provided no new information while consuming 3 rows. Simplifies
navigation — no tab lifecycle, no `1-9` switching. Aligns with
"Better Pine" philosophy (one thing at a time).
**Date:** 2026-04-11

### Three-sided frame with open left edge
**Decision:** Top `──┬──╮`, right `│`, bottom status bar
`──┴──╯`. No left border.
**Rationale:** Distinctive asymmetric frame that avoids the
junction problem at bottom-left where the status bar meets a
left border. The open left edge matches the bottom where the
grey status bar starts at column 0.
**Date:** 2026-04-11

### Account name in sidebar, switchable
**Decision:** Account name at top of sidebar, one account at a
time, key to cycle between accounts.
**Rationale:** Pine-style simplicity over stacked account trees.
When the sidebar collapses, the top frame line shows
`account · folder` for context.
**Date:** 2026-04-11

### Colorblind-accessible connection states
**Decision:** Connection states use shape + color + text: `●`
green filled (connected), `○` red hollow (offline), `◐` orange
half (reconnecting).
**Rationale:** Triple redundancy ensures accessibility across
colorblind conditions, monochrome terminals, and screen readers.
**Date:** 2026-04-11

### Footer group separators
**Decision:** `┊` (U+250A, light quadruple dash vertical) in
`fg_dim` between key groups. Custom rendering, not bubbles/help.
**Rationale:** Subtle enough to recede behind key hints, clear
enough to read groups. Spacing alone was insufficient.
**Date:** 2026-04-11
```

- [ ] **Step 2: Update STATUS.md**

Mark the chrome redesign as done. Update the pass table and starter prompt for the next pass.

- [ ] **Step 3: Update wireframes.md**

Replace wireframe sections 1, 2, 5, 7, 8 with the updated wireframes from the spec. Remove tab bar wireframes entirely.

- [ ] **Step 4: Commit**

```bash
git add docs/poplar/STATUS.md docs/poplar/architecture.md docs/poplar/wireframes.md
git commit -m "Update docs: architecture decisions, wireframes, status"
```

- [ ] **Step 5: Push**

```bash
git push
```

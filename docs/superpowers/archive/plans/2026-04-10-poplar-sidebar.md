# Poplar Sidebar Implementation Plan (Pass 2.5b-2)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the sidebar placeholder in AccountTab with a real folder list component — three folder groups, selection with `┃` indicator, unread badges, and j/k/G/gg navigation.

**Architecture:** New `Sidebar` model in `internal/ui/sidebar.go` renders folder rows using lipgloss. Folder data comes from `mail.Backend.ListFolders()`, grouped by role into Primary/Disposal/Custom with blank-line separators. The sidebar is a child of `AccountTab`, which delegates key events when `SidebarPanel` is focused. Folder jumps deferred to command mode (Pass 2.5b-7).

**Tech Stack:** Go, bubbletea, lipgloss, bubbles/key

**Spec:** Wireframes at `docs/poplar/wireframes.md` section 3, chrome shell spec at `docs/superpowers/specs/2026-04-10-poplar-chrome-shell-design.md`.

---

### File Structure

```
internal/ui/sidebar.go       — Sidebar model (folder groups, selection, rendering)
internal/ui/sidebar_test.go   — Sidebar unit tests (render, navigation, groups)
internal/ui/account_tab.go    — Modified: replace placeholder with Sidebar child
internal/ui/account_tab_test.go — Modified: update tests for sidebar integration
internal/ui/styles.go         — Modified: add sidebar-specific styles
internal/ui/styles_test.go    — Modified: add sidebar style tests
```

---

### Task 1: Add Sidebar Styles

**Files:**
- Modify: `internal/ui/styles.go`
- Modify: `internal/ui/styles_test.go`

- [ ] **Step 1: Add sidebar style fields to Styles struct**

Add these fields to the `Styles` struct in `styles.go`:

```go
// Sidebar
SidebarFolder       lipgloss.Style // folder name (fg_base)
SidebarFolderUnread lipgloss.Style // folder name when has unread (accent_tertiary)
SidebarIconUnread   lipgloss.Style // folder icon when has unread (accent_tertiary)
SidebarCount        lipgloss.Style // unread count badge (accent_tertiary)
SidebarSelected     lipgloss.Style // selected row bg (bg_selection)
SidebarIndicator    lipgloss.Style // ┃ focused indicator (accent_primary)
```

And populate them in `NewStyles`:

```go
SidebarFolder: lipgloss.NewStyle().
    Foreground(t.FgBase),
SidebarFolderUnread: lipgloss.NewStyle().
    Foreground(t.AccentTertiary),
SidebarIconUnread: lipgloss.NewStyle().
    Foreground(t.AccentTertiary),
SidebarCount: lipgloss.NewStyle().
    Foreground(t.AccentTertiary),
SidebarSelected: lipgloss.NewStyle().
    Background(t.BgSelection),
SidebarIndicator: lipgloss.NewStyle().
    Foreground(t.AccentPrimary),
```

- [ ] **Step 2: Add sidebar style tests**

Add test cases to `TestNewStyles` in `styles_test.go`:

```go
{"SidebarFolder", s.SidebarFolder.GetForeground(), t.FgBase},
{"SidebarFolderUnread", s.SidebarFolderUnread.GetForeground(), t.AccentTertiary},
{"SidebarIconUnread", s.SidebarIconUnread.GetForeground(), t.AccentTertiary},
{"SidebarCount", s.SidebarCount.GetForeground(), t.AccentTertiary},
{"SidebarIndicator", s.SidebarIndicator.GetForeground(), t.AccentPrimary},
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/ui/ -run TestNewStyles -v
```

Expected: PASS with all new style tests.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/styles.go internal/ui/styles_test.go
git commit -m "Add sidebar styles to compiled theme bridge"
```

---

### Task 2: Create Sidebar Model

**Files:**
- Create: `internal/ui/sidebar.go`
- Create: `internal/ui/sidebar_test.go`

- [ ] **Step 1: Write failing tests for sidebar rendering**

Create `internal/ui/sidebar_test.go`:

```go
package ui

import (
	"regexp"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestSidebar(t *testing.T) {
	styles := NewStyles(theme.Nord)
	folders := mockFolders()

	t.Run("renders all folders", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		view := sb.View()
		plain := stripANSI(view)
		for _, f := range folders {
			if !strings.Contains(plain, f.Name) {
				t.Errorf("missing folder %q in view", f.Name)
			}
		}
	})

	t.Run("groups separated by blank lines", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		view := sb.View()
		plain := stripANSI(view)
		lines := strings.Split(plain, "\n")

		// Find blank lines (only whitespace)
		var blankIdxs []int
		for i, line := range lines {
			if strings.TrimSpace(line) == "" {
				blankIdxs = append(blankIdxs, i)
			}
		}
		if len(blankIdxs) < 2 {
			t.Errorf("expected at least 2 blank separator lines, got %d", len(blankIdxs))
		}
	})

	t.Run("initial selection is first folder", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		if sb.Selected() != 0 {
			t.Errorf("initial selection = %d, want 0", sb.Selected())
		}
		if sb.SelectedFolder() != "Inbox" {
			t.Errorf("selected folder = %q, want Inbox", sb.SelectedFolder())
		}
	})

	t.Run("unread count shown only when positive", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		view := sb.View()
		plain := stripANSI(view)
		lines := strings.Split(plain, "\n")

		// Inbox has 3 unread — should show count
		inboxLine := findLineContaining(lines, "Inbox")
		if inboxLine == "" {
			t.Fatal("no line containing Inbox")
		}
		if !strings.Contains(inboxLine, "3") {
			t.Error("Inbox line missing unread count 3")
		}

		// Sent has 0 unread — should NOT show count
		sentLine := findLineContaining(lines, "Sent")
		if sentLine == "" {
			t.Fatal("no line containing Sent")
		}
		// "0" should not appear as a count
		// Check that no digit follows Sent (only whitespace)
		re := regexp.MustCompile(`Sent\s+\d`)
		if re.MatchString(sentLine) {
			t.Errorf("Sent line shows count when unseen=0: %q", sentLine)
		}
	})

	t.Run("selected row has focus indicator when focused", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		sb.focused = true
		view := sb.View()
		plain := stripANSI(view)
		lines := strings.Split(plain, "\n")
		if len(lines) == 0 {
			t.Fatal("empty view")
		}
		// First non-blank line should have ┃
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				if !strings.Contains(line, "┃") {
					t.Errorf("focused selected row missing ┃: %q", line)
				}
				break
			}
		}
	})

	t.Run("no focus indicator when unfocused", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		sb.focused = false
		view := sb.View()
		plain := stripANSI(view)
		if strings.Contains(plain, "┃") {
			t.Error("unfocused sidebar should not have ┃ indicator")
		}
	})

	t.Run("all lines same display width", func(t *testing.T) {
		sb := NewSidebar(styles, folders, 30, 20)
		view := sb.View()
		lines := strings.Split(view, "\n")
		for i, line := range lines {
			w := lipgloss.Width(line)
			if w != 30 {
				t.Errorf("line %d width = %d, want 30: %q",
					i, w, stripANSI(line))
			}
		}
	})
}

func mockFolders() []mail.Folder {
	return []mail.Folder{
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
	}
}

func findLineContaining(lines []string, substr string) string {
	for _, line := range lines {
		if strings.Contains(line, substr) {
			return line
		}
	}
	return ""
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/ -run TestSidebar -v
```

Expected: FAIL — `NewSidebar` not defined.

- [ ] **Step 3: Write Sidebar model**

Create `internal/ui/sidebar.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/mail"
)

// folderGroup classifies folders for visual grouping.
type folderGroup int

const (
	primaryGroup  folderGroup = iota // inbox, drafts, sent, archive
	disposalGroup                    // spam, trash
	customGroup                      // everything else
)

// folderEntry holds a folder and its display metadata.
type folderEntry struct {
	folder mail.Folder
	icon   string
	group  folderGroup
}

// Sidebar renders the folder list with groups, selection, and unread badges.
type Sidebar struct {
	entries  []folderEntry
	selected int
	focused  bool
	styles   Styles
	width    int
	height   int
}

// folderIcon returns the Nerd Font icon for a folder based on its name and role.
func folderIcon(f mail.Folder) string {
	switch strings.ToLower(f.Role) {
	case "inbox":
		return "󰇰"
	case "drafts":
		return "󰏫"
	case "sent":
		return "󰑚"
	case "archive":
		return "󰀼"
	case "junk":
		return "󰍷"
	case "trash":
		return "󰩺"
	}
	// Match by name for folders without a role
	lower := strings.ToLower(f.Name)
	switch {
	case strings.Contains(lower, "notification"):
		return "󰂚"
	case strings.Contains(lower, "remind"):
		return "󰑴"
	case strings.HasPrefix(lower, "lists/") || strings.HasPrefix(lower, "list/"):
		return "󰡡"
	default:
		return "󰡡"
	}
}

// classifyGroup assigns a folder to its visual group.
func classifyGroup(f mail.Folder) folderGroup {
	switch strings.ToLower(f.Role) {
	case "inbox", "drafts", "sent", "archive":
		return primaryGroup
	case "junk", "trash":
		return disposalGroup
	default:
		return customGroup
	}
}

// NewSidebar creates a Sidebar from a folder list.
func NewSidebar(styles Styles, folders []mail.Folder, width, height int) Sidebar {
	entries := make([]folderEntry, len(folders))
	for i, f := range folders {
		entries[i] = folderEntry{
			folder: f,
			icon:   folderIcon(f),
			group:  classifyGroup(f),
		}
	}
	return Sidebar{
		entries:  entries,
		selected: 0,
		focused:  true,
		styles:   styles,
		width:    width,
		height:   height,
	}
}

// Selected returns the index of the currently selected folder.
func (s Sidebar) Selected() int { return s.selected }

// SelectedFolder returns the name of the currently selected folder.
func (s Sidebar) SelectedFolder() string {
	if s.selected < len(s.entries) {
		return s.entries[s.selected].folder.Name
	}
	return ""
}

// SelectedIcon returns the icon of the currently selected folder.
func (s Sidebar) SelectedIcon() string {
	if s.selected < len(s.entries) {
		return s.entries[s.selected].icon
	}
	return ""
}

// SetFocused sets whether the sidebar has focus.
func (s *Sidebar) SetFocused(focused bool) { s.focused = focused }

// SetSize updates the sidebar dimensions.
func (s *Sidebar) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// MoveUp moves the selection up by one, skipping group separators.
func (s *Sidebar) MoveUp() {
	if s.selected > 0 {
		s.selected--
	}
}

// MoveDown moves the selection down by one.
func (s *Sidebar) MoveDown() {
	if s.selected < len(s.entries)-1 {
		s.selected++
	}
}

// MoveToTop moves the selection to the first folder.
func (s *Sidebar) MoveToTop() { s.selected = 0 }

// MoveToBottom moves the selection to the last folder.
func (s *Sidebar) MoveToBottom() {
	if len(s.entries) > 0 {
		s.selected = len(s.entries) - 1
	}
}

// View renders the sidebar as a vertical list of folder rows.
func (s Sidebar) View() string {
	if len(s.entries) == 0 || s.width == 0 || s.height == 0 {
		return ""
	}

	var lines []string
	prevGroup := s.entries[0].group

	for i, entry := range s.entries {
		// Add blank separator line between groups
		if i > 0 && entry.group != prevGroup {
			lines = append(lines, s.renderBlankLine())
			prevGroup = entry.group
		}

		lines = append(lines, s.renderRow(i, entry))
		prevGroup = entry.group
	}

	// Pad to height
	for len(lines) < s.height {
		lines = append(lines, s.renderBlankLine())
	}

	// Truncate to height
	if len(lines) > s.height {
		lines = lines[:s.height]
	}

	return strings.Join(lines, "\n")
}

// renderRow renders a single folder row.
func (s Sidebar) renderRow(idx int, entry folderEntry) string {
	isSelected := idx == s.selected
	hasUnread := entry.folder.Unseen > 0

	// Build indicator: ┃ when focused+selected, space otherwise
	indicator := " "
	if isSelected && s.focused {
		indicator = s.styles.SidebarIndicator.Render("┃")
	}

	// Icon: accent_tertiary when unread, fg_base otherwise
	var icon string
	if hasUnread {
		icon = s.styles.SidebarIconUnread.Render(entry.icon)
	} else {
		icon = s.styles.SidebarFolder.Render(entry.icon)
	}

	// Folder name
	var name string
	if hasUnread {
		name = s.styles.SidebarFolderUnread.Render(entry.folder.Name)
	} else {
		name = s.styles.SidebarFolder.Render(entry.folder.Name)
	}

	// Unread count (right-aligned, only when > 0)
	countStr := ""
	if hasUnread {
		countStr = s.styles.SidebarCount.Render(fmt.Sprintf("%d", entry.folder.Unseen))
	}

	// Layout: " ┃ icon  name          count "
	// indicator(1) + space(1) + icon(2) + space(2) + name + gap + count + space(1)
	leftContent := indicator + " " + icon + "  " + name
	leftWidth := lipgloss.Width(leftContent)
	countWidth := lipgloss.Width(countStr)

	// Right padding: fill between name and count, then 1 char margin
	rightMargin := 1
	gap := maxInt(1, s.width-leftWidth-countWidth-rightMargin)

	row := leftContent + strings.Repeat(" ", gap) + countStr
	// Pad to exact width
	rowWidth := lipgloss.Width(row)
	if rowWidth < s.width {
		row += strings.Repeat(" ", s.width-rowWidth)
	}

	// Apply selection background
	if isSelected {
		row = s.styles.SidebarSelected.Width(s.width).Render(
			stripANSIForStyle(row))
		// Re-render with background — need to rebuild with bg
		// Actually, use lipgloss overlay approach
	}

	return row
}

// renderBlankLine renders an empty line at the sidebar width.
func (s Sidebar) renderBlankLine() string {
	return strings.Repeat(" ", s.width)
}
```

**Note:** The `renderRow` selection background approach above has a flaw — you can't easily layer `bg_selection` on top of already-styled ANSI text. Instead, when the row is selected, render each segment with the selection background added:

Replace the `renderRow` method with this corrected version:

```go
// renderRow renders a single folder row.
func (s Sidebar) renderRow(idx int, entry folderEntry) string {
	isSelected := idx == s.selected
	hasUnread := entry.folder.Unseen > 0

	// When selected, all styles get bg_selection background
	var bgOpt lipgloss.Style
	if isSelected {
		bgOpt = s.styles.SidebarSelected
	}

	// Build indicator: ┃ when focused+selected, space otherwise
	indicator := " "
	if isSelected && s.focused {
		style := s.styles.SidebarIndicator
		if isSelected {
			style = style.Background(bgOpt.GetBackground())
		}
		indicator = style.Render("┃")
	} else if isSelected {
		indicator = bgOpt.Render(" ")
	}

	// Icon
	iconStyle := s.styles.SidebarFolder
	if hasUnread {
		iconStyle = s.styles.SidebarIconUnread
	}
	if isSelected {
		iconStyle = iconStyle.Background(bgOpt.GetBackground())
	}
	icon := iconStyle.Render(entry.icon)

	// Folder name
	nameStyle := s.styles.SidebarFolder
	if hasUnread {
		nameStyle = s.styles.SidebarFolderUnread
	}
	if isSelected {
		nameStyle = nameStyle.Background(bgOpt.GetBackground())
	}
	name := nameStyle.Render(entry.folder.Name)

	// Unread count
	countStr := ""
	countWidth := 0
	if hasUnread {
		countStyle := s.styles.SidebarCount
		if isSelected {
			countStyle = countStyle.Background(bgOpt.GetBackground())
		}
		countStr = countStyle.Render(fmt.Sprintf("%d", entry.folder.Unseen))
		countWidth = lipgloss.Width(countStr)
	}

	// Spacers with correct background
	spaceStyle := lipgloss.NewStyle()
	if isSelected {
		spaceStyle = spaceStyle.Background(bgOpt.GetBackground())
	}

	// Layout: indicator(1) + sp(1) + icon(~2) + sp(2) + name + gap + count + margin(1)
	leftContent := indicator + spaceStyle.Render(" ") + icon + spaceStyle.Render("  ") + name
	leftWidth := lipgloss.Width(leftContent)

	rightMargin := 1
	gap := maxInt(1, s.width-leftWidth-countWidth-rightMargin)

	row := leftContent + spaceStyle.Render(strings.Repeat(" ", gap)) + countStr + spaceStyle.Render(strings.Repeat(" ", rightMargin))

	// Ensure exact width
	rowWidth := lipgloss.Width(row)
	if rowWidth < s.width {
		row += spaceStyle.Render(strings.Repeat(" ", s.width-rowWidth))
	}

	return row
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/ui/ -run TestSidebar -v
```

Expected: PASS for all sidebar tests.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/sidebar.go internal/ui/sidebar_test.go
git commit -m "Add Sidebar model with folder groups, selection, and unread badges"
```

---

### Task 3: Add Navigation Tests and Logic

**Files:**
- Modify: `internal/ui/sidebar_test.go`
- Modify: `internal/ui/sidebar.go`

- [ ] **Step 1: Write failing navigation tests**

Add to `TestSidebar` in `sidebar_test.go`:

```go
t.Run("j moves down", func(t *testing.T) {
	sb := NewSidebar(styles, folders, 30, 20)
	sb.MoveDown()
	if sb.Selected() != 1 {
		t.Errorf("after MoveDown, selected = %d, want 1", sb.Selected())
	}
	if sb.SelectedFolder() != "Drafts" {
		t.Errorf("selected folder = %q, want Drafts", sb.SelectedFolder())
	}
})

t.Run("k moves up", func(t *testing.T) {
	sb := NewSidebar(styles, folders, 30, 20)
	sb.MoveDown()
	sb.MoveDown()
	sb.MoveUp()
	if sb.Selected() != 1 {
		t.Errorf("after Down+Down+Up, selected = %d, want 1", sb.Selected())
	}
})

t.Run("k at top stays at 0", func(t *testing.T) {
	sb := NewSidebar(styles, folders, 30, 20)
	sb.MoveUp()
	if sb.Selected() != 0 {
		t.Errorf("MoveUp at top: selected = %d, want 0", sb.Selected())
	}
})

t.Run("j at bottom stays at last", func(t *testing.T) {
	sb := NewSidebar(styles, folders, 30, 20)
	for i := 0; i < 20; i++ {
		sb.MoveDown()
	}
	last := len(folders) - 1
	if sb.Selected() != last {
		t.Errorf("MoveDown past end: selected = %d, want %d", sb.Selected(), last)
	}
})

t.Run("gg moves to top", func(t *testing.T) {
	sb := NewSidebar(styles, folders, 30, 20)
	sb.MoveDown()
	sb.MoveDown()
	sb.MoveDown()
	sb.MoveToTop()
	if sb.Selected() != 0 {
		t.Errorf("MoveToTop: selected = %d, want 0", sb.Selected())
	}
})

t.Run("G moves to bottom", func(t *testing.T) {
	sb := NewSidebar(styles, folders, 30, 20)
	sb.MoveToBottom()
	last := len(folders) - 1
	if sb.Selected() != last {
		t.Errorf("MoveToBottom: selected = %d, want %d", sb.Selected(), last)
	}
})

```

- [ ] **Step 2: Run tests**

```bash
go test ./internal/ui/ -run TestSidebar -v
```

Expected: PASS — `MoveUp`, `MoveDown`, `MoveToTop`, `MoveToBottom` are already defined in Task 2.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/sidebar_test.go
git commit -m "Add sidebar navigation tests"
```

---

### Task 4: Wire Sidebar into AccountTab

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Write failing tests for sidebar integration**

Add/update tests in `account_tab_test.go`:

```go
t.Run("view renders folder names", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	tab.width = 80
	tab.height = 20
	view := tab.View()
	plain := stripANSI(view)
	if !strings.Contains(plain, "Inbox") {
		t.Error("missing Inbox in sidebar")
	}
	if !strings.Contains(plain, "Drafts") {
		t.Error("missing Drafts in sidebar")
	}
	if !strings.Contains(plain, "Archive") {
		t.Error("missing Archive in sidebar")
	}
})

t.Run("j/k navigates sidebar when focused", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	tab.width = 80
	tab.height = 20
	// Sidebar is focused by default
	tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tab.sidebar.SelectedFolder() != "Drafts" {
		t.Errorf("after j, selected = %q, want Drafts", tab.sidebar.SelectedFolder())
	}
	tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if tab.sidebar.SelectedFolder() != "Inbox" {
		t.Errorf("after k, selected = %q, want Inbox", tab.sidebar.SelectedFolder())
	}
})

t.Run("title tracks selected folder", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	tab.width = 80
	tab.height = 20
	tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if tab.Title() != "Drafts" {
		t.Errorf("Title() = %q, want Drafts", tab.Title())
	}
})

t.Run("sidebar unfocused when msglist focused", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	tab.width = 80
	tab.height = 20
	tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyTab})
	if tab.sidebar.focused {
		t.Error("sidebar should be unfocused after Tab")
	}
})
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/ui/ -run TestAccountTab -v
```

Expected: FAIL — `tab.sidebar` field doesn't exist yet.

- [ ] **Step 3: Replace placeholder with Sidebar in AccountTab**

Rewrite `account_tab.go`:

```go
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/mail"
)

// Panel identifies which panel of the AccountTab is focused.
type Panel int

const (
	SidebarPanel Panel = iota
	MsgListPanel
)

// sidebarWidth is the fixed width of the sidebar panel.
const sidebarWidth = 30

// AccountTab implements Tab for the main account view with sidebar and message list.
type AccountTab struct {
	styles  Styles
	backend mail.Backend
	focused Panel
	sidebar Sidebar
	width   int
	height  int
}

// NewAccountTab creates an AccountTab using the given styles and backend.
func NewAccountTab(styles Styles, backend mail.Backend) AccountTab {
	folders, _ := backend.ListFolders()
	sb := NewSidebar(styles, folders, sidebarWidth, 1)

	return AccountTab{
		styles:  styles,
		backend: backend,
		focused: SidebarPanel,
		sidebar: sb,
	}
}

// Title returns the current folder name.
func (m AccountTab) Title() string { return m.sidebar.SelectedFolder() }

// Icon returns the folder's Nerd Font icon.
func (m AccountTab) Icon() string { return m.sidebar.SelectedIcon() }

// Closeable returns false — the account tab cannot be closed.
func (m AccountTab) Closeable() bool { return false }

// Init returns no initial command.
func (m AccountTab) Init() tea.Cmd { return nil }

// Update satisfies tea.Model. Delegates to updateTab for typed access.
func (m AccountTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.updateTab(msg)
}

// updateTab handles key events and window size changes, returning the typed model.
func (m AccountTab) updateTab(msg tea.Msg) (AccountTab, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		sw := minInt(sidebarWidth, m.width/2)
		m.sidebar.SetSize(sw, m.height)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab:
			if m.focused == SidebarPanel {
				m.focused = MsgListPanel
			} else {
				m.focused = SidebarPanel
			}
			m.sidebar.SetFocused(m.focused == SidebarPanel)

		default:
			if m.focused == SidebarPanel {
				m.handleSidebarKey(msg)
			}
		}
	}
	return m, nil
}

// handleSidebarKey routes key events to sidebar actions.
func (m *AccountTab) handleSidebarKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "j", "down":
		m.sidebar.MoveDown()
	case "k", "up":
		m.sidebar.MoveUp()
	case "G":
		m.sidebar.MoveToBottom()
	}
}

// View renders the sidebar + divider + message list placeholder.
func (m AccountTab) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	sw := minInt(sidebarWidth, m.width/2)
	mw := m.width - sw - 1 // -1 for divider

	sidebarView := m.sidebar.View()
	divider := renderDivider(m.height, m.styles)
	msglistView := renderPlaceholder("Message List", mw, m.height, m.focused == MsgListPanel, m.styles)

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, divider, msglistView)
}

// renderDivider renders a vertical line of │ characters.
func renderDivider(height int, s Styles) string {
	div := s.PanelDivider.Render("│")
	lines := make([]string, height)
	for i := range lines {
		lines[i] = div
	}
	return strings.Join(lines, "\n")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
```

Remove the `renderPlaceholder` function's sidebar usage but keep it for the message list placeholder (it's still needed for 2.5b-3 replacement later).

- [ ] **Step 4: Update existing tests that check old behavior**

Update the "title returns folder name" test (now sources from sidebar), and the "view renders two panels with divider" test:

```go
t.Run("title returns folder name", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	if tab.Title() != "Inbox" {
		t.Errorf("Title() = %q, want Inbox", tab.Title())
	}
})

t.Run("view renders two panels with divider", func(t *testing.T) {
	tab := NewAccountTab(styles, backend)
	tab.width = 80
	tab.height = 20
	result := tab.View()
	plain := stripANSI(result)
	if !strings.Contains(plain, "Inbox") {
		t.Error("missing Inbox in sidebar")
	}
	if !strings.Contains(plain, "│") {
		t.Error("missing panel divider")
	}
})
```

- [ ] **Step 5: Run all tests**

```bash
go test ./internal/ui/ -v
```

Expected: PASS for all tests.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "Wire Sidebar into AccountTab, replace placeholder with real folder list"
```

---

### Task 5: Update App and Status Bar Integration

**Files:**
- Modify: `internal/ui/app.go`

- [ ] **Step 1: Update status bar to track sidebar selection**

The status bar currently gets its folder info in `NewApp`. Add a method to update it when the sidebar selection changes. In `app.go`, update the `updateFooterContext` method to also sync the status bar:

```go
func (m *App) updateFooterContext() {
	if acct, ok := m.tabs[m.activeTab].(AccountTab); ok {
		if acct.focused == SidebarPanel {
			m.footer.SetContext(SidebarContext)
		} else {
			m.footer.SetContext(MsgListContext)
		}
		// Sync status bar with sidebar selection
		folders, _ := acct.backend.ListFolders()
		for _, f := range folders {
			if f.Name == acct.sidebar.SelectedFolder() {
				m.statusBar.SetFolder(acct.sidebar.SelectedIcon(), f.Name, f.Exists, f.Unseen)
				break
			}
		}
	}
}
```

- [ ] **Step 2: Run all tests**

```bash
go test ./internal/ui/ -v
```

Expected: PASS.

- [ ] **Step 3: Run make check**

```bash
make check
```

Expected: PASS — vet and all tests clean.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/app.go
git commit -m "Sync status bar with sidebar folder selection"
```

---

### Task 6: Mode 2 Visual Verification

**Files:**
- Read: `internal/ui/app_test.go`

- [ ] **Step 1: Add a snapshot test for visual review**

Add a test to `app_test.go` that dumps the stripped grid for Mode 2 inspection:

```go
t.Run("sidebar renders in composite layout", func(t *testing.T) {
	app := NewApp(theme.Nord, backend)
	app, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	view := app.View()
	plain := stripANSI(view)
	lines := strings.Split(plain, "\n")

	// Sidebar should show folder names in first 30 columns
	for _, name := range []string{"Inbox", "Drafts", "Sent", "Archive", "Spam", "Trash"} {
		found := false
		for _, line := range lines {
			runes := []rune(line)
			if len(runes) >= 30 {
				sidebarPart := string(runes[:30])
				if strings.Contains(sidebarPart, name) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("folder %q not found in sidebar region", name)
		}
	}

	// Divider column should have │ on content rows
	// (reuses the vertical line test logic)
	row3 := []rune(strings.TrimRight(lines[2], " "))
	dividerCol := -1
	for i, r := range row3 {
		if r == '┬' {
			dividerCol = i
			break
		}
	}
	if dividerCol >= 0 {
		for i := 3; i < len(lines)-2; i++ {
			runes := []rune(lines[i])
			if dividerCol < len(runes) && runes[dividerCol] != '│' {
				t.Errorf("line %d: char at divider col %d = %c, want │",
					i+1, dividerCol, runes[dividerCol])
			}
		}
	}
})
```

- [ ] **Step 2: Run the test and review output**

```bash
go test ./internal/ui/ -run "TestApp/sidebar_renders" -v
```

Expected: PASS. If any positional check fails, debug using Mode 1 (check exact column positions).

- [ ] **Step 3: Commit**

```bash
git add internal/ui/app_test.go
git commit -m "Add composite layout test with sidebar region checks"
```

---

### Task 7: Build and Live Test

**Files:** None (verification only)

- [ ] **Step 1: Build the binary**

```bash
make build
```

- [ ] **Step 2: Run poplar and visually inspect**

```bash
./poplar
```

Verify:
- Sidebar shows folder names with icons
- Three groups separated by blank lines
- Inbox selected by default with `┃` indicator
- j/k moves selection up/down
- G jumps to bottom, gg to top
- Tab cycles focus — `┃` appears/disappears
- Status bar updates when selection changes
- Unread counts visible for Inbox (3) and Spam (12)
- q exits cleanly

- [ ] **Step 3: Install**

```bash
make install
```

- [ ] **Step 4: Final commit if any fixes were needed**

If any visual issues were found and fixed during live testing, commit the fixes.

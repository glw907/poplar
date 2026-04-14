# Sidebar Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship local filter-and-hide message search for Pass 2.5b-7, hosted in a 3-row shelf pinned to the bottom of the sidebar column. Default `[name]` mode matches subject+sender; `[all]` mode adds date text. Thread-level predicate — any match keeps the whole thread visible.

**Architecture:** A new `SidebarSearch` subcomponent wraps `bubbles/textinput` and owns search state (Idle/Typing/Active). `MessageList` gains a `filter` field that adds one step to the group→sort→flatten pipeline. `AccountTab` bridges them through new `SearchUpdatedMsg` / `SearchCommittedMsg` / `SearchClearedMsg` / `SearchResultsMsg` types; neither child reaches into the other. The account view is composed top-to-bottom as `header(2) → folder region(flex, scrollable) → search shelf(3, pinned)`.

**Tech Stack:** Go 1.25, bubbletea, lipgloss, `github.com/charmbracelet/bubbles/textinput` (already transitively in `go.mod` via the top-level `bubbles` dep).

**Spec:** `docs/superpowers/specs/2026-04-13-sidebar-search-design.md`

---

## File Structure

**New files:**

- `internal/ui/sidebar_search.go` — `SidebarSearch` subcomponent: state machine, textinput wrapper, rendering
- `internal/ui/sidebar_search_test.go` — unit tests for states, transitions, rendering, mode cycle
- `docs/poplar/decisions/0064-sidebar-search-shelf.md` — ADR

**Modified files:**

- `internal/ui/cmds.go` — add `SearchMode`, `SearchState`, `SearchActivateMsg`, `SearchUpdatedMsg`, `SearchCommittedMsg`, `SearchClearedMsg`, `SearchResultsMsg`
- `internal/ui/styles.go` — add 7 new style slots; populate in `NewStyles`
- `internal/ui/styles_test.go` — test new slots are populated
- `internal/ui/msglist.go` — add `filter` field, `searchFilter` type, `SetFilter`/`ClearFilter`/`FilterResultCount`, filter step in `rebuild`, cursor save/restore, fold shadowing, "No matches" placeholder wording
- `internal/ui/msglist_test.go` — filter predicate, mode gating, cursor save/restore, thread-level, fold preservation, result count, placeholder wording
- `internal/ui/account_tab.go` — `searchShelfRows` const, `sidebarSearch` field, height math, View composition, key routing for all search states, folder-jump clears
- `internal/ui/account_tab_test.go` — key routing, state transitions, folder-jump clears, q-stolen, fold no-op
- `docs/poplar/invariants.md` — 2 new UX facts (edited in place), decision index row
- `docs/poplar/wireframes.md` — new §2.1 Sidebar Search, updated §1 Composite, replaced §7 #15
- `docs/poplar/keybindings.md` — promote `/`, `n`, `N` to live; search-context routing notes; design decision note
- `docs/poplar/STATUS.md` — drop 2.5b-3.7, promote 2.5b-7 as current pass, rewrite starter prompt for 2.5b-4

---

## Phase 1 — Foundation

### Task 1: Add search enums and Msg types

**Files:**
- Modify: `internal/ui/cmds.go` (append)

- [ ] **Step 1: Add new types and Msg structs to `cmds.go`**

Append to `internal/ui/cmds.go` (after the existing `folderChangedCmd`):

```go
// SearchMode selects which fields the message filter matches against.
type SearchMode int

const (
	// SearchModeName matches subject + sender. Default.
	SearchModeName SearchMode = iota
	// SearchModeAll matches subject + sender + date text.
	SearchModeAll
)

// SearchState is the lifecycle state of the sidebar search UI.
type SearchState int

const (
	// SearchIdle — no filter, shelf shows hint row.
	SearchIdle SearchState = iota
	// SearchTyping — prompt focused, printable runes append to query,
	// filter updates live on each keystroke.
	SearchTyping
	// SearchActive — query is live but prompt is unfocused; normal
	// account-view key routing resumes.
	SearchActive
)

// SearchUpdatedMsg is emitted by SidebarSearch.Update on each
// keystroke or mode change in Typing state, carrying the current
// query and mode. AccountTab handles it by calling
// MessageList.SetFilter and then pushing the thread count into
// SidebarSearch.SetResultCount.
type SearchUpdatedMsg struct {
	Query string
	Mode  SearchMode
}
```

- [ ] **Step 2: Run the package build to verify syntax**

Run: `go build ./internal/ui/...`
Expected: exits 0, no output.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/cmds.go
git commit -m "ui: add search Msg types and state enums

Pass 2.5b-7 groundwork. No behavior change yet — these types are
consumed by SidebarSearch and MessageList in later tasks."
```

---

### Task 2: Add new style slots for the search shelf

**Files:**
- Modify: `internal/ui/styles.go`
- Modify: `internal/ui/styles_test.go`

- [ ] **Step 1: Write a failing test for the new slots**

Append to `internal/ui/styles_test.go`:

```go
func TestSearchStyles(t *testing.T) {
	styles := NewStyles(theme.Nord)

	checks := map[string]lipgloss.Style{
		"SearchIcon":        styles.SearchIcon,
		"SearchHint":        styles.SearchHint,
		"SearchPrompt":      styles.SearchPrompt,
		"SearchModeBadge":   styles.SearchModeBadge,
		"SearchResultCount": styles.SearchResultCount,
		"SearchNoResults":   styles.SearchNoResults,
		"MsgListPlaceholder": styles.MsgListPlaceholder,
	}
	for name, s := range checks {
		if s.GetForeground() == nil {
			t.Errorf("%s has no foreground color", name)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestSearchStyles -v`
Expected: compile error — the `Styles` struct does not have these fields yet.

- [ ] **Step 3: Add the style fields to the struct**

Edit `internal/ui/styles.go`. In the `Styles` struct, after the `Dim lipgloss.Style` line, insert:

```go
	// Search shelf and search-related placeholder
	SearchIcon         lipgloss.Style
	SearchHint         lipgloss.Style
	SearchPrompt       lipgloss.Style
	SearchModeBadge    lipgloss.Style
	SearchResultCount  lipgloss.Style
	SearchNoResults    lipgloss.Style
	MsgListPlaceholder lipgloss.Style
```

- [ ] **Step 4: Populate the new slots in `NewStyles`**

In `internal/ui/styles.go`, in `NewStyles`, after the `Dim:` field and before the `TopLine:` field, insert:

```go
		SearchIcon: lipgloss.NewStyle().
			Foreground(t.FgDim),
		SearchHint: lipgloss.NewStyle().
			Foreground(t.FgDim),
		SearchPrompt: lipgloss.NewStyle().
			Foreground(t.FgBase),
		SearchModeBadge: lipgloss.NewStyle().
			Foreground(t.FgDim),
		SearchResultCount: lipgloss.NewStyle().
			Foreground(t.AccentTertiary),
		SearchNoResults: lipgloss.NewStyle().
			Foreground(t.ColorWarning),
		MsgListPlaceholder: lipgloss.NewStyle().
			Foreground(t.FgDim),
```

- [ ] **Step 5: Run the test to verify it passes**

Run: `go test ./internal/ui/ -run TestSearchStyles -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/ui/styles.go internal/ui/styles_test.go
git commit -m "ui: add search shelf and placeholder style slots

Seven new Styles fields for Pass 2.5b-7: SearchIcon, SearchHint,
SearchPrompt, SearchModeBadge, SearchResultCount, SearchNoResults,
MsgListPlaceholder. Populated from the existing theme palette — no
new colors."
```

---

## Phase 2 — MessageList filter

### Task 3: Add filter field and subject/sender match ([name] mode)

**Files:**
- Modify: `internal/ui/msglist.go`
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Write a failing test for the filter**

Append to `internal/ui/msglist_test.go`:

```go
func TestMessageListFilter(t *testing.T) {
	styles := NewStyles(theme.Nord)

	msgs := []mail.MessageInfo{
		{UID: "1", ThreadID: "1", Subject: "Project update", From: "Alice", Date: "Apr 10"},
		{UID: "2", ThreadID: "2", Subject: "Weekend plans", From: "Bob", Date: "Apr 09"},
		{UID: "3", ThreadID: "3", Subject: "Invoice #2847", From: "Billing", Date: "Apr 08"},
	}

	t.Run("empty query keeps all rows", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("", SearchModeName)
		if got := len(ml.rows); got != 3 {
			t.Errorf("len(rows) after empty filter = %d, want 3", got)
		}
	})

	t.Run("substring match on subject", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("project", SearchModeName)
		if got := len(ml.rows); got != 1 {
			t.Errorf("len(rows) = %d, want 1", got)
		}
		if ml.rows[0].msg.UID != "1" {
			t.Errorf("matched row = %q, want 1", ml.rows[0].msg.UID)
		}
	})

	t.Run("substring match on sender", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("bob", SearchModeName)
		if got := len(ml.rows); got != 1 {
			t.Errorf("len(rows) = %d, want 1", got)
		}
		if ml.rows[0].msg.UID != "2" {
			t.Errorf("matched row = %q, want 2", ml.rows[0].msg.UID)
		}
	})

	t.Run("case-insensitive", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("ALICE", SearchModeName)
		if got := len(ml.rows); got != 1 {
			t.Errorf("len(rows) = %d, want 1", got)
		}
	})

	t.Run("no matches returns empty rows", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("zzz-nothing", SearchModeName)
		if got := len(ml.rows); got != 0 {
			t.Errorf("len(rows) = %d, want 0", got)
		}
	})

	t.Run("ClearFilter restores all rows", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("project", SearchModeName)
		ml.ClearFilter()
		if got := len(ml.rows); got != 3 {
			t.Errorf("len(rows) after clear = %d, want 3", got)
		}
	})

	t.Run("[name] mode does not match date", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("Apr 10", SearchModeName)
		if got := len(ml.rows); got != 0 {
			t.Errorf("len(rows) for Apr 10 under [name] = %d, want 0", got)
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestMessageListFilter -v`
Expected: compile error — `SetFilter`, `ClearFilter` are undefined.

- [ ] **Step 3: Add the filter field and type to `MessageList`**

Edit `internal/ui/msglist.go`. Replace the `MessageList` struct (lines 74-84) with:

```go
type MessageList struct {
	source          []mail.MessageInfo
	rows            []displayRow
	folded          map[mail.UID]bool
	sort            SortOrder
	selected        int
	offset          int
	styles          Styles
	width           int
	height          int
	filter          searchFilter
	preSearchCursor int
	savedByFilter   bool
}

// searchFilter holds the active filter's query and mode. The zero
// value (empty query, SearchModeName) means "no filter."
type searchFilter struct {
	query string
	mode  SearchMode
}
```

- [ ] **Step 4: Add `SetFilter`, `ClearFilter`, and the filter pipeline step**

In `internal/ui/msglist.go`, add these methods before `SetSort` (around line 327):

```go
// SetFilter applies a search filter to the message list, rebuilding
// the display rows through the filterBuckets pipeline step. On the
// first transition from unfiltered to filtered, saves the pre-search
// cursor row so ClearFilter can restore it. Subsequent keystrokes do
// not overwrite the saved row — the save gate stays armed until clear.
func (m *MessageList) SetFilter(q string, mode SearchMode) {
	if !m.savedByFilter && q != "" {
		m.preSearchCursor = m.selected
		m.savedByFilter = true
	}
	m.filter = searchFilter{query: q, mode: mode}
	m.rebuild()
	m.clampOffset()
}

// ClearFilter removes any active filter, rebuilds rows, and restores
// the pre-search cursor row if one was saved. A cursor that points
// past the new end of rows clamps to 0.
func (m *MessageList) ClearFilter() {
	m.filter = searchFilter{}
	m.rebuild()
	if m.savedByFilter {
		m.selected = m.preSearchCursor
		if m.selected >= len(m.rows) {
			m.selected = 0
		}
		m.savedByFilter = false
	}
	m.clampOffset()
}
```

- [ ] **Step 5: Wire `filterBuckets` into `rebuild`**

In `internal/ui/msglist.go`, replace the `rebuild` function body (lines 121-136) with:

```go
func (m *MessageList) rebuild() {
	buckets := bucketByThreadID(m.source)
	buckets = m.filterBuckets(buckets)
	sort.SliceStable(buckets, func(i, j int) bool {
		if m.sort == SortDateAsc {
			return latestActivity(buckets[i]) < latestActivity(buckets[j])
		}
		return latestActivity(buckets[i]) > latestActivity(buckets[j])
	})

	rows := make([]displayRow, 0, len(m.source))
	for _, bucket := range buckets {
		rows = appendThreadRows(rows, bucket)
	}
	applyFoldState(rows, m.folded)
	m.rows = rows
}
```

- [ ] **Step 6: Add the `filterBuckets` helper and match predicates**

In `internal/ui/msglist.go`, add these helpers after `bucketByThreadID` (around line 159):

```go
// filterBuckets is the filter step of the build pipeline. When the
// filter query is empty, it returns buckets unchanged. When non-empty,
// it keeps any bucket containing at least one matching message — the
// thread-level predicate from ADR 0064.
func (m *MessageList) filterBuckets(buckets [][]mail.MessageInfo) [][]mail.MessageInfo {
	if m.filter.query == "" {
		return buckets
	}
	q := strings.ToLower(m.filter.query)
	out := buckets[:0]
	for _, bucket := range buckets {
		for _, msg := range bucket {
			if matchMessage(msg, q, m.filter.mode) {
				out = append(out, bucket)
				break
			}
		}
	}
	return out
}

// matchMessage tests one message against a pre-lowercased query under
// the given mode. [name] matches subject + sender; [all] additionally
// matches the date text.
func matchMessage(msg mail.MessageInfo, lowerQuery string, mode SearchMode) bool {
	if containsFold(msg.Subject, lowerQuery) {
		return true
	}
	if containsFold(msg.From, lowerQuery) {
		return true
	}
	if mode == SearchModeAll && containsFold(msg.Date, lowerQuery) {
		return true
	}
	return false
}

// containsFold tests whether s (any case) contains lowerNeedle
// (already lowercased by the caller). Lowercase-once-per-field is
// fine for the small mock datasets we test against; Pass 3 scale
// profiling may argue for a cached lowered-field table on MessageList.
func containsFold(s, lowerNeedle string) bool {
	return strings.Contains(strings.ToLower(s), lowerNeedle)
}
```

- [ ] **Step 7: Run the test to verify it passes**

Run: `go test ./internal/ui/ -run TestMessageListFilter -v`
Expected: all subtests PASS.

- [ ] **Step 8: Run the full msglist test suite to verify no regressions**

Run: `go test ./internal/ui/ -run TestMessageList -v`
Expected: all existing tests still PASS.

- [ ] **Step 9: Commit**

```bash
git add internal/ui/msglist.go internal/ui/msglist_test.go
git commit -m "MessageList: add filter field and [name] mode predicate

Adds SetFilter/ClearFilter with thread-level filter semantics and
cursor save/restore. Filter step slots between bucketing and thread
sorting in the existing rebuild pipeline. [all] mode gating and fold
shadowing land in follow-up commits."
```

---

### Task 4: [all] mode matches date text

**Files:**
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Write a failing test for [all] mode**

Append to `TestMessageListFilter` in `internal/ui/msglist_test.go`:

```go
	t.Run("[all] mode matches date", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("Apr 10", SearchModeAll)
		if got := len(ml.rows); got != 1 {
			t.Errorf("len(rows) for Apr 10 under [all] = %d, want 1", got)
		}
		if ml.rows[0].msg.UID != "1" {
			t.Errorf("matched row = %q, want 1", ml.rows[0].msg.UID)
		}
	})

	t.Run("[all] mode also matches subject and sender", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("project", SearchModeAll)
		if got := len(ml.rows); got != 1 {
			t.Errorf("len(rows) for project under [all] = %d, want 1", got)
		}
	})

	t.Run("[all] and [name] differ on date-only queries", func(t *testing.T) {
		mlName := NewMessageList(styles, msgs, 90, 20)
		mlAll := NewMessageList(styles, msgs, 90, 20)
		mlName.SetFilter("Apr 09", SearchModeName)
		mlAll.SetFilter("Apr 09", SearchModeAll)
		if len(mlName.rows) != 0 {
			t.Errorf("[name] matched date: len(rows) = %d, want 0", len(mlName.rows))
		}
		if len(mlAll.rows) != 1 {
			t.Errorf("[all] missed date: len(rows) = %d, want 1", len(mlAll.rows))
		}
	})
```

- [ ] **Step 2: Run the test to verify it passes**

Run: `go test ./internal/ui/ -run TestMessageListFilter -v`
Expected: all subtests PASS. The `matchMessage` helper already gates
date matching by mode; these new tests exercise that gating.

- [ ] **Step 3: Commit**

```bash
git add internal/ui/msglist_test.go
git commit -m "MessageList: test [all] mode date matching and mode differentiation

Exercises the SearchMode gating in matchMessage — date text is
searchable under [all] but not [name]."
```

---

### Task 5: Cursor save/restore with savedByFilter gate

**Files:**
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Write failing tests for cursor save/restore**

Append to `TestMessageListFilter` in `internal/ui/msglist_test.go`:

```go
	t.Run("cursor saved on first filter application", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.MoveDown()
		ml.MoveDown()  // cursor now on row 2 (UID 3)
		if ml.selected != 2 {
			t.Fatalf("setup: selected = %d, want 2", ml.selected)
		}
		ml.SetFilter("project", SearchModeName)
		if ml.preSearchCursor != 2 {
			t.Errorf("preSearchCursor = %d, want 2", ml.preSearchCursor)
		}
	})

	t.Run("subsequent keystrokes don't overwrite saved cursor", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.MoveDown()
		ml.MoveDown()
		ml.SetFilter("p", SearchModeName)
		// cursor may have clamped to 0 after the filter narrowed rows;
		// simulate subsequent typing, which should leave preSearchCursor at 2
		ml.SetFilter("pr", SearchModeName)
		ml.SetFilter("pro", SearchModeName)
		if ml.preSearchCursor != 2 {
			t.Errorf("preSearchCursor after more typing = %d, want 2", ml.preSearchCursor)
		}
	})

	t.Run("clear restores pre-search cursor", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.MoveDown()
		ml.MoveDown()
		ml.SetFilter("project", SearchModeName)
		ml.ClearFilter()
		if ml.selected != 2 {
			t.Errorf("selected after clear = %d, want 2", ml.selected)
		}
	})

	t.Run("clear with invalid saved cursor clamps to 0", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.MoveDown()
		ml.MoveDown()  // cursor = 2
		ml.SetFilter("project", SearchModeName)
		// Replace source with a shorter list while filter is active
		ml.SetMessages(msgs[:1])
		// The SetMessages call rebuilds, losing saved state. Then
		// ClearFilter should not crash and should land on 0.
		ml.ClearFilter()
		if ml.selected != 0 {
			t.Errorf("selected after clear with shorter source = %d, want 0", ml.selected)
		}
	})

	t.Run("re-activating search after clear starts fresh save", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("project", SearchModeName)
		ml.ClearFilter()
		ml.MoveDown()  // cursor = 1
		ml.SetFilter("weekend", SearchModeName)
		if ml.preSearchCursor != 1 {
			t.Errorf("preSearchCursor on re-activate = %d, want 1", ml.preSearchCursor)
		}
	})
```

- [ ] **Step 2: Ensure `SetMessages` resets the savedByFilter gate**

Currently `SetMessages` (line 101-107) doesn't touch filter state.
When source is replaced mid-search (Pass 3 background refresh, or
the folder-jump clear path), the filter stays active and the save
gate stays armed. That's correct for the refresh case but not for
folder jumps — the folder-jump path calls `ClearFilter` first in
`AccountTab` (Task 18), so `SetMessages` on the new source sees an
already-idle filter.

For this task, update `SetMessages` to clear the filter and reset
the save gate defensively. This is a behavior change: any caller
that replaces `source` loses any active filter.

Edit `internal/ui/msglist.go`. Replace `SetMessages` (lines 101-107)
with:

```go
// SetMessages replaces the source slice and rebuilds the displayRow
// list. Resets fold state, cursor, viewport, and any active filter.
func (m *MessageList) SetMessages(msgs []mail.MessageInfo) {
	m.source = msgs
	m.folded = map[mail.UID]bool{}
	m.selected = 0
	m.offset = 0
	m.filter = searchFilter{}
	m.savedByFilter = false
	m.preSearchCursor = 0
	m.rebuild()
}
```

- [ ] **Step 3: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestMessageListFilter -v`
Expected: all subtests PASS.

- [ ] **Step 4: Commit**

```bash
git add internal/ui/msglist.go internal/ui/msglist_test.go
git commit -m "MessageList: save/restore cursor row across search

First filter application saves the current row into preSearchCursor.
Subsequent keystrokes do not overwrite the save — the gate stays
armed until ClearFilter or SetMessages. ClearFilter restores the
row, clamping to 0 if the saved index is now invalid."
```

---

### Task 6: Fold shadowing during active filter

**Files:**
- Modify: `internal/ui/msglist.go`
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Write failing tests for fold shadowing**

Append to `internal/ui/msglist_test.go`:

```go
func TestMessageListFilterFoldShadow(t *testing.T) {
	styles := NewStyles(theme.Nord)

	msgs := []mail.MessageInfo{
		{UID: "10", ThreadID: "T1", InReplyTo: "", Subject: "Server migration", From: "Eve", Date: "Apr 05"},
		{UID: "11", ThreadID: "T1", InReplyTo: "10", Subject: "Re: Server migration", From: "Grace", Date: "Apr 06"},
		{UID: "12", ThreadID: "T1", InReplyTo: "11", Subject: "Re: Server migration", From: "Frank", Date: "Apr 07"},
		{UID: "20", ThreadID: "T2", InReplyTo: "", Subject: "Lunch", From: "Carol", Date: "Apr 08"},
	}

	t.Run("filter expands folded thread when any message matches", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.FoldAll()  // fold T1 (3 msgs) — T2 is size-1 and not folded
		// Pre-filter: T1 folded, T2 visible. 2 visible rows (T1 root + T2 root).
		visibleBefore := 0
		for _, r := range ml.rows {
			if !r.hidden {
				visibleBefore++
			}
		}
		if visibleBefore != 2 {
			t.Fatalf("setup: visible rows = %d, want 2", visibleBefore)
		}

		ml.SetFilter("server", SearchModeName)
		// T1 matches (all messages), T2 does not. T1 should be fully
		// expanded regardless of saved fold state.
		visibleAfter := 0
		for _, r := range ml.rows {
			if !r.hidden {
				visibleAfter++
			}
		}
		if visibleAfter != 3 {
			t.Errorf("filtered visible rows = %d, want 3 (full T1)", visibleAfter)
		}
	})

	t.Run("clear filter restores saved fold state", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.FoldAll()
		ml.SetFilter("server", SearchModeName)
		ml.ClearFilter()

		// T1 root should be folded again; its children hidden.
		var rootRow displayRow
		var childCount int
		for _, r := range ml.rows {
			if r.isThreadRoot && r.msg.UID == "10" {
				rootRow = r
			}
			if !r.isThreadRoot && r.msg.ThreadID == "T1" && r.hidden {
				childCount++
			}
		}
		if !strings.HasPrefix(rootRow.prefix, "[") {
			t.Errorf("T1 root prefix after clear = %q, want folded badge", rootRow.prefix)
		}
		if childCount != 2 {
			t.Errorf("hidden children after clear = %d, want 2", childCount)
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run TestMessageListFilterFoldShadow -v`
Expected: FAIL — the first subtest fails because `applyFoldState`
doesn't know about the active filter; the folded root still has
children hidden.

- [ ] **Step 3: Gate fold application on an inactive filter**

Edit `internal/ui/msglist.go`. In `rebuild` (the function you
updated in Task 3), replace the `applyFoldState(rows, m.folded)`
line with:

```go
	if m.filter.query == "" {
		applyFoldState(rows, m.folded)
	}
```

This is the fold-shadowing rule: during an active filter, fold
state is preserved on `m.folded` but not applied to the rendered
rows. When the filter clears, the next `rebuild` (triggered by
`ClearFilter`) applies `m.folded` again, restoring the pre-search
fold layout.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestMessageListFilterFoldShadow -v`
Expected: PASS.

- [ ] **Step 5: Run the full msglist test suite**

Run: `go test ./internal/ui/ -run TestMessageList -v`
Expected: all tests still PASS — fold state behavior outside of
filter mode is unchanged.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/msglist.go internal/ui/msglist_test.go
git commit -m "MessageList: shadow fold state during active filter

When filter.query != \"\", the rebuild pipeline skips applyFoldState
so all matching threads render fully expanded. Saved fold state on
m.folded is preserved (not mutated) so ClearFilter restores the
pre-search fold layout via the next rebuild."
```

---

### Task 7: FilterResultCount accessor

**Files:**
- Modify: `internal/ui/msglist.go`
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Write failing tests for `FilterResultCount`**

Append to `internal/ui/msglist_test.go`:

```go
func TestMessageListFilterResultCount(t *testing.T) {
	styles := NewStyles(theme.Nord)

	msgs := []mail.MessageInfo{
		{UID: "1", ThreadID: "1", Subject: "Project alpha", From: "Alice", Date: "Apr 10"},
		{UID: "2", ThreadID: "2", Subject: "Project beta", From: "Bob", Date: "Apr 09"},
		{UID: "3", ThreadID: "3", Subject: "Weekend", From: "Carol", Date: "Apr 08"},
		// Thread T4 with 3 matching replies
		{UID: "10", ThreadID: "T4", InReplyTo: "", Subject: "Project gamma", From: "Dave", Date: "Apr 05"},
		{UID: "11", ThreadID: "T4", InReplyTo: "10", Subject: "Re: Project gamma", From: "Eve", Date: "Apr 06"},
		{UID: "12", ThreadID: "T4", InReplyTo: "11", Subject: "Re: Project gamma", From: "Frank", Date: "Apr 07"},
	}

	t.Run("count is thread count, not message count", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("project", SearchModeName)
		if got := ml.FilterResultCount(); got != 3 {
			t.Errorf("FilterResultCount = %d, want 3 (2 singletons + 1 thread)", got)
		}
	})

	t.Run("zero when no matches", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("zzz-nothing", SearchModeName)
		if got := ml.FilterResultCount(); got != 0 {
			t.Errorf("FilterResultCount = %d, want 0", got)
		}
	})

	t.Run("zero when no filter active", func(t *testing.T) {
		ml := NewMessageList(styles, msgs, 90, 20)
		if got := ml.FilterResultCount(); got != 0 {
			t.Errorf("FilterResultCount with no filter = %d, want 0", got)
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestMessageListFilterResultCount -v`
Expected: compile error — `FilterResultCount` is undefined.

- [ ] **Step 3: Add `FilterResultCount` and a per-filter thread count field**

Edit `internal/ui/msglist.go`. In the `MessageList` struct (the one
you added fields to in Task 3), add one more field after
`savedByFilter`:

```go
	filterResults   int // thread count from the most recent filter application
```

- [ ] **Step 4: Update `rebuild` to track the filter result count**

In `internal/ui/msglist.go`, update `rebuild` so it records the
number of buckets that survived the filter:

```go
func (m *MessageList) rebuild() {
	buckets := bucketByThreadID(m.source)
	buckets = m.filterBuckets(buckets)
	if m.filter.query != "" {
		m.filterResults = len(buckets)
	} else {
		m.filterResults = 0
	}
	sort.SliceStable(buckets, func(i, j int) bool {
		if m.sort == SortDateAsc {
			return latestActivity(buckets[i]) < latestActivity(buckets[j])
		}
		return latestActivity(buckets[i]) > latestActivity(buckets[j])
	})

	rows := make([]displayRow, 0, len(m.source))
	for _, bucket := range buckets {
		rows = appendThreadRows(rows, bucket)
	}
	if m.filter.query == "" {
		applyFoldState(rows, m.folded)
	}
	m.rows = rows
}
```

- [ ] **Step 5: Add the `FilterResultCount` accessor**

In `internal/ui/msglist.go`, add after the `Count` method (around
line 424):

```go
// FilterResultCount returns the number of threads matching the
// active filter, or 0 if no filter is active. Thread count — not
// message count — because the filter predicate runs per bucket and
// keeps whole threads as units.
func (m MessageList) FilterResultCount() int {
	return m.filterResults
}
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `go test ./internal/ui/ -run TestMessageListFilterResultCount -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/msglist.go internal/ui/msglist_test.go
git commit -m "MessageList: add FilterResultCount accessor

Returns thread count from the most recent filter rebuild — not
message count. A thread with 4 matching replies counts as 1.
Status bar and sidebar shelf consume this value downstream."
```

---

### Task 8: "No matches" placeholder under filter

**Files:**
- Modify: `internal/ui/msglist.go`
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 1: Write failing tests for placeholder wording**

Append to `internal/ui/msglist_test.go`:

```go
func TestMessageListPlaceholder(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("empty source shows No messages", func(t *testing.T) {
		ml := NewMessageList(styles, nil, 90, 20)
		plain := stripANSI(ml.View())
		if !strings.Contains(plain, "No messages") {
			t.Error("empty source should render 'No messages'")
		}
		if strings.Contains(plain, "No matches") {
			t.Error("empty source should not render 'No matches'")
		}
	})

	t.Run("filter with no matches shows No matches", func(t *testing.T) {
		msgs := []mail.MessageInfo{
			{UID: "1", ThreadID: "1", Subject: "Hello", From: "Alice", Date: "Apr 10"},
		}
		ml := NewMessageList(styles, msgs, 90, 20)
		ml.SetFilter("nothing-here-zzz", SearchModeName)
		plain := stripANSI(ml.View())
		if !strings.Contains(plain, "No matches") {
			t.Error("filter with no matches should render 'No matches'")
		}
		if strings.Contains(plain, "No messages") {
			t.Error("filter with no matches should not render 'No messages'")
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run TestMessageListPlaceholder -v`
Expected: FAIL — the second subtest renders "No messages" (current
`renderEmpty` always uses that text).

- [ ] **Step 3: Update `renderEmpty` to pick the wording based on filter state**

Edit `internal/ui/msglist.go`. Replace `renderEmpty` (around line
638) with:

```go
// renderEmpty renders the centered placeholder. Wording depends on
// why the list is empty: "No messages" when the source has no
// messages at all, "No matches" when a filter is active and matched
// nothing.
func (m MessageList) renderEmpty() string {
	label := "No messages"
	if m.filter.query != "" {
		label = "No matches"
	}
	labelLine := m.styles.MsgListBg.Width(m.width).
		Foreground(m.styles.MsgListPlaceholder.GetForeground()).
		Align(lipgloss.Center).
		Render(label)

	mid := m.height / 2
	lines := make([]string, m.height)
	for i := range lines {
		if i == mid {
			lines[i] = labelLine
		} else {
			lines[i] = m.renderBlankLine()
		}
	}
	return strings.Join(lines, "\n")
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestMessageListPlaceholder -v`
Expected: PASS.

- [ ] **Step 5: Run the full package tests to catch regressions**

Run: `go test ./internal/ui/ -v`
Expected: all tests PASS. Any existing test asserting "No messages"
on an empty source still passes because that path is unchanged.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/msglist.go internal/ui/msglist_test.go
git commit -m "MessageList: render 'No matches' placeholder under active filter

Distinguishes empty folder from empty filter result. Uses the new
MsgListPlaceholder style slot for foreground color. Existing empty-
source behavior (renders 'No messages') is unchanged."
```

---

## Phase 3 — SidebarSearch component

### Task 9: Scaffold SidebarSearch with idle rendering

**Files:**
- Create: `internal/ui/sidebar_search.go`
- Create: `internal/ui/sidebar_search_test.go`

- [ ] **Step 1: Create the test file with idle-state expectations**

Create `internal/ui/sidebar_search_test.go`:

```go
package ui

import (
	"strings"
	"testing"

	"github.com/glw907/poplar/internal/theme"
)

func TestSidebarSearchIdle(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("idle state shows hint row", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "/ to search") {
			t.Errorf("idle view missing hint: %q", plain)
		}
	})

	t.Run("idle state reports SearchIdle", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		if s.State() != SearchIdle {
			t.Errorf("State() = %v, want SearchIdle", s.State())
		}
	})

	t.Run("idle renders exactly 3 rows", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		lines := strings.Split(s.View(), "\n")
		if len(lines) != 3 {
			t.Errorf("idle view rows = %d, want 3", len(lines))
		}
	})

	t.Run("idle Query is empty", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		if s.Query() != "" {
			t.Errorf("Query() = %q, want empty", s.Query())
		}
	})

	t.Run("idle Mode is SearchModeName", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		if s.Mode() != SearchModeName {
			t.Errorf("Mode() = %v, want SearchModeName", s.Mode())
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestSidebarSearchIdle -v`
Expected: compile error — `SidebarSearch`, `NewSidebarSearch`,
`State`, `Query`, `Mode` are undefined.

- [ ] **Step 3: Create `sidebar_search.go` with the minimal idle-only implementation**

Create `internal/ui/sidebar_search.go` with exactly this content —
a minimal type and idle-only rendering. Typing/Active rendering,
mode cycle, and `Update` routing all land in Tasks 10-12.

```go
package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
)

// SidebarSearch is the 3-row shelf pinned to the bottom of the
// sidebar column. Owns the text input, mode toggle, and state
// machine for the search feature. Communicates with AccountTab via
// SearchUpdatedMsg during Typing; state transitions (Activate,
// Commit, Clear) are driven by direct method calls from AccountTab.
type SidebarSearch struct {
	input   textinput.Model
	mode    SearchMode
	state   SearchState
	results int
	styles  Styles
	width   int
}

// NewSidebarSearch constructs an idle search shelf at the given
// width. The textinput is created but not focused.
func NewSidebarSearch(styles Styles, width int) SidebarSearch {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 0
	return SidebarSearch{
		input:  ti,
		mode:   SearchModeName,
		state:  SearchIdle,
		styles: styles,
		width:  width,
	}
}

// State returns the current search state.
func (s SidebarSearch) State() SearchState { return s.state }

// Query returns the current textinput value. Empty when idle.
func (s SidebarSearch) Query() string { return s.input.Value() }

// Mode returns the current match mode.
func (s SidebarSearch) Mode() SearchMode { return s.mode }

// SetSize updates the shelf's width. Height is fixed at 3 rows.
func (s *SidebarSearch) SetSize(width int) {
	s.width = width
}

// View renders the shelf's 3 rows: blank separator, prompt/hint,
// mode/count row.
func (s SidebarSearch) View() string {
	if s.width <= 0 {
		return ""
	}
	return strings.Join([]string{
		s.renderBlankRow(),
		s.renderPromptRow(),
		s.renderInfoRow(),
	}, "\n")
}

// renderBlankRow renders a full-width blank row using the sidebar
// background.
func (s SidebarSearch) renderBlankRow() string {
	return s.styles.SidebarBg.Width(s.width).Render("")
}

// renderPromptRow renders the prompt line. Idle state only in this
// task — typing/active rendering lands in Task 10.
func (s SidebarSearch) renderPromptRow() string {
	if s.state != SearchIdle {
		return s.renderBlankRow()
	}
	icon := applyBg(s.styles.SearchIcon, s.styles.SidebarBg).Render("󰍉")
	hint := applyBg(s.styles.SearchHint, s.styles.SidebarBg).Render(" / to search")
	content := s.styles.SidebarBg.Render(" ") + icon + hint
	return fillRowToWidth(content, s.width, s.styles.SidebarBg)
}

// renderInfoRow renders the mode badge and result count. Blank in
// idle state — the active-state info row lands in Task 10.
func (s SidebarSearch) renderInfoRow() string {
	return s.renderBlankRow()
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestSidebarSearchIdle -v`
Expected: PASS.

- [ ] **Step 5: Run a full-package build**

Run: `go build ./...`
Expected: exits 0.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/sidebar_search.go internal/ui/sidebar_search_test.go
git commit -m "ui: scaffold SidebarSearch with idle rendering

Initial SidebarSearch subcomponent for Pass 2.5b-7. Idle state
renders a 3-row shelf: blank separator, '/ to search' hint row, and
a blank info row. Typing/Active states are stubbed; they land in
follow-up tasks."
```

---

### Task 10: Activate/Clear transitions + Typing render

**Files:**
- Modify: `internal/ui/sidebar_search.go`
- Modify: `internal/ui/sidebar_search_test.go`

- [ ] **Step 1: Write failing tests for Activate/Clear + typing**

Append to `internal/ui/sidebar_search_test.go`:

```go
func TestSidebarSearchActivate(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("Activate transitions Idle → Typing and focuses input", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		if s.State() != SearchTyping {
			t.Errorf("State() = %v, want SearchTyping", s.State())
		}
		if !s.input.Focused() {
			t.Error("input should be focused after Activate")
		}
	})

	t.Run("Clear returns to Idle and resets query", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("hello")
		s.Clear()
		if s.State() != SearchIdle {
			t.Errorf("State() = %v, want SearchIdle", s.State())
		}
		if s.Query() != "" {
			t.Errorf("Query() = %q, want empty", s.Query())
		}
		if s.input.Focused() {
			t.Error("input should be blurred after Clear")
		}
	})

	t.Run("Clear also resets mode to SearchModeName", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.mode = SearchModeAll
		s.Clear()
		if s.Mode() != SearchModeName {
			t.Errorf("Mode() after Clear = %v, want SearchModeName", s.Mode())
		}
	})

	t.Run("typing state renders icon + slash + query", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("proj")
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "󰍉") {
			t.Errorf("typing view missing search icon: %q", plain)
		}
		if !strings.Contains(plain, "/proj") {
			t.Errorf("typing view missing '/proj' prompt: %q", plain)
		}
	})

	t.Run("typing state renders [name] mode badge", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "[name]") {
			t.Errorf("typing view missing [name] badge: %q", plain)
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run TestSidebarSearchActivate -v`
Expected: compile error — `Activate`, `Clear` are undefined.

- [ ] **Step 3: Add `Activate` and `Clear` methods**

Edit `internal/ui/sidebar_search.go`. Add after the `SetSize` method:

```go
// Activate transitions Idle → Typing and focuses the text input.
// Safe to call from any state: re-activates an Active shelf into
// Typing without losing the query.
func (s *SidebarSearch) Activate() {
	s.state = SearchTyping
	s.input.Focus()
}

// Clear returns the shelf to Idle, empties the query, blurs the
// input, and resets the mode to SearchModeName.
func (s *SidebarSearch) Clear() {
	s.state = SearchIdle
	s.input.Reset()
	s.input.Blur()
	s.mode = SearchModeName
	s.results = 0
}
```

- [ ] **Step 4: Configure the textinput to use "/" as its prompt**

Edit `NewSidebarSearch` in `internal/ui/sidebar_search.go`.
Replace the `ti.Prompt = ""` line with `ti.Prompt = "/"`. Full
updated constructor:

```go
// NewSidebarSearch constructs an idle search shelf at the given
// width. The textinput is created with "/" as its prompt so the
// rendered view shows "/query▏" directly without our shelf having
// to stitch a prefix in front of it.
func NewSidebarSearch(styles Styles, width int) SidebarSearch {
	ti := textinput.New()
	ti.Prompt = "/"
	ti.CharLimit = 0
	return SidebarSearch{
		input:  ti,
		mode:   SearchModeName,
		state:  SearchIdle,
		styles: styles,
		width:  width,
	}
}
```

- [ ] **Step 5: Update `renderPromptRow` to render Typing and Active states**

Replace `renderPromptRow` in `internal/ui/sidebar_search.go` with:

```go
// renderPromptRow renders the prompt line.
//   - Idle: shows "󰍉 / to search" hint in dim color.
//   - Typing: shows "󰍉" + textinput.View() which renders "/query▏"
//     (cursor ▏ drawn automatically because the input is Focused).
//   - Active: shows "󰍉" + a manually-rendered "/query" with a
//     brighter foreground to signal "committed query." No cursor
//     because the input is Blurred.
func (s SidebarSearch) renderPromptRow() string {
	if s.state == SearchIdle {
		icon := applyBg(s.styles.SearchIcon, s.styles.SidebarBg).Render("󰍉")
		hint := applyBg(s.styles.SearchHint, s.styles.SidebarBg).Render(" / to search")
		content := s.styles.SidebarBg.Render(" ") + icon + hint
		return fillRowToWidth(content, s.width, s.styles.SidebarBg)
	}

	// Tint the icon with the accent while typing; keep the dim color
	// once committed.
	iconStyle := s.styles.SearchIcon
	if s.state == SearchTyping {
		iconStyle = iconStyle.Foreground(s.styles.SearchResultCount.GetForeground())
	}
	icon := applyBg(iconStyle, s.styles.SidebarBg).Render("󰍉")

	var prompt string
	if s.state == SearchTyping {
		// textinput.View renders "/" (our configured prompt) + the
		// query + the cursor, all in its own styling. We wrap it
		// with the sidebar background via applyBg.
		prompt = applyBg(s.styles.SearchPrompt, s.styles.SidebarBg).Render(s.input.View())
	} else {
		// Active: manual render so we can use a brighter foreground.
		// SidebarAccount uses fg_bright bold — the right weight for
		// a locked query.
		text := "/" + s.input.Value()
		prompt = applyBg(s.styles.SidebarAccount, s.styles.SidebarBg).Render(text)
	}

	content := s.styles.SidebarBg.Render(" ") + icon + s.styles.SidebarBg.Render(" ") + prompt
	return fillRowToWidth(content, s.width, s.styles.SidebarBg)
}
```

- [ ] **Step 6: Update `renderInfoRow` to show the mode badge**

Replace `renderInfoRow` in `internal/ui/sidebar_search.go` with:

```go
// renderInfoRow renders the mode badge and result count. Blank in
// idle state; in typing/active renders "[name]" or "[all]" on the
// left and the result count or "no results" on the right.
func (s SidebarSearch) renderInfoRow() string {
	if s.state == SearchIdle {
		return s.renderBlankRow()
	}
	modeLabel := "[name]"
	if s.mode == SearchModeAll {
		modeLabel = "[all]"
	}
	mode := applyBg(s.styles.SearchModeBadge, s.styles.SidebarBg).Render(modeLabel)

	var countText string
	var countStyled string
	if s.Query() != "" && s.results == 0 {
		countText = "no results"
		countStyled = applyBg(s.styles.SearchNoResults, s.styles.SidebarBg).Render(countText)
	} else {
		countText = formatResultCount(s.results)
		countStyled = applyBg(s.styles.SearchResultCount, s.styles.SidebarBg).Render(countText)
	}

	indent := s.styles.SidebarBg.Render("  ")
	margin := s.styles.SidebarBg.Render(" ")
	contentCells := 2 + runewidth.StringWidth(modeLabel) + runewidth.StringWidth(countText) + 1
	gap := max(1, s.width-contentCells)
	content := indent + mode + s.styles.SidebarBg.Render(strings.Repeat(" ", gap)) + countStyled + margin
	return fillRowToWidth(content, s.width, s.styles.SidebarBg)
}

// formatResultCount returns the visible text for a result count.
func formatResultCount(n int) string {
	if n == 1 {
		return "1 result"
	}
	return itoa(n) + " results"
}

// itoa stringifies a non-negative int without allocation beyond the
// returned string. Used in render paths that fire on every keystroke.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
```

- [ ] **Step 7: Add `runewidth` to the imports**

Edit the import block at the top of `internal/ui/sidebar_search.go`:

```go
import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/mattn/go-runewidth"
)
```

- [ ] **Step 8: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestSidebarSearchActivate -v`
Expected: PASS.

- [ ] **Step 9: Run the full suite for regressions**

Run: `go test ./internal/ui/ -v`
Expected: all PASS.

- [ ] **Step 10: Commit**

```bash
git add internal/ui/sidebar_search.go internal/ui/sidebar_search_test.go
git commit -m "SidebarSearch: Activate/Clear transitions and typing render

Activate enters Typing and focuses the input. Clear resets to Idle,
empties the query, blurs the input, and resets the mode. The prompt
row and info row now render the typing state with '/query', a mode
badge, and a result count."
```

---

### Task 11: Commit transition and Update plumbing

**Files:**
- Modify: `internal/ui/sidebar_search.go`
- Modify: `internal/ui/sidebar_search_test.go`

- [ ] **Step 1: Write failing tests for Commit and Update**

Append to `internal/ui/sidebar_search_test.go`:

```go
func TestSidebarSearchCommit(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("Commit transitions Typing → Active", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("hello")
		s.Commit()
		if s.State() != SearchActive {
			t.Errorf("State() = %v, want SearchActive", s.State())
		}
		if s.Query() != "hello" {
			t.Errorf("Query() preserved = %q, want 'hello'", s.Query())
		}
		if s.input.Focused() {
			t.Error("input should be blurred in Active state")
		}
	})

	t.Run("re-Activate from Active preserves query", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("hello")
		s.Commit()
		s.Activate() // re-enter Typing
		if s.State() != SearchTyping {
			t.Errorf("State() = %v, want SearchTyping", s.State())
		}
		if s.Query() != "hello" {
			t.Errorf("Query() preserved = %q, want 'hello'", s.Query())
		}
		if !s.input.Focused() {
			t.Error("input should be focused after re-Activate")
		}
	})
}

func TestSidebarSearchUpdate(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("printable rune during typing appends to query", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
		if s.Query() != "pro" {
			t.Errorf("Query() = %q, want 'pro'", s.Query())
		}
	})

	t.Run("Update emits SearchUpdatedMsg on keystroke", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
		if cmd == nil {
			t.Fatal("Update should return a Cmd emitting SearchUpdatedMsg")
		}
		msg := cmd()
		upd, ok := msg.(SearchUpdatedMsg)
		if !ok {
			t.Fatalf("Cmd returned %T, want SearchUpdatedMsg", msg)
		}
		if upd.Query != "p" {
			t.Errorf("SearchUpdatedMsg.Query = %q, want 'p'", upd.Query)
		}
		if upd.Mode != SearchModeName {
			t.Errorf("SearchUpdatedMsg.Mode = %v, want SearchModeName", upd.Mode)
		}
	})

	t.Run("Backspace during typing emits SearchUpdatedMsg with shorter query", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("proj")
		_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		if s.Query() != "pro" {
			t.Errorf("Query() after backspace = %q, want 'pro'", s.Query())
		}
		msg := cmd()
		upd, ok := msg.(SearchUpdatedMsg)
		if !ok || upd.Query != "pro" {
			t.Errorf("expected SearchUpdatedMsg{Query: 'pro'}, got %v", msg)
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run 'TestSidebarSearchCommit|TestSidebarSearchUpdate' -v`
Expected: compile error — `Commit`, `Update` are undefined, and
the test file references `tea.KeyMsg` which it needs to import.

- [ ] **Step 3: Add tea import to the test file**

Edit the top of `internal/ui/sidebar_search_test.go`:

```go
import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glw907/poplar/internal/theme"
)
```

- [ ] **Step 4: Add `Commit` and `Update` methods**

Edit `internal/ui/sidebar_search.go`. Add after `Clear`:

```go
// Commit transitions Typing → Active, leaving the query intact and
// blurring the input. Safe to call from Active (no-op).
func (s *SidebarSearch) Commit() {
	s.state = SearchActive
	s.input.Blur()
}

// Update routes a bubbletea Msg through the textinput and returns
// the possibly-mutated shelf plus a Cmd that emits a
// SearchUpdatedMsg whenever the query changed. Only meaningful in
// SearchTyping state — callers in other states should not route
// keys here.
func (s SidebarSearch) Update(msg tea.Msg) (SidebarSearch, tea.Cmd) {
	if s.state != SearchTyping {
		return s, nil
	}
	prev := s.input.Value()
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	cur := s.input.Value()
	if cur == prev {
		return s, cmd
	}
	query := cur
	mode := s.mode
	emitCmd := func() tea.Msg {
		return SearchUpdatedMsg{Query: query, Mode: mode}
	}
	if cmd == nil {
		return s, emitCmd
	}
	return s, tea.Batch(cmd, emitCmd)
}
```

- [ ] **Step 5: Add the missing `tea` import to `sidebar_search.go`**

Edit the import block at the top of `internal/ui/sidebar_search.go`:

```go
import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)
```

- [ ] **Step 6: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run 'TestSidebarSearchCommit|TestSidebarSearchUpdate' -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/sidebar_search.go internal/ui/sidebar_search_test.go
git commit -m "SidebarSearch: Commit transition and Update routing

Commit moves Typing → Active without losing the query. Update
routes keystrokes through the embedded textinput.Model and emits
SearchUpdatedMsg when the query value changed."
```

---

### Task 12: Tab mode cycle and SetResultCount

**Files:**
- Modify: `internal/ui/sidebar_search.go`
- Modify: `internal/ui/sidebar_search_test.go`

- [ ] **Step 1: Write failing tests for Tab cycle and SetResultCount**

Append to `internal/ui/sidebar_search_test.go`:

```go
func TestSidebarSearchModeCycle(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("Tab cycles mode [name] → [all] → [name]", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()

		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyTab})
		if s.Mode() != SearchModeAll {
			t.Errorf("after first Tab: Mode = %v, want SearchModeAll", s.Mode())
		}

		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyTab})
		if s.Mode() != SearchModeName {
			t.Errorf("after second Tab: Mode = %v, want SearchModeName", s.Mode())
		}
	})

	t.Run("Tab cycle emits SearchUpdatedMsg with new mode", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("proj")

		_, cmd := s.Update(tea.KeyMsg{Type: tea.KeyTab})
		if cmd == nil {
			t.Fatal("Tab should emit a Cmd")
		}
		msg := cmd()
		upd, ok := msg.(SearchUpdatedMsg)
		if !ok {
			t.Fatalf("Cmd returned %T, want SearchUpdatedMsg", msg)
		}
		if upd.Mode != SearchModeAll {
			t.Errorf("SearchUpdatedMsg.Mode = %v, want SearchModeAll", upd.Mode)
		}
		if upd.Query != "proj" {
			t.Errorf("SearchUpdatedMsg.Query = %q, want 'proj'", upd.Query)
		}
	})

	t.Run("view shows [all] after Tab", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyTab})
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "[all]") {
			t.Errorf("view missing [all] badge after Tab: %q", plain)
		}
	})
}

func TestSidebarSearchResultCount(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("SetResultCount stores the value", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("proj")
		s.SetResultCount(3)
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "3 results") {
			t.Errorf("view missing '3 results': %q", plain)
		}
	})

	t.Run("zero results with non-empty query shows 'no results'", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("asdf")
		s.SetResultCount(0)
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "no results") {
			t.Errorf("view missing 'no results': %q", plain)
		}
	})

	t.Run("singular '1 result' for count 1", func(t *testing.T) {
		s := NewSidebarSearch(styles, 30)
		s.Activate()
		s.input.SetValue("proj")
		s.SetResultCount(1)
		plain := stripANSI(s.View())
		if !strings.Contains(plain, "1 result") {
			t.Errorf("view missing '1 result': %q", plain)
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run 'TestSidebarSearchModeCycle|TestSidebarSearchResultCount' -v`
Expected: FAIL — Tab doesn't cycle mode, and `SetResultCount` is
undefined.

- [ ] **Step 3: Add `SetResultCount` method**

Edit `internal/ui/sidebar_search.go`. Add after `Commit`:

```go
// SetResultCount stores the most recent filter result count (thread
// count) for display in the info row. Called by AccountTab when it
// receives a SearchResultsMsg.
func (s *SidebarSearch) SetResultCount(n int) {
	s.results = n
}
```

- [ ] **Step 4: Intercept Tab in Update to cycle the mode**

Edit `Update` in `internal/ui/sidebar_search.go`. Replace it with:

```go
// Update routes a bubbletea Msg through the textinput and returns
// the possibly-mutated shelf plus a Cmd that emits a
// SearchUpdatedMsg whenever the query or mode changed. Only
// meaningful in SearchTyping state.
func (s SidebarSearch) Update(msg tea.Msg) (SidebarSearch, tea.Cmd) {
	if s.state != SearchTyping {
		return s, nil
	}

	// Intercept Tab: cycle the mode without routing to textinput.
	if key, ok := msg.(tea.KeyMsg); ok && key.Type == tea.KeyTab {
		if s.mode == SearchModeName {
			s.mode = SearchModeAll
		} else {
			s.mode = SearchModeName
		}
		query := s.input.Value()
		mode := s.mode
		return s, func() tea.Msg {
			return SearchUpdatedMsg{Query: query, Mode: mode}
		}
	}

	prev := s.input.Value()
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	cur := s.input.Value()
	if cur == prev {
		return s, cmd
	}
	query := cur
	mode := s.mode
	emitCmd := func() tea.Msg {
		return SearchUpdatedMsg{Query: query, Mode: mode}
	}
	if cmd == nil {
		return s, emitCmd
	}
	return s, tea.Batch(cmd, emitCmd)
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run 'TestSidebarSearchModeCycle|TestSidebarSearchResultCount' -v`
Expected: PASS.

- [ ] **Step 6: Run the full package tests**

Run: `go test ./internal/ui/ -v`
Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/sidebar_search.go internal/ui/sidebar_search_test.go
git commit -m "SidebarSearch: Tab cycles mode and SetResultCount stores count

Tab is intercepted in Update before reaching textinput so it cycles
the mode between [name] and [all] and emits SearchUpdatedMsg with
the new mode. SetResultCount lets AccountTab feed back the count
for display in the info row."
```

---

## Phase 4 — AccountTab integration

### Task 13: Layout — searchShelfRows constant, height math, View composition

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Write a failing test for the new layout**

Append to `internal/ui/account_tab_test.go`:

```go
func TestAccountTabSearchShelf(t *testing.T) {
	t.Run("view renders the search hint at the bottom of the sidebar", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		view := stripANSI(tab.View())
		if !strings.Contains(view, "/ to search") {
			t.Error("sidebar should show '/ to search' hint")
		}
	})

	t.Run("search hint is in the last few rows of the sidebar column", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		lines := strings.Split(stripANSI(tab.View()), "\n")
		// Sidebar column is the left portion of each line. Find the
		// row that contains the hint.
		hintRow := -1
		for i, line := range lines {
			if strings.Contains(line, "/ to search") {
				hintRow = i
				break
			}
		}
		if hintRow < 0 {
			t.Fatal("hint not found in view")
		}
		// Hint row should be within the bottom 3 rows of content (the
		// 3-row shelf), allowing for the blank separator and info row.
		contentRows := len(lines)
		if hintRow < contentRows-3 || hintRow >= contentRows {
			t.Errorf("hint row %d not in bottom shelf (content rows: %d)", hintRow, contentRows)
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestAccountTabSearchShelf -v`
Expected: FAIL — the hint is not rendered anywhere yet.

- [ ] **Step 3: Add the `sidebarSearch` field and `searchShelfRows` constant**

Edit `internal/ui/account_tab.go`. Replace the constant block (lines
12-19) with:

```go
// sidebarWidth is the fixed width of the sidebar panel.
const sidebarWidth = 30

// sidebarHeaderRows is the account-name line plus the blank line
// below it, reserved at the top of the sidebar before the folder
// list. AccountTab.View and the sidebar's own sizing both depend on
// this number matching.
const sidebarHeaderRows = 2

// searchShelfRows is the height of the SidebarSearch shelf pinned
// to the bottom of the sidebar column.
const searchShelfRows = 3
```

Add `sidebarSearch` to the `AccountTab` struct. Replace the struct
(lines 22-36) with:

```go
// AccountTab is the main account view. One pane (like pine): every
// key is always live. J/K/G navigate folders, j/k navigate messages.
type AccountTab struct {
	styles Styles
	// backend is held as a read-only reference so Update can build
	// tea.Cmd closures that call backend methods. It is never
	// mutated and its results are never cached as owned state —
	// they come back as Msg types through the normal Update flow.
	// This is the elm-conventions Rule 5 exception.
	backend       mail.Backend
	uiCfg         config.UIConfig
	sidebar       Sidebar
	sidebarSearch SidebarSearch
	msglist       MessageList
	width         int
	height        int
}
```

- [ ] **Step 4: Initialize `sidebarSearch` in `NewAccountTab`**

Replace `NewAccountTab` (lines 40-48) with:

```go
// NewAccountTab builds an empty AccountTab. The initial folder list is
// fetched via Init's returned Cmd, not synchronously.
func NewAccountTab(styles Styles, backend mail.Backend, uiCfg config.UIConfig) AccountTab {
	return AccountTab{
		styles:        styles,
		backend:       backend,
		uiCfg:         uiCfg,
		sidebar:       NewSidebar(styles, nil, uiCfg, sidebarWidth, 1),
		sidebarSearch: NewSidebarSearch(styles, sidebarWidth),
		msglist:       NewMessageList(styles, nil, 1, 1),
	}
}
```

- [ ] **Step 5: Update the `WindowSizeMsg` handler to size the shelf and the new folder region**

In `updateTab`, replace the `tea.WindowSizeMsg` case (lines 72-79):

```go
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		sw := min(sidebarWidth, m.width/2)
		folderHeight := max(1, m.height-sidebarHeaderRows-searchShelfRows)
		m.sidebar.SetSize(sw, folderHeight)
		m.sidebarSearch.SetSize(sw)
		mw := max(1, m.width-sw-1) // -1 for divider
		m.msglist.SetSize(mw, m.height)
		return m, nil
```

- [ ] **Step 6: Update `View` to compose account header + folder region + shelf**

Replace `View` in `internal/ui/account_tab.go` (around line 157)
with:

```go
// View renders the sidebar + divider + message list. The sidebar
// column is composed top-to-bottom as: account header (2 rows),
// folder region (flex), search shelf (3 rows pinned to bottom).
func (m AccountTab) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	sw := min(sidebarWidth, m.width/2)

	acctLine := m.styles.SidebarAccount.Width(sw).Render(" " + m.backend.AccountName())
	blank := m.styles.SidebarBg.Width(sw).Render("")

	sidebarFolders := m.sidebar.View()
	shelfView := m.sidebarSearch.View()

	var sidebarLines []string
	sidebarLines = append(sidebarLines, acctLine, blank)
	if sidebarFolders != "" {
		sidebarLines = append(sidebarLines, strings.Split(sidebarFolders, "\n")...)
	}
	// Pad the folder region with blank rows so the shelf lands at
	// the bottom of the column regardless of how many folders exist.
	targetFolderEnd := m.height - searchShelfRows
	for len(sidebarLines) < targetFolderEnd {
		sidebarLines = append(sidebarLines, blank)
	}
	if len(sidebarLines) > targetFolderEnd {
		sidebarLines = sidebarLines[:targetFolderEnd]
	}
	sidebarLines = append(sidebarLines, strings.Split(shelfView, "\n")...)
	if len(sidebarLines) > m.height {
		sidebarLines = sidebarLines[:m.height]
	}

	sidebarView := strings.Join(sidebarLines, "\n")
	divider := renderDivider(m.height, m.styles)
	msglistView := m.msglist.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, divider, msglistView)
}
```

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestAccountTabSearchShelf -v`
Expected: PASS.

- [ ] **Step 8: Run the full package tests**

Run: `go test ./internal/ui/ -v`
Expected: all PASS. Existing `TestAccountTab` tests still pass
because the account header + divider + folders rendering is
unchanged.

- [ ] **Step 9: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "AccountTab: embed SidebarSearch shelf pinned at column bottom

New sidebarSearch field initialized in NewAccountTab; sized in the
WindowSizeMsg handler. View composes the sidebar column as header
(2) + folder region (flex, padded to push content up) + shelf (3,
pinned to bottom). Folder region height now subtracts searchShelfRows
from the total."
```

---

### Task 14: Route `/` in Idle state to activate search

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Write a failing test for `/` activation**

Append to `internal/ui/account_tab_test.go`:

```go
func TestAccountTabSearchActivation(t *testing.T) {
	t.Run("/ in Idle activates search", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		if tab.sidebarSearch.State() != SearchTyping {
			t.Errorf("state after / = %v, want SearchTyping", tab.sidebarSearch.State())
		}
	})

	t.Run("/ in Idle does not start filtering yet (empty query)", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		rowCountBefore := len(tab.msglist.rows)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		if got := len(tab.msglist.rows); got != rowCountBefore {
			t.Errorf("row count after / = %d, want %d (no filter yet)", got, rowCountBefore)
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestAccountTabSearchActivation -v`
Expected: FAIL — the `/` key currently hits no handler and the
sidebarSearch stays in SearchIdle.

- [ ] **Step 3: Route `/` in `handleKey` when search is Idle**

Edit `internal/ui/account_tab.go`. Replace `handleKey` (around line
107) with:

```go
// handleKey dispatches navigation keys by identity. J/K/G move the
// sidebar (and dispatch a folder-load Cmd); j/k/Ctrl-d/Ctrl-u move the
// message list cursor. During an active search, printable keys flow
// through the SidebarSearch instead of the account-view handlers.
func (m AccountTab) handleKey(msg tea.KeyMsg) (AccountTab, tea.Cmd) {
	// Route to SidebarSearch when we're in Typing state — it owns
	// the input routing for this modal slice.
	if m.sidebarSearch.State() == SearchTyping {
		var cmd tea.Cmd
		m.sidebarSearch, cmd = m.sidebarSearch.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "/":
		if m.sidebarSearch.State() == SearchIdle {
			m.sidebarSearch.Activate()
			return m, nil
		}
	case "J":
		m.sidebar.MoveDown()
		return m, m.selectionChangedCmds()
	case "K":
		m.sidebar.MoveUp()
		return m, m.selectionChangedCmds()
	case "G":
		m.msglist.MoveToBottom()
	case "g":
		m.msglist.MoveToTop()
	case "j", "down":
		m.msglist.MoveDown()
	case "k", "up":
		m.msglist.MoveUp()
	case "ctrl+d":
		m.msglist.HalfPageDown()
	case "ctrl+u":
		m.msglist.HalfPageUp()
	case "ctrl+f", "pgdown":
		m.msglist.PageDown()
	case "ctrl+b", "pgup":
		m.msglist.PageUp()
	case " ":
		m.msglist.ToggleFold()
	case "F":
		m.msglist.FoldAll()
	case "U":
		m.msglist.UnfoldAll()
	}
	return m, nil
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestAccountTabSearchActivation -v`
Expected: PASS.

- [ ] **Step 5: Run the full package tests**

Run: `go test ./internal/ui/ -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "AccountTab: activate SidebarSearch on / keypress

Pressing / in Idle state calls SidebarSearch.Activate, which
transitions the shelf to Typing and focuses its textinput. All
subsequent keypresses are routed to the shelf via the Typing-state
early-return in handleKey."
```

---

### Task 15: Handle SearchUpdatedMsg → call MessageList.SetFilter

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Write a failing test for typing driving the filter**

Append to `internal/ui/account_tab_test.go`:

```go
func TestAccountTabSearchFilter(t *testing.T) {
	t.Run("typing during search filters the message list", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		// Inbox seeded by mock backend — the mock messages include
		// an "Alice" sender (see mail/mock.go). Activate search and
		// type 'a' 'l' 'i' to narrow to Alice.
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		rowsBefore := len(tab.msglist.rows)

		// Drive three keystrokes through the handler; each returns a
		// Cmd that emits SearchUpdatedMsg, which drives SetFilter.
		for _, r := range []rune{'a', 'l', 'i'} {
			tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
			drain(t, &tab, cmd)
			_ = tab
		}

		if got := len(tab.msglist.rows); got >= rowsBefore {
			t.Errorf("row count after typing = %d, want < %d", got, rowsBefore)
		}
	})

	t.Run("SearchUpdatedMsg directly sets the filter", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		rowsBefore := len(tab.msglist.rows)
		tab, _ = tab.updateTab(SearchUpdatedMsg{Query: "alice", Mode: SearchModeName})
		if got := len(tab.msglist.rows); got >= rowsBefore {
			t.Errorf("row count after SearchUpdatedMsg = %d, want < %d", got, rowsBefore)
		}
	})

	t.Run("SearchUpdatedMsg feeds the count back to SidebarSearch", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(SearchUpdatedMsg{Query: "alice", Mode: SearchModeName})
		if tab.sidebarSearch.results != tab.msglist.FilterResultCount() {
			t.Errorf("sidebarSearch.results = %d, want %d (mirrors FilterResultCount)",
				tab.sidebarSearch.results, tab.msglist.FilterResultCount())
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestAccountTabSearchFilter -v`
Expected: FAIL — `SearchUpdatedMsg` is not handled, so the filter
never runs.

- [ ] **Step 3: Handle `SearchUpdatedMsg` and `SearchResultsMsg` in `updateTab`**

Edit `internal/ui/account_tab.go`. In `updateTab`, add two cases
after `backendErrMsg` (around line 96):

```go
	case SearchUpdatedMsg:
		m.msglist.SetFilter(msg.Query, msg.Mode)
		m.sidebarSearch.SetResultCount(m.msglist.FilterResultCount())
		return m, nil
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestAccountTabSearchFilter -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "AccountTab: drive MessageList filter from SearchUpdatedMsg

SearchUpdatedMsg handler calls msglist.SetFilter, then pushes the
resulting thread count into SidebarSearch.SetResultCount so the
info row shows the current count. No intermediate Msg round-trip —
AccountTab owns both children and wires them directly."
```

---

### Task 16: Commit and clear handlers (Enter → Active, Esc → Idle)

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`

- [ ] **Step 1: Write failing tests for Enter and Esc routing**

Append to `internal/ui/account_tab_test.go`:

```go
func TestAccountTabSearchCommitClear(t *testing.T) {
	t.Run("Enter in Typing transitions shelf to Active", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		if tab.sidebarSearch.State() != SearchActive {
			t.Errorf("state after Enter = %v, want SearchActive", tab.sidebarSearch.State())
		}
	})

	t.Run("Enter keeps the filter live (query preserved)", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		filteredRows := len(tab.msglist.rows)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		if got := len(tab.msglist.rows); got != filteredRows {
			t.Errorf("row count after Enter = %d, want %d (filter preserved)", got, filteredRows)
		}
	})

	t.Run("Esc in Typing clears the filter and returns to Idle", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		rowsBefore := len(tab.msglist.rows)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, cmd = tab.updateTab(tea.KeyMsg{Type: tea.KeyEsc})
		drain(t, &tab, cmd)
		if tab.sidebarSearch.State() != SearchIdle {
			t.Errorf("state after Esc = %v, want SearchIdle", tab.sidebarSearch.State())
		}
		if got := len(tab.msglist.rows); got != rowsBefore {
			t.Errorf("row count after Esc = %d, want %d (full restore)", got, rowsBefore)
		}
	})

	t.Run("Esc in Active clears the filter", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		rowsBefore := len(tab.msglist.rows)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		// Now in Active state with a filter live.
		tab, cmd = tab.updateTab(tea.KeyMsg{Type: tea.KeyEsc})
		drain(t, &tab, cmd)
		if tab.sidebarSearch.State() != SearchIdle {
			t.Errorf("state after Esc in Active = %v, want SearchIdle", tab.sidebarSearch.State())
		}
		if got := len(tab.msglist.rows); got != rowsBefore {
			t.Errorf("row count after Esc in Active = %d, want %d", got, rowsBefore)
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run TestAccountTabSearchCommitClear -v`
Expected: FAIL — Enter/Esc are not yet wired to the shelf.

- [ ] **Step 3: Intercept Enter and Esc inside the Typing branch of handleKey**

Edit `internal/ui/account_tab.go`. Replace the Typing-state early
return at the top of `handleKey` (inserted in Task 14) with:

```go
	// Route to SidebarSearch when we're in Typing state — it owns
	// the input routing for this modal slice, except for Enter and
	// Esc which transition state.
	if m.sidebarSearch.State() == SearchTyping {
		switch msg.Type {
		case tea.KeyEnter:
			m.sidebarSearch.Commit()
			return m, nil
		case tea.KeyEsc:
			m.sidebarSearch.Clear()
			m.msglist.ClearFilter()
			return m, nil
		}
		var cmd tea.Cmd
		m.sidebarSearch, cmd = m.sidebarSearch.Update(msg)
		return m, cmd
	}
```

- [ ] **Step 4: Handle Esc in Active state (outside the Typing branch)**

Also in `handleKey`, add an Esc case inside the main switch (after
the `/` case you added in Task 14):

```go
	case "/":
		if m.sidebarSearch.State() == SearchIdle {
			m.sidebarSearch.Activate()
			return m, nil
		}
		if m.sidebarSearch.State() == SearchActive {
			m.sidebarSearch.Activate()
			return m, nil
		}
	case "esc":
		if m.sidebarSearch.State() == SearchActive {
			m.sidebarSearch.Clear()
			m.msglist.ClearFilter()
			return m, nil
		}
```

Note the double handling of `/`: the second branch re-activates
from Active (sends back to Typing with the query preserved), which
exercises the behavior we tested in Task 11 (`re-Activate from
Active preserves query`).

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestAccountTabSearchCommitClear -v`
Expected: PASS.

- [ ] **Step 6: Run the full package tests**

Run: `go test ./internal/ui/ -v`
Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "AccountTab: Enter commits search, Esc clears it

Enter in Typing state transitions the shelf to Active without
clearing the filter. Esc in either Typing or Active clears the
query, resets the shelf to Idle, and calls msglist.ClearFilter to
restore the pre-search cursor row. / in Active re-enters Typing
with the existing query preserved."
```

---

### Task 17: Folder jump clears search; q-stolen in Active; fold no-op in Active

**Files:**
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/account_tab_test.go`
- Modify: `internal/ui/app.go` (for the q-stolen path)

- [ ] **Step 1: Write failing tests for folder jump, q, and fold no-op**

Append to `internal/ui/account_tab_test.go`:

```go
func TestAccountTabSearchFolderJump(t *testing.T) {
	t.Run("J during Active clears the search", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		if tab.sidebarSearch.State() != SearchActive {
			t.Fatalf("setup: state = %v, want SearchActive", tab.sidebarSearch.State())
		}
		tab, cmd = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
		if tab.sidebarSearch.State() != SearchIdle {
			t.Errorf("state after J = %v, want SearchIdle", tab.sidebarSearch.State())
		}
		_ = cmd
	})

	t.Run("K during Active clears the search", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
		if tab.sidebarSearch.State() != SearchIdle {
			t.Errorf("state after K = %v, want SearchIdle", tab.sidebarSearch.State())
		}
	})
}

func TestAccountTabSearchFoldNoOp(t *testing.T) {
	t.Run("Space during Active does not crash and does not exit search", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
		if tab.sidebarSearch.State() != SearchActive {
			t.Errorf("state after Space = %v, want SearchActive", tab.sidebarSearch.State())
		}
	})

	t.Run("F during Active does not crash", func(t *testing.T) {
		tab := newLoadedTab(t, 80, 30)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		tab, cmd := tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		drain(t, &tab, cmd)
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyEnter})
		tab, _ = tab.updateTab(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'F'}})
		if tab.sidebarSearch.State() != SearchActive {
			t.Errorf("state after F = %v, want SearchActive", tab.sidebarSearch.State())
		}
	})
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/ui/ -run 'TestAccountTabSearchFolderJump|TestAccountTabSearchFoldNoOp' -v`
Expected: FAIL — J/K do not currently clear search; Space/F/U still
call ToggleFold/FoldAll/FoldAll on the msglist during Active.

- [ ] **Step 3: Add a `clearSearchIfActive` helper and wire folder jumps**

Edit `internal/ui/account_tab.go`. Add a small helper at the end of
the file (after `selectionChangedCmds`):

```go
// clearSearchIfActive clears the shelf and the filter if the shelf
// is in any non-Idle state. Returns true if anything was cleared —
// callers use this to decide whether to run follow-up logic.
func (m *AccountTab) clearSearchIfActive() bool {
	if m.sidebarSearch.State() == SearchIdle {
		return false
	}
	m.sidebarSearch.Clear()
	m.msglist.ClearFilter()
	return true
}
```

Then update `handleKey` to call it from `J` and `K`:

```go
	case "J":
		m.clearSearchIfActive()
		m.sidebar.MoveDown()
		return m, m.selectionChangedCmds()
	case "K":
		m.clearSearchIfActive()
		m.sidebar.MoveUp()
		return m, m.selectionChangedCmds()
```

- [ ] **Step 4: Gate fold keys on search-Idle state**

In `handleKey`, replace the `" "`, `"F"`, `"U"` cases with:

```go
	case " ":
		if m.sidebarSearch.State() == SearchActive {
			return m, nil
		}
		m.msglist.ToggleFold()
	case "F":
		if m.sidebarSearch.State() == SearchActive {
			return m, nil
		}
		m.msglist.FoldAll()
	case "U":
		if m.sidebarSearch.State() == SearchActive {
			return m, nil
		}
		m.msglist.UnfoldAll()
```

During Typing state these keys are already intercepted by the
Typing-branch early return (they reach the textinput instead), so
we only need to gate them for Active state here.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run 'TestAccountTabSearchFolderJump|TestAccountTabSearchFoldNoOp' -v`
Expected: PASS.

- [ ] **Step 6: Run the full package tests**

Run: `go test ./internal/ui/ -v`
Expected: all PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/ui/account_tab.go internal/ui/account_tab_test.go
git commit -m "AccountTab: folder jumps clear search; fold keys no-op during Active

J/K (folder nav) now call clearSearchIfActive before moving the
sidebar cursor, matching the design's 'folder jumps clear search'
rule from ADR 0064. Space/F/U are no-ops while SidebarSearch is in
Active state — there are no folded threads under an active filter."
```

---

### Task 18: App-level q handling — steal q when search is non-idle

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/app_test.go`

- [ ] **Step 1: Write a failing test for q-stolen behavior**

Append to `internal/ui/app_test.go`:

```go
func TestAppQuitStolenDuringSearch(t *testing.T) {
	t.Run("q during Active clears search, does not quit", func(t *testing.T) {
		styles := NewStyles(theme.Nord)
		backend := mail.NewMockBackend()
		app := NewApp(theme.Nord, backend, config.DefaultUIConfig())
		app, _ = app.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
		// Drive the initial Init chain.
		initCmd := app.Init()
		if initCmd != nil {
			msg := initCmd()
			app, cmd := app.Update(msg)
			// Drain the batch from selectionChangedCmds.
			if cmd != nil {
				inner := cmd()
				if batch, ok := inner.(tea.BatchMsg); ok {
					for _, sub := range batch {
						if sub != nil {
							app, _ = app.Update(sub())
						}
					}
				}
			}
			_ = styles
			_ = app
		}

		// Activate search and type a character, commit, then press q.
		app, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
		app, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		if cmd != nil {
			msg := cmd()
			if msg != nil {
				app, _ = app.Update(msg)
			}
		}
		app, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// Now in Active. Press q.
		_, cmd = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd != nil {
			msg := cmd()
			if _, isQuit := msg.(tea.QuitMsg); isQuit {
				t.Error("q during Active returned tea.Quit; should have cleared search")
			}
		}
	})
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ui/ -run TestAppQuitStolenDuringSearch -v`
Expected: FAIL — `App.Update` unconditionally returns `tea.Quit` on
`q` regardless of child state.

- [ ] **Step 3: Gate `q`-quit on the search being idle**

Edit `internal/ui/app.go`. Replace the `tea.KeyMsg` case in `Update`
(around line 64) with:

```go
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.acct.sidebarSearch.State() != SearchIdle {
				// Steal q while search is active so it doesn't quit
				// the app mid-search. Delegate to AccountTab which
				// clears the filter.
				var cmd tea.Cmd
				m.acct, cmd = m.acct.Update(tea.KeyMsg{Type: tea.KeyEsc})
				return m, cmd
			}
			return m, tea.Quit
		case "ctrl+c":
			return m, tea.Quit
		case "?":
			// Stubbed for 2.5b-5 (help popover)
			return m, nil
		}
```

Note: accessing `m.acct.sidebarSearch` directly is a reach into
child state. This is the minimum reach necessary — no state is
mutated and no field is written. The alternative (add an
`IsSearching()` method on `AccountTab`) is a nicer encapsulation
but adds an indirection for one call site. Keeping it direct for
now; revisit if a second consumer appears.

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/ui/ -run TestAppQuitStolenDuringSearch -v`
Expected: PASS.

- [ ] **Step 5: Run the full package tests**

Run: `go test ./internal/ui/ -v`
Expected: all PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/ui/app.go internal/ui/app_test.go
git commit -m "App: q clears search instead of quitting during Active

When SidebarSearch is non-idle, q is rewritten to Esc and delegated
to AccountTab, which clears the shelf and filter. This prevents
accidental quit while searching. q still quits normally from the
Idle account view."
```

---

## Phase 5 — Documentation

### Task 19: Write ADR 0064

**Files:**
- Create: `docs/poplar/decisions/0064-sidebar-search-shelf.md`

- [ ] **Step 1: Create the ADR**

Create `docs/poplar/decisions/0064-sidebar-search-shelf.md`:

```markdown
---
title: Sidebar search shelf with filter-and-hide semantics
status: accepted
date: 2026-04-13
---

## Context

Pass 2.5b-7 adds message search to the account view. Before this
ADR, `/` was reserved in keybindings but unbound, and wireframe
§7 #15 placed a post-commit search indicator in the status bar
without specifying where the input prompt lived or how thread
semantics interact with filter results. The previous 2.5b-3.7
"sidebar filter UI" pass (folder name filter) was deleted — a
handful of folders doesn't need a find affordance.

Two design axes needed to be settled: (1) where the search UI
lives, and (2) how matches affect the list.

For placement, status-bar prompts fight with existing transient
slots (undo bar, error banner, compose review prompt), and modal
overlays violate the one-pane rule. A 3-row shelf pinned to the
bottom of the sidebar column leaves the status bar free, preserves
the vim convention of "command line at the bottom," and keeps
folders as the primary content in the top of the sidebar.

For filter semantics, highlight-and-jump is vim-coherent but does
not map onto JMAP `Email/query` / IMAP `UID SEARCH` result sets,
which are natively "here is the set of matching messages." A
filter-and-hide shape composes cleanly with the backend search
that Pass 3 will wire — the local filter step becomes a backend
call, and the UI contract stays identical.

## Decision

Poplar ships message search in Pass 2.5b-7 with the following
shape:

- **Placement.** A 3-row shelf pinned to the bottom of the sidebar
  column, below the folder region. Always visible. Idle state
  shows `󰍉 / to search` as a hint; active states show the query
  and result count. The folder region's height subtracts
  `searchShelfRows = 3` from the total sidebar height.

- **Activation.** `/` from the account view (Idle state) enters
  Typing. Pressing `/` again from Active re-focuses the prompt
  with the existing query preserved.

- **Focus model.** Three states: Idle, Typing, Active. Typing is a
  brief modal state — printable runes append to the query,
  `Enter` commits to Active, `Esc` cancels. Active is the stable
  live-filter state — all normal account-view keys route normally.

- **Match semantics.** Filter-and-hide: non-matching threads are
  removed from the display. Thread-level predicate: any message
  in a thread matches → the whole thread is visible (root + all
  children) regardless of saved fold state. Fold state is
  preserved (not mutated) so `Esc` restores the pre-search layout.

- **Match algorithm.** Case-insensitive substring. Two modes:
  `[name]` (subject + sender, default) and `[all]` (subject +
  sender + date text). `Tab` cycles the mode while typing. No
  fuzzy matching, no regex.

- **Scope.** Current folder only. Folder jumps (`I/D/S/A/X/T`,
  `J/K`) clear the active search before loading the new folder.
  `q` is stolen from the quit handler while search is non-idle,
  clearing instead of quitting.

- **Result count.** Thread count (not message count). A thread
  with 4 matching replies counts as 1 result.

- **Backend contract reserved.** Pass 3 wires `backend.Search()`
  behind the same UI. JMAP `Email/query` returns a set of email
  IDs which get rendered through the same filter-and-hide
  pipeline. IMAP requires `UID THREAD REFERENCES` + `UID SEARCH`
  and client-side thread expansion — aerc's forked worker already
  supports both.

## Consequences

- A 3-row shelf becomes permanent sidebar chrome. Folder region
  loses 3 rows of vertical space on every render. The tradeoff
  is discoverability and consistent layout.
- Search is the second narrow modal state in poplar after
  visual-select (Pass 6). The "every key always live, no focus
  cycling" rule gains a documented exception for text input.
- `n/N` are aliased to `j/k` under filter-and-hide. Pass 3 may
  reinterpret `n/N` as "next/prev page of backend results" once
  backend pagination exists — the keys are reserved but not
  load-bearing in this pass.
- Fold keys (`Space/F/U`) become no-ops during Active search.
  This is a silent no-op for Pass 2.5b-7; Pass 2.5b-6 can add a
  toast explaining the behavior.
- `bubbles/textinput` is the first bubbles input component poplar
  imports. Further text-input features (compose, rename folder,
  etc.) should reuse the same library.
- Highlight-and-jump mode and configurable search behavior are
  explicitly deferred. Adding a global config knob for search
  mode would soften the "opinionated and not configurable in v1"
  invariant — that softening requires its own ADR.
```

- [ ] **Step 2: Commit**

```bash
git add docs/poplar/decisions/0064-sidebar-search-shelf.md
git commit -m "docs: add ADR 0064 for sidebar search shelf

Records placement (bottom-pinned 3-row shelf), filter-and-hide
semantics, thread-level predicate, focus model, mode toggle, scope
(current folder), and the Pass 3 backend contract."
```

---

### Task 20: Update invariants.md and the decision index

**Files:**
- Modify: `docs/poplar/invariants.md`

- [ ] **Step 1: Add the two new UX facts and the decision-index row**

Edit `docs/poplar/invariants.md`. Inside the `## UX` section, find
the line about `/` search activation ("Search" section of keybindings)
or add the new facts near the other keybinding / search facts.

Append these two facts to the UX section, as new bullets:

```markdown
- Search is activated by `/` from the account view. The search
  shelf lives in the bottom 3 rows of the sidebar column.
  Filter-and-hide: non-matching threads disappear; matching
  threads render with all children visible regardless of fold
  state. `Esc` clears the query and restores the pre-search
  cursor row.
- Search mode cycles between `[name]` (subject + sender) and
  `[all]` (subject + sender + date text) via `Tab` while the
  prompt is focused. Case-insensitive substring match. Scope is
  the current folder only. Folder jumps (`I/D/S/A/X/T` and
  `J/K`) clear the active search.
```

- [ ] **Step 2: Add the ADR to the decision index table**

At the end of `docs/poplar/invariants.md` in the Decision index
table, add a new row:

```markdown
| Search shelf, filter-and-hide, thread-level | 0064 |
```

- [ ] **Step 3: Verify line count constraint**

Invariants is supposed to stay around 150 lines. Run:

```bash
wc -l docs/poplar/invariants.md
```

If the file is growing past ~160 lines, look for existing facts
that are now obsolete or narrower than they need to be and trim.
The current state is unlikely to require trimming.

- [ ] **Step 4: Commit**

```bash
git add docs/poplar/invariants.md
git commit -m "docs: record sidebar search invariants

Adds two UX facts for the search shelf (activation, placement,
filter semantics, mode cycle, scope) and the ADR 0064 row in the
decision index."
```

---

### Task 21: Update wireframes.md

**Files:**
- Modify: `docs/poplar/wireframes.md`

- [ ] **Step 1: Add §2.1 Sidebar Search after the existing §2 Sidebar section**

Edit `docs/poplar/wireframes.md`. Find the `## 3. Message List`
heading (around line 131) and insert before it:

```markdown
## 2.1 Sidebar Search (Shelf)

The search shelf is a 3-row region pinned to the bottom of the
sidebar column. It is always visible, and hosts the query input,
match mode badge, and result count. See ADR 0064 for rationale.

### Idle

```
│   󰡡  Lists/golang        │
│   󰡡  Lists/rust          │
│                          │    <- unused space (folders don't fill)
│                          │
│                          │    <- shelf row 1 (blank separator)
│  󰍉 / to search           │    <- shelf row 2 (hint)
│                          │    <- shelf row 3 (reserved for mode/count)
```

### Typing (/ pressed, "proj" typed)

```
│                          │
│                          │
│                          │
│  󰍉 /proj▏                │
│  [name]       3 results  │
```

### Committed (Enter pressed)

```
│                          │
│                          │
│                          │
│  󰍉 /proj                 │
│  [name]       3 results  │
```

### No results

```
│                          │
│                          │
│                          │
│  󰍉 /asdf▏                │
│  [name]      no results  │
```

And the message list simultaneously shows a centered placeholder
distinct from the empty-folder state:

```
│                         │                                                                               │
│                         │                                                                               │
│                         │                       No matches                                               │
│                         │                                                                               │
```

**Annotations:**

- **Activation.** `/` in Idle state. `/` in Active re-focuses the
  prompt with the existing query preserved. `Tab` cycles the mode
  badge between `[name]` and `[all]`. `Enter` commits (Typing →
  Active). `Esc` clears (any state → Idle).
- **Colors.** Icon `󰍉` in `fg_dim` (idle) or `accent_primary`
  (active). Query text in `fg_base` (typing) or `fg_bright`
  (committed). Mode badge brackets in `fg_dim`, label in `fg_base`
  (idle) or `accent_tertiary` (typing). Result count in
  `accent_tertiary`. "no results" in `color_warning` dim.
- **Layout.** 30-col sidebar. Prompt row has 25 cells for query
  text (1 indent + 2-cell icon + 1 space + 1 "/" = 5 cells chrome).
  Mode/count row right-aligns the count with a flex gap of at
  least 1 cell between the mode badge and the count text.
- **Pinned.** The 3-row shelf is always at the bottom of the
  sidebar column. Folders flow from the top; any empty space sits
  between folders and the shelf. The folder region's height is
  `accountTabHeight − sidebarHeaderRows − searchShelfRows`.

```

- [ ] **Step 2: Update §1 Composite Layout to mention the shelf**

In `docs/poplar/wireframes.md`, find the §1 Composite Layout
annotations block (around line 48). Locate the bullet about
**Sidebar** and append a new bullet after the existing sidebar
bullet:

```markdown
- **Sidebar search shelf**: The bottom 3 rows of the sidebar
  column are the persistent search shelf (see §2.1). When idle,
  they show the hint `󰍉 / to search`; when active, they host the
  query input and result count. The sidebar column composition
  top-to-bottom is: account header (2 rows) + folder region
  (flex, scrollable) + search shelf (3 rows, pinned).
```

- [ ] **Step 3: Replace §7 #15 Search Results**

In `docs/poplar/wireframes.md`, find `### Search results (#15)`
(around line 487). Replace the entire subsection (heading through
the empty line before the next subsection) with:

```markdown
### Search filter applied (#15)

Search is hosted in the sidebar shelf (§2.1), not the status bar.
When a filter is active, the message list displays only matching
threads (root + all children when any message in the thread
matches). The status bar retains its normal contents — message
counts, connection status — and is not used as a search indicator.

```
│ geoff@907.life           │                                                                               │
│                          │  󰇮  Alice Johnson            Re: Project update for Q2 launch         10:32 AM │
│   󰇰  Inbox           3  │  󰑚  Carol White              Re: Project budget review              Yesterday │
│   󰏫  Drafts              │                                                                               │
│   󰑚  Sent                │                                                                               │
│   󰀼  Archive             │                                                                               │
│                          │                                                                               │
│   󰍷  Spam           12   │                                                                               │
│                          │                                                                               │
│  󰍉 /proj                 │                                                                               │
│  [name]       2 results  │                                                                               │
 ──────────────────────────┴──────────────────────────────── 10 messages · 3 unread · ● connected ─╯
```

When no messages match, the list shows a centered placeholder
distinct from the empty-folder state of §7 #13:

```
│                         │                                                                               │
│                         │                                                                               │
│                         │                       No matches                                               │
│                         │                                                                               │
```

`n/N` walk the filtered row set (aliases for `j/k`). `Esc` clears
the filter and restores the full list plus the pre-search cursor
row.
```

- [ ] **Step 4: Commit**

```bash
git add docs/poplar/wireframes.md
git commit -m "docs(wireframes): add §2.1 sidebar search, update §1 and §7 #15

New §2.1 shows all four states of the search shelf: idle, typing,
committed, no-results. §1 annotations gain a sidebar-search bullet.
§7 #15 is retitled 'Search filter applied' and shows the message
list under an active filter with the shelf visible in the sidebar —
the old status-bar mockup is removed because search no longer lives
there."
```

---

### Task 22: Update keybindings.md

**Files:**
- Modify: `docs/poplar/keybindings.md`

- [ ] **Step 1: Promote `/`, `n`, `N` from stub to live**

Edit `docs/poplar/keybindings.md`. Find the `## Search` table
(around line 65). The entries for `/`, `n`, `N` are already listed
as live; update the action description so it reflects the sidebar
shelf and filter-and-hide semantics:

```markdown
## Search

| Key | Action | Context |
|-----|--------|---------|
| `/` | Start or re-focus sidebar search shelf | A |
| `n` | Next match (alias for `j` under filter) | A |
| `N` | Previous match (alias for `k` under filter) | A |
| `Tab` | Cycle match mode `[name]` ↔ `[all]` (while typing) | A |
| `Enter` | Commit query (Typing → Active) | A |
| `Esc` | Clear query, restore pre-search cursor | A |
```

- [ ] **Step 2: Add a search-context routing note in the Design Decisions section**

At the end of `docs/poplar/keybindings.md`, in the `## Design
Decisions` section, after the `**Non-modal for pane focus...**`
bullet, add:

```markdown
**Search is the second narrow modal state.** After visual-select
(Pass 6), sidebar search is the only other place poplar accepts a
narrow modal keyboard routing. In Typing state, every printable
rune — letters, digits, punctuation, space, `q`, `F`, `U`, `?`,
`j`, `k`, arrows — is appended to the query. Only `Tab`, `Enter`,
`Esc`, `Backspace`, and `Left/Right` arrows have special meaning.
Once the query is committed with `Enter`, the shelf enters Active
state and all normal account-view keys (including `j/k`, folder
jumps, triage) route normally again. `q` is stolen while the shelf
is non-idle to prevent accidental quit. See ADR 0064.
```

- [ ] **Step 3: Commit**

```bash
git add docs/poplar/keybindings.md
git commit -m "docs(keybindings): promote /, n, N to live; add Tab/Enter/Esc in Search

Pass 2.5b-7 ships the sidebar search shelf. Search table gains
Tab (mode cycle), Enter (commit), and Esc (clear) rows. Design
decisions section gains a narrow-modal-state note explaining
Typing-state key routing and the q-stolen rule."
```

---

### Task 23: Update STATUS.md

**Files:**
- Modify: `docs/poplar/STATUS.md`

- [ ] **Step 1: Drop 2.5b-3.7, promote 2.5b-7, rewrite starter prompt**

Edit `docs/poplar/STATUS.md`. Replace the entire file contents with:

```markdown
# Poplar Status

**Current pass:** Pass 2.5b-7 (sidebar search) in progress.
ADR 0064 records filter-and-hide semantics with a bottom-pinned
sidebar shelf.

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1 | Scaffold + Fork | done |
| 2 | Backend Adapter + Connect | done |
| 2.5-render | Lipgloss migration: block model + compiled themes | done |
| 2.5-fix | Fix first-level blockquote wrapping | done |
| 2.5a | Text wireframes for all screens | done |
| 2.5b-1 | Prototype: chrome shell | done |
| 2.5b-keys | Keybinding design | done |
| 2.5b-chrome | Chrome redesign | done |
| 2.5b-2 | Prototype: sidebar | done |
| 2.5b-3 | Prototype: message list | done |
| 2.5b-3.5 | Prototype: UI config + sidebar polish | done |
| 2.5b-3.6 | Prototype: threading + fold | done |
| 2.5b-7 | Prototype: sidebar search | in progress |
| 2.5b-4 | Prototype: message viewer | next |
| 2.5b-5 | Prototype: help popover | pending |
| 2.5b-6 | Prototype: status/toast system | pending |
| 2.5b-train | Tooling: mailrender training capture system | pending (after Pass 3) |
| 3 | Wire prototype to live backend | pending |
| 6 | Triage actions | pending |
| 8 | Gmail IMAP | pending |
| 9 | Compose + send (Catkin editor) | pending |
| 9.5 | Tidytext in compose | pending |
| 10 | Config | pending |
| 11 | Polish for daily use | pending |
| 1.1 | Neovim embedding (nvim --embed RPC) | pending |

## Next starter prompt (Pass 2.5b-4)

> **Goal.** Message viewer prototype — open a message in the
> right panel (replacing the message list), render header block +
> body via the existing ParseBlocks/RenderBody pipeline, support
> `q` to close.
>
> **Settled.** Sidebar remains visible (ADR 0025); viewer opens
> in the right panel with `q` returning to the list (wireframes
> §4); content pipeline already exists from Pass 2.5-render.
>
> **Still open — brainstorm:** body fetch path (sync vs async
> with spinner); link picker scope (in or out of this pass); auto
> mark-read on open (in or out); interaction with active search
> state if viewer opens into a filtered list.
>
> **Approach.** Brainstorm, write spec + plan under
> `docs/superpowers/{specs,plans}/`, implement via
> `subagent-driven-development`. Pass-end via `poplar-pass`.
```

- [ ] **Step 2: Commit**

```bash
git add docs/poplar/STATUS.md
git commit -m "docs(STATUS): drop 2.5b-3.7, promote 2.5b-7 current, queue 2.5b-4 next

The sidebar folder filter pass (2.5b-3.7) is deleted — a handful
of folders doesn't need a find affordance. Message search
(formerly 2.5b-7) is pulled forward as the current pass. Message
viewer (2.5b-4) becomes the next starter prompt."
```

---

## Phase 6 — Verification

### Task 24: Run `make check` and fix any issues

**Files:**
- (any that need fixes)

- [ ] **Step 1: Run vet and tests**

Run: `make check`
Expected: exits 0 with all tests passing.

- [ ] **Step 2: If vet reports issues, fix them before proceeding**

Common issues to watch for:
- Unused imports from earlier intermediate states
- Shadowed variables
- Missing error handling (shouldn't occur in this pass — no new
  error paths were added)

Fix inline and re-run `make check` until it passes.

- [ ] **Step 3: Run gofmt**

Run: `gofmt -l internal/ui/`
Expected: no output. If any files are listed, they have formatting
issues — run `gofmt -w internal/ui/` to fix.

- [ ] **Step 4: Commit any fixes**

If Step 2 or Step 3 produced changes, commit them:

```bash
git add -u internal/ui/
git commit -m "ui: fix vet/format issues after search pass"
```

If no changes were needed, skip the commit.

---

### Task 25: Install and verify live UI

**Files:**
- (no file changes, verification only)

- [ ] **Step 1: Install the binary**

Run: `make install`
Expected: `poplar` built and copied to `~/.local/bin/poplar`.

- [ ] **Step 2: Start poplar under tmux for capture**

Follow `.claude/docs/tmux-testing.md` — start a named tmux session,
launch `poplar` inside it, and capture a screenshot of the default
view.

Expected: sidebar shows folders + bottom shelf with `󰍉 / to
search` hint in the bottom 3 rows.

- [ ] **Step 3: Press `/` and type a query**

Press `/`, then type a sender name that exists in the mock data
(e.g., `alice`). Expected:

- Shelf transitions to Typing state — cursor appears, query text
  shows as you type.
- Mode badge `[name]` is visible.
- Message list filters to matching threads only.
- Result count updates on each keystroke.

Capture a screenshot of the mid-typing state.

- [ ] **Step 4: Press `Tab` to cycle mode**

Expected: mode badge switches to `[all]`. If any message in the
list has a date text matching your current query, it should now
also match; count may change. Press `Tab` again to return to
`[name]`.

- [ ] **Step 5: Press `Enter` to commit**

Expected: cursor on prompt disappears; query text remains; mode
badge and count remain. State transitioned to Active.

- [ ] **Step 6: Try `j/k` navigation**

Expected: cursor walks the filtered row set. `n/N` also walk the
set (aliases for `j/k`).

- [ ] **Step 7: Press `Esc`**

Expected: shelf returns to Idle; full message list restored;
cursor is back on its pre-search row.

- [ ] **Step 8: Try the folder-jump-clears flow**

Press `/`, type `alice`, `Enter`, then press `D` (Drafts) or `J`.
Expected: search clears, folder changes, Drafts (or the next
folder) loads in Idle state.

- [ ] **Step 9: Try the no-results flow**

Press `/`, type `zzz-nothing-matches`. Expected:

- Shelf shows `no results` in the info row (warning color).
- Message list shows `No matches` centered placeholder.
- `Esc` clears and restores normally.

- [ ] **Step 10: Try `q` during search**

Press `/`, type `alice`, `Enter` (now in Active), then `q`.
Expected: search clears, poplar does NOT quit. Press `q` again
from Idle → poplar quits.

- [ ] **Step 11: If any live test fails, investigate and fix**

Common issues:
- Width math off at very small terminal sizes (the shelf fights
  the folder region)
- Cursor rendering glitches (textinput width may need tuning)
- Color mismatch (check the style slots match the expected theme
  palette slot)

Fix inline, re-run `make check`, re-install, re-verify.

- [ ] **Step 12: Commit any live-verification fixes**

If Step 11 produced changes:

```bash
git add -u internal/ui/
git commit -m "ui: live verification fixes for sidebar search"
```

- [ ] **Step 13: Report pass completion**

At this point the implementation is done. Proceed to the pass-end
ritual via the `poplar-pass` skill: simplify, ADRs already written,
invariants already updated, STATUS already updated, commit, push,
install.

---

## Scope note: Out of plan

The following items from the spec's scope fence are NOT implemented
in this plan — they are deferred to later passes:

- **Backend search (JMAP / IMAP).** This plan is local-only against
  the mock backend. Pass 3 wires real search behind the same UI.
- **Highlight-and-jump mode.** Filter-and-hide is the single mode;
  no config knob.
- **Fuzzy or regex matching.** Substring only.
- **Cross-folder search.** Scope is the current folder.
- **Search history / saved queries.**
- **Toast on "fold disabled during search."** Silent no-op for
  now; Pass 2.5b-6 adds the toast system.

If a task in this plan accidentally introduces any of these, treat
it as a bug and strip the extra code before committing.

---
title: Sidebar search — design
pass: 2.5b-7 (pulled forward from later in the 2.5b prototype track)
status: approved, ready for plan
date: 2026-04-13
---

# Sidebar Search — Design

Local, filter-and-hide message search activated by `/`, hosted in a
3-row shelf pinned to the bottom of the sidebar column. This pass
ships the full UI against the mock backend. Real backend-side search
(JMAP `Email/query` / IMAP `UID SEARCH`) is reserved for Pass 3 and
the UI contract is designed to accept it without rework.

## Goal

Give the user a way to narrow the current folder's message list by
typing a substring. Matching threads (whole threads, root + all
children) remain; everything else disappears. `Esc` restores the
full list and the pre-search cursor position.

The sidebar gains a persistent 3-row shelf at the bottom that shows
either the search hint (idle) or the live query + mode + result count
(active). The shelf is always visible — discoverable without having
to learn `/`.

## Scope

### In scope

- New `SidebarSearch` subcomponent rendered at the bottom of the
  sidebar column.
- Filter state inside `MessageList` with cursor save/restore.
- Key routing for `/`, `Esc`, `Enter`, `Tab`, and printable runes
  during typing.
- Two-stop mode toggle: `[name]` (subject + sender) and `[all]`
  (subject + sender + date text). `Tab` cycles while typing.
- Case-insensitive substring match.
- Thread-level predicate (any message matches ⇒ whole thread kept).
- Bottom-pinned layout: sidebar column becomes `account header (2) →
  folder region (flex, scrollable) → search shelf (3, pinned)`.
- Promoting `n/N` from reserved to live as aliases for `j/k` over
  the filtered row set.
- ADR 0064 capturing all of the above.
- Wireframes, keybindings, invariants, and STATUS.md documentation
  updates.

### Out of scope

- **Backend search.** This pass is local-only against messages
  already loaded into `MessageList.source`. Pass 3 will wire
  `backend.Search(query)` behind the same UI.
- **Highlight-and-jump mode.** Filter-and-hide is the single mode.
  No config knob. Revisit in a later pass if real usage reveals a
  missing affordance.
- **Fuzzy or regex matching.** Substring only.
- **Cross-folder search.** Scope is the current folder.
- **Search history / saved queries.**
- **`n/N` as "next page of backend results."** Meaning is reserved
  for Pass 3 to reinterpret if backend pagination demands it.
- **Toast on "fold disabled during search."** Pass 2.5b-6 will add
  the toast infrastructure; for this pass, fold keys are silent
  no-ops during search.
- **Search against folders.** A standalone 2.5b-3.7 "sidebar filter
  UI" pass was previously queued and is now deleted — a handful of
  folders doesn't need a find affordance.

## UX

### States

```
         ┌──────┐
         │ Idle │ ◄────────────────┐
         └──┬───┘                  │
            │ /                    │
            ▼                      │
        ┌────────┐  Esc            │
        │ Typing │ ──────────────► │
        └──┬─────┘                 │
           │ Enter                 │
           ▼                       │
        ┌────────┐  Esc            │
        │ Active │ ──────────────► │
        └───┬────┘                 │
            │ /                    │
            └──► back to Typing    │
```

- **Idle** — no filter; shelf shows the hint row. Normal account-view
  keybindings. Pressing `/` focuses the prompt and enters Typing.
- **Typing** — prompt focused. Printable runes append to the query;
  the filter updates live on each keystroke. `Tab` cycles the mode
  badge. `Esc` cancels (clears the query, restores cursor, → Idle).
  `Enter` commits (unfocuses the prompt, keeps the filter live,
  → Active).
- **Active** — filter is live, prompt is unfocused. All normal
  account-view keys route normally. `j/k` walk the filtered row
  set. `n/N` alias `j/k`. `Esc` clears the query and returns to
  Idle. `/` re-focuses the prompt with the existing query and
  returns to Typing. Folder jumps (`I/D/S/A/X/T`, `J/K`) clear
  search before loading the new folder.

### Key routing

| Key | Idle | Typing | Active |
|---|---|---|---|
| `/` | → Typing | (typed rune) | → Typing (existing query preserved) |
| printable rune | normal | append to query, filter updates | normal |
| `Backspace` | normal | delete char, filter updates | normal |
| `Tab` | normal | cycle mode (`[name]` ↔ `[all]`) | normal |
| `Enter` | open msg | → Active | open msg |
| `Esc` | normal | cancel (clear query, restore cursor, → Idle) | clear query, restore cursor, → Idle |
| `j/k` | nav | **ignored** | nav filtered list |
| `J/K` | folder nav | **ignored** | folder nav → clears search first |
| `g/G` | top/bottom | ignored | top/bottom of filtered list |
| `I/D/S/A/X/T` | folder jump | ignored | folder jump → clears search first |
| `d/a/s/.` | triage | ignored | triage (within filtered list) |
| `r/R/f/c` | reply/compose | ignored | reply/compose |
| `Space/F/U` | fold | (typed as runes) | **no-op** (threads already expanded) |
| `n/N` | (stubs) | (typed as runes) | alias for `j/k` |
| `q` | quit | (typed as rune) | → clear → Idle (no quit) |
| `?` | help | (typed as rune) | help |

Notes:

- During Typing, **any printable rune** (letters, digits, punctuation,
  space) is appended to the query. This includes `q`, `Space`, `F`,
  `U`, `j`, `k`, `?`, `n`, `N` — they are all just characters the
  user is typing into the search box. Only named keys (`Tab`,
  `Enter`, `Esc`, `Backspace`, `Left`/`Right` arrows) are handled
  specially by the input.
- `q` is stolen only in Active state (not Typing) to prevent
  accidental quit while searching. Same pattern as visual-select
  mode (Pass 6). In Typing state `q` is just a character; the user
  exits Typing with `Esc`.
- Fold keys are no-ops during Active state because the filter
  expands all matching threads regardless of saved fold state. They
  are not ignored or consumed — they simply find no folded thread
  to operate on.
- The `j/k/J/K/g/G/I/D/S/A/X/T/d/a/s/./r/R/f/c` rows above say
  "ignored" during Typing. That's shorthand — technically, every
  one of those keys is a printable rune and gets appended to the
  query. They are not routed to their normal handlers. The user
  sees them in the query text as they type.

### Visual layout — sidebar shelf

The sidebar column is composed top-to-bottom as:

```
account header  (2 rows, fixed)       " geoff@907.life" + blank
folder region   (flex, scrollable)    sidebar.View()
search shelf    (3 rows, pinned)      sidebarSearch.View()
```

The folder region's height is `accountTabHeight − sidebarHeaderRows
− searchShelfRows`. The `accountTabHeight` passed in via
`tea.WindowSizeMsg` is already post-chrome (see `App.contentHeight`
in `internal/ui/app.go` which subtracts 3 rows for top frame +
status + footer). So the math at `AccountTab.updateTab` time is:

```go
const searchShelfRows = 3
sw := min(sidebarWidth, m.width/2)
folderHeight := m.height - sidebarHeaderRows - searchShelfRows
m.sidebar.SetSize(sw, max(1, folderHeight))
m.sidebarSearch.SetSize(sw)
```

If folders exceed `folderHeight` they clip with `J/K` scrolling;
the shelf stays pinned and is never scrolled out of view.

### Shelf states

**Idle:**

```
│   󰡡  Lists/golang        │
│   󰡡  Lists/rust          │
│                          │    <- unused space
│                          │    <- unused space
│                          │    <- shelf row 1 (blank separator)
│  󰍉 / to search           │    <- shelf row 2 (hint)
│                          │    <- shelf row 3 (reserved)
```

**Typing (`/` pressed, `proj` typed so far):**

```
│   󰡡  Lists/golang        │
│   󰡡  Lists/rust          │
│                          │
│                          │
│                          │
│  󰍉 /proj▏                │
│  [name]       3 results  │
```

**Committed (post-Enter, prompt unfocused):**

```
│   󰡡  Lists/golang        │
│   󰡡  Lists/rust          │
│                          │
│                          │
│                          │
│  󰍉 /proj                 │
│  [name]       3 results  │
```

Difference from Typing: cursor `▏` is gone; query text shifts from
`fg_base` to `fg_bright` (signaling "locked query").

**No results:**

```
│                          │
│                          │
│                          │
│  󰍉 /asdf▏                │
│  [name]      no results  │
```

With the message list showing a centered placeholder distinct from
the `§7 #13` empty-folder state:

```
│                         │                                                                               │
│                         │                                                                               │
│                         │                       No matches                                               │
│                         │                                                                               │
│                         │                                                                               │
```

### Colors

| Element | Style slot | Color |
|---|---|---|
| `󰍉` icon | `SearchIcon` (new) | `fg_dim` idle, `accent_primary` active |
| Query text (`/proj`) | `SearchPrompt` (new) | `fg_base` typing, `fg_bright` committed |
| Cursor `▏` | textinput default | `accent_primary` |
| Mode badge `[name]` / `[all]` | `SearchModeBadge` (new) | `fg_dim` brackets, `fg_base` label idle, `accent_tertiary` label during typing |
| Result count (`3 results`) | `SearchResultCount` (new) | `accent_tertiary` |
| "no results" text | `SearchNoResults` (new) | `color_warning` dim |
| "No matches" placeholder in list | `MsgListPlaceholder` (new) | `fg_dim`, centered |
| Hint text (`󰍉 / to search`) | `SearchHint` (new) | `fg_dim` |

Seven new style slots. All are populated in `NewStyles` from the
existing `theme.CompiledTheme` palette — no new palette colors.

### Width math (30-col sidebar)

- **Prompt row (row 2):** 1 indent + icon (2) + space + `/` + query
  text. Query area = `30 − 5 = 25 cells`.
- **Mode/count row (row 3):** 1 indent + `[name]` or `[all]`
  (6 cells) + flex gap + right-aligned count (worst case
  `no results` = 10 cells) + 1 margin. Flex gap worst case =
  `30 − 2 − 6 − 10 − 1 = 11 cells`.

## Architecture

### New component: `SidebarSearch`

New file `internal/ui/sidebar_search.go`.

```go
type SearchState int

const (
    SearchIdle SearchState = iota
    SearchTyping
    SearchActive
)

type SearchMode int

const (
    SearchModeName SearchMode = iota  // subject + sender (default)
    SearchModeAll                     // + date text
)

type SidebarSearch struct {
    input   textinput.Model
    mode    SearchMode
    state   SearchState
    results int
    styles  Styles
    width   int
}

func NewSidebarSearch(styles Styles, width int) SidebarSearch
func (s *SidebarSearch) Update(msg tea.Msg) (SidebarSearch, tea.Cmd)
func (s SidebarSearch) View() string
func (s SidebarSearch) State() SearchState
func (s SidebarSearch) Query() string
func (s SidebarSearch) Mode() SearchMode
func (s *SidebarSearch) SetResultCount(n int)
func (s *SidebarSearch) Activate()    // Idle → Typing
func (s *SidebarSearch) Commit()      // Typing → Active
func (s *SidebarSearch) Clear()       // any → Idle
```

`SidebarSearch` does NOT know about `MessageList`. It owns only its
own input, mode, and display state. All communication to the rest
of the UI goes through Msg types that flow through `AccountTab`.

The `bubbles/textinput` dependency is added to `go.mod` — it's a
canonical charm library, not a new ecosystem.

### New state in `MessageList`

Two new fields plus filter logic:

```go
type MessageList struct {
    // ... existing fields
    filter          searchFilter
    preSearchCursor int
    savedByFilter   bool
}

type searchFilter struct {
    query string
    mode  SearchMode
}

func (m *MessageList) SetFilter(q string, mode SearchMode)
func (m *MessageList) ClearFilter()
func (m *MessageList) FilterResultCount() int
```

### New Msg types

One Msg type in `internal/ui/cmds.go`:

```go
type SearchUpdatedMsg struct{ Query string; Mode SearchMode } // keystroke or mode cycle
```

`SidebarSearch.Update` emits `SearchUpdatedMsg` as a Cmd whenever
the query value or mode changes during Typing state. No other
search-related Msg types exist — state transitions (Activate,
Commit, Clear) are driven by direct method calls from `AccountTab`,
which holds both children and is the natural place to sequence
them.

### Data flow

```
/ pressed (Idle)
  → AccountTab.handleKey: m.sidebarSearch.Activate()
  → SidebarSearch: state = Typing, input.Focus()

printable rune pressed (Typing)
  → AccountTab.handleKey: m.sidebarSearch, cmd = m.sidebarSearch.Update(msg)
  → SidebarSearch.Update: textinput appends, emits SearchUpdatedMsg as Cmd
  → cmd bubbles up to bubbletea runtime, SearchUpdatedMsg returns to Update
  → AccountTab.updateTab on SearchUpdatedMsg:
      m.msglist.SetFilter(query, mode)
      m.sidebarSearch.SetResultCount(m.msglist.FilterResultCount())

Tab pressed (Typing)
  → AccountTab.handleKey: route to SidebarSearch.Update
  → SidebarSearch.Update: intercepts Tab, cycles mode, emits SearchUpdatedMsg
  → (same downstream path as printable runes)

Enter pressed (Typing)
  → AccountTab.handleKey: m.sidebarSearch.Commit()
  → SidebarSearch: state = Active, input.Blur()
  → filter is unchanged — stays live from the last SearchUpdatedMsg

Esc pressed (Typing or Active)
  → AccountTab.handleKey: m.sidebarSearch.Clear(); m.msglist.ClearFilter()
  → SidebarSearch: state = Idle, input.Reset(), mode reset
  → MessageList: filter cleared, pre-search cursor restored

Folder jump (J/K/I/D/S/A/X/T) during Active
  → AccountTab.handleKey: clearSearchIfActive(), then run the jump
```

State ownership at `AccountTab`: it holds both children and
sequences them via direct method calls for state transitions, plus
one Msg (`SearchUpdatedMsg`) for the textinput-driven query-change
feedback loop. This is a deliberate trade-off against "pure Msg
flow only" — `AccountTab` is the single natural coordination point
and adding two more Msg types per state transition would be
ceremony without added isolation. Elm-conventions Rule 4 still
holds: neither child reaches into the other.

### Files touched

**New:**

- `internal/ui/sidebar_search.go`
- `internal/ui/sidebar_search_test.go`
- `docs/poplar/decisions/0064-sidebar-search-shelf.md`

**Modified:**

- `internal/ui/msglist.go` — filter field, `SetFilter`/`ClearFilter`,
  `filterBuckets` pipeline step, cursor save/restore, fold-shadow
  during filter, result count accessor
- `internal/ui/msglist_test.go` — filter tests (see Test Plan)
- `internal/ui/account_tab.go` — key routing for `/`, `Tab`, `Esc`,
  `Enter`, printable runes during Typing; wire Msg types;
  `searchShelfRows` constant; sidebar height math; folder-jump clears
  search
- `internal/ui/account_tab_test.go` — activation + clear flow tests
- `internal/ui/cmds.go` — new Msg types
- `internal/ui/styles.go` — seven new style slots
- `internal/ui/styles_test.go` — new-slot tests
- `internal/ui/keys.go` — bind `/` if not already
- `go.mod` / `go.sum` — add `github.com/charmbracelet/bubbles/textinput`

**Documentation:**

- `docs/poplar/wireframes.md` — new §2.1, updated §1, replaced §7 #15
- `docs/poplar/keybindings.md` — promote `/`, `n`, `N` to live; add
  Typing state routing notes; add "second narrow modal" design note
- `docs/poplar/invariants.md` — 2 new UX facts, updated decision
  index
- `docs/poplar/STATUS.md` — drop 2.5b-3.7, promote 2.5b-7, rewrite
  starter prompt for 2.5b-4 (next pass after this one)

## Filter mechanics

### Pipeline

Current:
```
source → bucket by ThreadID → sort threads → flatten → rows
```

New:
```
source → bucket by ThreadID → filterBuckets → sort threads → flatten → rows
```

`filterBuckets` is a no-op when `filter.query == ""`. When non-empty:

```go
func (m *MessageList) filterBuckets(buckets []threadBucket) []threadBucket {
    if m.filter.query == "" {
        return buckets
    }
    q := strings.ToLower(m.filter.query)
    out := buckets[:0]
    for _, b := range buckets {
        if anyMessageMatches(b.messages, q, m.filter.mode) {
            out = append(out, b)
        }
    }
    return out
}

func matchOne(msg mail.MessageInfo, q string, mode SearchMode) bool {
    if containsFold(msg.Subject, q) || containsFold(msg.From, q) {
        return true
    }
    if mode == SearchModeAll && containsFold(formatDate(msg.Date), q) {
        return true
    }
    return false
}

func containsFold(s, q string) bool {
    return strings.Contains(strings.ToLower(s), q)
}
```

The query `q` is pre-lowercased once per filter application.
Per-field lowercasing happens inside `containsFold`. No regex, no
fuzzy ranker. Zero allocation beyond the lowered field strings
that Go's `strings.ToLower` returns.

### Cursor save/restore

```go
func (m *MessageList) SetFilter(q string, mode SearchMode) {
    if !m.savedByFilter && q != "" {
        m.preSearchCursor = m.cursor
        m.savedByFilter = true
    }
    m.filter = searchFilter{query: q, mode: mode}
    m.rebuild()
    m.clampCursor()
}

func (m *MessageList) ClearFilter() {
    m.filter = searchFilter{}
    m.rebuild()
    if m.savedByFilter {
        m.cursor = m.preSearchCursor
        m.clampCursor()
        m.savedByFilter = false
    }
}
```

The `savedByFilter` gate means:

- First keystroke in Typing: saves pre-search cursor.
- Subsequent keystrokes while Typing: do NOT overwrite the save.
- Clear in any state: restores and clears the gate.
- Re-activating search after clear: saves the new pre-search cursor.

If the pre-search cursor points at a row that no longer exists
(e.g., source was replaced mid-search in Pass 3), `clampCursor()`
handles it with `min(cursor, len(rows)-1)`.

### Thread-level fold shadowing

Active filter forces all matching threads expanded regardless of
saved fold state:

```go
func (m *MessageList) effectiveFold(threadID string) bool {
    if m.filter.query != "" {
        return false
    }
    return m.isFolded(threadID)
}
```

`appendThreadRows` and render paths query `effectiveFold` instead
of `isFolded`. Saved fold state is preserved (not mutated) so that
`Esc` restores the pre-search fold layout.

### Result count semantics

`FilterResultCount()` returns the **thread count**, not the message
count. A thread of 4 replies with all 4 matching is 1 result, not 4.

Rationale: the filter's unit of work is the thread. Counting
messages would show "5 results" alongside a list of 3 visible
threads (because 3 of the 5 collapsed into one thread), which
would confuse users.

Status bar is untouched by search — the normal "10 messages ·
3 unread · ● connected" remains. Result count lives only in the
sidebar shelf.

### No-results state

When `filter.query != ""` and `len(rows) == 0`:

- Shelf shows `no results` in `color_warning` dim.
- Message list renders centered `No matches` in `fg_dim`.
- Cursor is 0; `j/k`/`n/N` are no-ops.
- Triage keys are no-ops (no current message).
- `Esc` restores normally.

The "No matches" wording is deliberately distinct from `§7 #13`'s
"No messages" (empty folder) so users can distinguish "I filtered
to nothing" from "there's nothing in this folder."

### Edge cases

- **Folder reload mid-search** (Pass 3 background refresh). New
  `source` replaces old, filter re-runs automatically through
  `SetMessages` → `rebuild`. Result count updates. Cursor clamps if
  invalid.
- **Folder switch mid-search.** Handled at `AccountTab`: folder jump
  calls `msglist.ClearFilter()` before `loadFolderCmd`. The clear
  resolves cursor restore against the old source (safe), and the
  new folder loads in Idle state.
- **New message arrives mid-search.** Same path as reload — new
  source, new filter run, matching new messages appear in the
  filtered view.

## Documentation impact

### ADR 0064

New file `docs/poplar/decisions/0064-sidebar-search-shelf.md`.
Records the placement (bottom-pinned 3-row sidebar shelf), filter
semantics (filter-and-hide, thread-level, case-insensitive
substring), focus model (Typing brief-modal → Active live), mode
toggle (`[name]` / `[all]` via `Tab`), folder-jump-clears-search,
and the deferrals to Pass 3.

### Invariants.md

Add to the UX section (edited in place, not appended):

1. Search is activated by `/` from the account view. The search
   shelf lives in the bottom 3 rows of the sidebar column.
   Filter-and-hide: non-matching threads disappear; matching
   threads render with all children visible regardless of fold
   state. `Esc` clears the query and restores the pre-search
   cursor row.

2. Search mode cycles between `[name]` (subject + sender) and
   `[all]` (subject + sender + date text) via `Tab` while the
   prompt is focused. Case-insensitive substring match. Scope is
   the current folder only. Folder jumps (`I/D/S/A/X/T` and `J/K`)
   clear the active search.

Update the decision index with a new row:

| Search shelf, filter-and-hide, thread-level | 0064 |

### Wireframes.md

- **New §2.1 Sidebar Search** — mockups for idle / typing /
  committed / no-results states, color annotations, width math.
- **§1 Composite Layout updated** — sidebar composition note
  showing the bottom-pinned shelf.
- **§7 #15 Search Results replaced** — retitled "Search Filter
  Applied." Shows the message list under an active filter with
  the shelf visible in the sidebar, plus the no-matches
  placeholder. Status bar search indicator mockup removed.

### Keybindings.md

- Promote `/`, `n`, `N` from reserved/stub to live.
- Add search-context routing notes for Typing and Active states.
- Add a design decision note: "Search is the second narrow modal
  state after visual-select. Typing mode captures printable runes
  for the duration of the text input. Everywhere else poplar is
  one-pane-live."

### STATUS.md

- Mark 2.5b-3.6 `done` (already correct).
- **Delete** the `2.5b-3.7 Prototype: sidebar filter UI` row.
  That pass was a mistake — folders don't need a find affordance.
- Promote `2.5b-7 Prototype: search` to `next`.
- After this pass, the next starter prompt should be for
  `2.5b-4 Prototype: message viewer`.

## Test plan

### `internal/ui/sidebar_search_test.go` (new)

- **Idle rendering** — hint row visible, blank separator and
  reserved row present, mode badge hidden.
- **Typing rendering** — prompt visible with cursor, query text in
  `fg_base`, mode badge visible with label in `accent_tertiary`,
  result count visible.
- **Committed rendering** — prompt visible without cursor, query in
  `fg_bright`, mode badge and count visible.
- **No-results rendering** — `no results` text replaces the count
  in `color_warning`.
- **Mode cycle** — `Tab` cycles `[name]` → `[all]` → `[name]`.
- **Activate transition** — `SidebarIdle` → `Typing` on `Activate()`,
  input focused.
- **Commit transition** — `Typing` → `Active` on Enter; input
  blurred; query preserved.
- **Clear transitions** — both `Typing → Idle` and `Active → Idle`
  on `Clear()`; input reset; mode reset to `SearchModeName`.
- **Re-activate preserves nothing** — after Clear, Activate starts
  with an empty query.
- **Re-focus from Active** — `/` in Active re-focuses with the
  existing query intact.

### `internal/ui/msglist_test.go` (additions)

- **Filter matches subject** — thread kept.
- **Filter matches sender** — thread kept.
- **Filter matches date text under `[all]` only** — correct gating
  (date text matches under `[all]`, not under `[name]`).
- **Filter with no matches** — empty rows; `FilterResultCount() == 0`.
- **Filter across thread** — any message matches ⇒ root + all
  children visible regardless of saved fold state.
- **Result count = thread count** — thread with 4 matching replies
  counts as 1.
- **Pre-search cursor saved on first filter** — subsequent
  keystrokes don't overwrite.
- **Clear restores pre-search cursor** — cursor row restored when
  the saved row still exists.
- **Clear with invalid saved cursor clamps to 0** — source shortened
  since save.
- **Filter preserves fold state** — folds restored post-clear.
- **Match mode differentiation** — query "Apr 05" matches under
  `[all]` but not under `[name]` when no subject/sender contains it.
- **Filter is case-insensitive** — uppercase query matches lowercase
  field and vice versa.

### `internal/ui/account_tab_test.go` (additions)

- **`/` activates search from Idle** — `SidebarSearch.State() ==
  Typing` after the keystroke.
- **Printable runes during Typing append to query** — the query
  grows as letters are typed; `msglist.SetFilter` receives the
  updated query on each keystroke.
- **`j/k` during Typing append to query, don't nav** — `j` becomes
  part of the query; message cursor unchanged.
- **`J/K` during Typing append to query, don't nav folders** —
  folder cursor unchanged.
- **All triage keys during Typing append to query** — `d`, `a`,
  `s`, `r`, `f`, `c`, `q`, space, `F`, `U`, `n`, `N` all appear in
  the query string and none of their normal handlers fire.
- **`Enter` during Typing transitions to Active** — state change,
  filter still live, query preserved.
- **`Esc` during Typing clears and returns to Idle** — filter
  cleared, cursor restored.
- **`Esc` during Active clears and returns to Idle** — filter
  cleared, cursor restored.
- **Folder jump (`I/D/S/A`) during Active clears search before
  loading new folder** — `ClearFilter` called before `loadFolderCmd`.
- **Folder jump (`J/K`) during Active clears search** — same.
- **Triage keys during Active act on current row** — normal routing
  resumes.
- **`q` during Active clears, doesn't quit** — `tea.Quit` not
  returned; state → Idle.
- **Fold keys (`Space`, `F`, `U`) during Active are no-ops** — no
  state change; no crash. (They do not append to anything because
  the prompt is unfocused.)
- **`/` in Active re-focuses the prompt with query preserved** —
  Typing state, query text unchanged, cursor at end of query.

## Performance notes

- Filter predicate is `O(messages)` per filter application, which
  fires on each keystroke during Typing. At mock-backend scale
  (<100 messages), this is free. For Pass 3 live-backend scale
  (tens of thousands of messages in a folder), the local filter
  remains viable as long as it runs against what's loaded; backend
  escalation handles deeper searches.
- `filterBuckets` reuses the backing array via `out := buckets[:0]`
  to avoid allocation on every keystroke.
- `strings.ToLower` allocates per field per match attempt; if this
  proves measurable in profiling, a cached lowered-field table on
  `MessageList` is the next step. Not doing it preemptively.

## Risks and open questions

**Risks:**

- `bubbles/textinput` is a new dependency. It's maintained by the
  same org as bubbletea and widely used, so the risk is low, but
  worth mentioning.
- The fold-disabled-during-search silent no-op will surprise users
  who press `Space` expecting to collapse a thread in the filtered
  view. Mitigated later by the toast system (Pass 2.5b-6).
- Pre-search cursor restore semantics could break if `SetMessages`
  is ever called with a completely different folder's data during
  an active search — cursor restore would point at a row index that
  semantically belongs to the old folder. The `AccountTab`
  folder-jump-clears-search contract prevents this, but a bug
  that ignores the contract would produce confusing behavior.
  Guarded by: cursor clamp and test coverage.

**Open questions deferred to Pass 3:**

- Does `query-survives-folder-jump` become the right default once
  backend search exists? Currently ruled out; revisit with real
  usage.
- Does highlight-and-jump become valuable enough to add as a
  config option? Would require softening the "opinionated, not
  configurable in v1" invariant. Revisit with real usage.
- Should `n/N` mean "next/prev page of backend results" once the
  backend paginates? Currently aliases for `j/k`; Pass 3 can
  reinterpret if needed.

## Pass sequencing

This pass is pulled forward from its original `2.5b-7` slot. The
2.5b-3.7 slot (sidebar folder filter) is deleted entirely.

Revised order:

| Pass | Goal | Status |
|------|------|--------|
| 2.5b-3.6 | Threading + fold | done |
| ~~2.5b-3.7~~ | ~~Sidebar filter UI~~ | **removed** |
| 2.5b-7 | Search (this pass) | in progress |
| 2.5b-4 | Prototype: message viewer | next |
| 2.5b-5 | Prototype: help popover | pending |
| 2.5b-6 | Prototype: status/toast system | pending |
| 2.5b-train | Tooling: mailrender training capture | pending (after Pass 3) |
| 3 | Wire prototype to live backend | pending |

2.5b-train blocks on Pass 3 (real messages needed to validate
mailrender). The viewer, help popover, and toast passes stay
in the prototype track against the mock backend.

# UI Code Quality Pass

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Normalize sub-model mutation in `internal/ui/` to the
value-returning pattern used by `Sidebar`/`MessageList`, and apply
small polish items uncovered during the midway code quality review.

**Architecture:** Two phases. Phase A normalizes the `StatusBar`,
`TopLine`, and `Footer` sub-models so every mutation method is a
value receiver that returns the updated value, and `App.Update`
reassigns the result — matching `AccountTab`'s handling of
`Sidebar`/`MessageList`. Phase B applies polish items: a sidebar
header constant, integer formatting, plain-text width measurement,
footer padding without a hot-path `lipgloss.NewStyle()`, and a
Rule-5 exception comment on `AccountTab.backend`. Each phase leaves
the tree green and is committed separately.

**Tech Stack:** Go 1.26, bubbletea, lipgloss, runewidth.

**Review source:** Midway code quality review, 2026-04-12.

---

## File Structure

Modified files:

- `internal/ui/status_bar.go` — `SetCounts`, `SetConnectionState` to value receivers
- `internal/ui/status_bar_test.go` — reassign after setter calls
- `internal/ui/top_line.go` — `SetToast`, `ClearToast` to value receivers
- `internal/ui/top_line_test.go` — reassign after setter calls
- `internal/ui/footer.go` — `SetContext` to value receiver; `View` pads without `lipgloss.NewStyle`
- `internal/ui/footer_test.go` — reassign after `SetContext` calls
- `internal/ui/app.go` — reassign `statusBar`/`topLine`/`footer` after mutation
- `internal/ui/account_tab.go` — `sidebarHeaderRows` constant; Rule-5 exception comment on `backend` field
- `internal/ui/sidebar.go` — `strconv.Itoa` in place of `fmt.Sprintf("%d", ...)`

No new files, no deletions.

---

## Phase A — Sub-Model Receiver Normalization

Every setter on `StatusBar`, `TopLine`, `Footer` becomes a value
receiver returning the updated value. Callers reassign the field.
Tests update in lockstep. The tree stays green across each task —
within a task the setter type changes and its call sites update
together in one edit.

### Task A1: `StatusBar` setters to value receivers

**Files:**
- Modify: `internal/ui/status_bar.go:35-44`
- Modify: `internal/ui/status_bar_test.go` (all setter call sites)
- Modify: `internal/ui/app.go:30,61`

- [ ] **Step 1: Convert the two setters in `status_bar.go`**

Replace lines 35–44 with:

```go
// SetCounts returns a copy of sb with the message and unread counts updated.
func (sb StatusBar) SetCounts(total, unread int) StatusBar {
	sb.total = total
	sb.unread = unread
	return sb
}

// SetConnectionState returns a copy of sb with the connection state set.
func (sb StatusBar) SetConnectionState(state ConnectionState) StatusBar {
	sb.connState = state
	return sb
}
```

- [ ] **Step 2: Update `status_bar_test.go`**

Every current line of the form `sb.SetCounts(10, 3)` becomes
`sb = sb.SetCounts(10, 3)`. Same for `SetConnectionState`. Affected
lines (call sites): 16, 17, 32, 33, 42, 43, 53, 54, 64, 65, 74, 75,
84, 85. Line numbers will shift as you edit — use the test function
names to locate them if needed.

- [ ] **Step 3: Update `app.go` call sites**

At `app.go:30` (inside `NewApp`):

```go
sb := NewStatusBar(styles)
sb = sb.SetConnectionState(Connected)
```

At `app.go:61` (inside `Update`, `FolderChangedMsg` branch):

```go
case FolderChangedMsg:
	m.statusBar = m.statusBar.SetCounts(msg.Exists, msg.Unseen)
	return m, nil
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/ui/...`
Expected: PASS (all existing `StatusBar` tests still pass against
the new reassignment pattern).

### Task A2: `TopLine` setters to value receivers

**Files:**
- Modify: `internal/ui/top_line.go:20-28`
- Modify: `internal/ui/top_line_test.go:52,64,65`

- [ ] **Step 1: Convert the two setters in `top_line.go`**

Replace lines 20–28 with:

```go
// SetToast returns a copy of tl with a toast message set on the right side.
func (tl TopLine) SetToast(msg string) TopLine {
	tl.toast = msg
	return tl
}

// ClearToast returns a copy of tl with the toast message removed.
func (tl TopLine) ClearToast() TopLine {
	tl.toast = ""
	return tl
}
```

- [ ] **Step 2: Update `top_line_test.go`**

Rewrite the three affected lines:

```go
tl = tl.SetToast("✓ 3 archived")      // was tl.SetToast(...)
tl = tl.SetToast("✓ done")            // was tl.SetToast(...)
tl = tl.ClearToast()                  // was tl.ClearToast()
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/ui/...`
Expected: PASS.

Note: `App` has no `TopLine` setter call sites today. The only
references are the constructor and `View`. No `app.go` edit needed
for this task.

### Task A3: `Footer.SetContext` to value receiver

**Files:**
- Modify: `internal/ui/footer.go:108-111`
- Modify: `internal/ui/footer_test.go` (all 11 `SetContext` call sites)

- [ ] **Step 1: Convert the setter in `footer.go`**

Replace lines 108–111 with:

```go
// SetContext returns a copy of f with the displayed keybinding set switched.
func (f Footer) SetContext(ctx FooterContext) Footer {
	f.context = ctx
	return f
}
```

- [ ] **Step 2: Update `footer_test.go`**

Every call `f.SetContext(AccountContext)` becomes
`f = f.SetContext(AccountContext)`. Same for the single
`f.SetContext(ViewerContext)` at line 148. Eleven total call sites
(lines 15, 24, 36, 47, 59, 71, 80, 98, 116, 127, 148).

- [ ] **Step 3: Run tests**

Run: `go test ./internal/ui/...`
Expected: PASS.

Note: `App` has no `Footer.SetContext` call site today. Footer
context switching lives in `footer_test.go` only until the viewer
lands in a later pass. No `app.go` edit needed.

### Task A4: Phase A gate + commit

- [ ] **Step 1: Run the full check**

Run: `make check`
Expected: vet clean, all tests pass.

- [ ] **Step 2: Run `/simplify`**

Invoke the `simplify` skill. Review findings, apply genuine wins
only. Re-run `make check` if anything changed.

- [ ] **Step 3: Commit Phase A**

```bash
git add internal/ui/status_bar.go internal/ui/status_bar_test.go \
        internal/ui/top_line.go internal/ui/top_line_test.go \
        internal/ui/footer.go internal/ui/footer_test.go \
        internal/ui/app.go
git commit -m "$(cat <<'EOF'
Normalize UI sub-model setters to value receivers

Converts StatusBar, TopLine, and Footer mutation methods to value
receivers that return the updated value. Matches the pattern used
by Sidebar and MessageList via AccountTab and removes the paradigm
crack where pointer-receiver setters mutated value-embedded fields
via Go's auto-addressing rule.

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

---

## Phase B — Polish

Five small fixes bundled together. None are architectural; all are
local. One test run and one commit at the end.

### Task B1: `sidebarHeaderRows` constant

**Files:**
- Modify: `internal/ui/account_tab.go:12-14,65`

- [ ] **Step 1: Add the constant**

Replace the `sidebarWidth` block near the top of `account_tab.go`:

```go
// sidebarWidth is the fixed width of the sidebar panel.
const sidebarWidth = 30

// sidebarHeaderRows is the account-name line plus the blank line
// below it, reserved at the top of the sidebar before the folder
// list. AccountTab.View and the sidebar's own sizing both depend on
// this number matching.
const sidebarHeaderRows = 2
```

- [ ] **Step 2: Use it in `updateTab`**

Replace the `-2 for account name + blank line` line at `account_tab.go:65`:

```go
m.sidebar.SetSize(sw, m.height-sidebarHeaderRows)
```

Remove the trailing comment — the constant's own doc carries the
explanation now.

### Task B2: `AccountTab.backend` Rule-5 exception comment

**Files:**
- Modify: `internal/ui/account_tab.go:17-25`

- [ ] **Step 1: Add the comment**

Replace the `AccountTab` struct definition with:

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
	backend mail.Backend
	uiCfg   config.UIConfig
	sidebar Sidebar
	msglist MessageList
	width   int
	height  int
}
```

### Task B3: `strconv.Itoa` in sidebar

**Files:**
- Modify: `internal/ui/sidebar.go` (import block + line 180)

- [ ] **Step 1: Add `strconv` import**

Find the import block at the top of `internal/ui/sidebar.go` and
add `"strconv"` to it if not already present. Keep the existing
import order (stdlib first, then third-party).

- [ ] **Step 2: Replace the Sprintf**

At `sidebar.go:180`:

```go
countStr = applyBg(textStyle, bgStyle).Render(strconv.Itoa(entry.cf.Folder.Unseen))
```

- [ ] **Step 3: Drop `fmt` if no longer used**

Search the file for any other `fmt.` usage. If there is none,
remove `"fmt"` from the import block. If there is, leave it.

Run: `go vet ./internal/ui/...`
Expected: clean. (Will fail with "imported and not used" if `fmt`
became orphaned and wasn't removed.)

### Task B4: `runewidth` on plain status-bar text

**Files:**
- Modify: `internal/ui/status_bar.go` (import block + line 87)

- [ ] **Step 1: Add `runewidth` import**

Add `"github.com/mattn/go-runewidth"` to the import block in
`status_bar.go`. It is already a module dependency (used by
`msglist.go`).

- [ ] **Step 2: Replace the width measurement**

At `status_bar.go:87`:

```go
// Measure right portion width using plain text (no ANSI).
rightPlain := " " + counts + " · " + connIcon + " " + connText + " ─╯"
rightWidth := runewidth.StringWidth(rightPlain)
```

Leave the `lipgloss.Width(result)` call at line 99 untouched —
`result` contains ANSI escapes and needs lipgloss's ANSI-aware
width.

### Task B5: Footer `View` padding without `NewStyle`

**Files:**
- Modify: `internal/ui/footer.go:146-148`

- [ ] **Step 1: Replace the padding line**

At `footer.go:146-148`:

```go
line := " " + strings.Join(parts, sep)
pad := max(0, width-lipgloss.Width(line))
return line + strings.Repeat(" ", pad)
```

Matches the padding approach already used in `app.go:90-91` and
avoids allocating a fresh `lipgloss.Style` on every frame.

### Task B6: Phase B gate + commit

- [ ] **Step 1: Run the full check**

Run: `make check`
Expected: vet clean, all tests pass.

- [ ] **Step 2: Run `/simplify`**

Invoke the `simplify` skill. Apply genuine wins only. Re-run
`make check` if anything changed.

- [ ] **Step 3: Commit Phase B**

```bash
git add internal/ui/account_tab.go internal/ui/sidebar.go \
        internal/ui/status_bar.go internal/ui/footer.go
git commit -m "$(cat <<'EOF'
Polish UI layer per midway code quality review

- Extract sidebarHeaderRows constant; remove magic -2 in AccountTab
- Document AccountTab.backend as the elm-conventions Rule 5 exception
- strconv.Itoa for bare integer in sidebar unread count
- runewidth.StringWidth for plain-text measurement in status bar
- Drop per-frame lipgloss.NewStyle allocation in Footer.View padding

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

---

## Phase C — Install + Smoke

No code changes. Rebuild the installed binary and spot-check that
the visible surfaces still render correctly.

### Task C1: Install

- [ ] **Step 1: Install**

Run: `make install`
Expected: `poplar` written to `~/.local/bin/poplar`.

### Task C2: Smoke check

- [ ] **Step 1: Launch against the mock backend**

Run: `poplar` (the default build runs against the mock backend
when no JMAP credentials are loaded — same as every prior pass).

- [ ] **Step 2: Verify**

- Connection dot (`●` green) renders in the status bar.
- Message count + unread count update when navigating folders
  (`J`/`K`).
- Status bar and top line still render at the correct width with no
  truncation artifacts.
- Footer still pads to full width with no trailing column gap.

If anything looks off, capture the render and treat it as a bug
against Phase A or Phase B before declaring the pass done. This is
a behavior-preserving refactor — the visible output should be
byte-identical to before.

- [ ] **Step 3: Exit cleanly**

Press `q`. No leftover terminal state, no orphaned goroutines.

---

## Self-Review Notes

- **Review coverage:** every Significant and Minor finding from the
  midway review is addressed. The one deferred item is the
  `types.WorkerMessages` global channel in `mailjmap/jmap.go` —
  that's tracked by ADR-0012 and is a multi-account Pass-3 concern,
  not a code quality item to fix here.
- **No placeholders:** every step has exact file paths, exact code,
  and exact commands.
- **Type consistency:** the three setter conversions produce a
  consistent shape (`Set*(...) T { ...; return t }`), and every
  reassignment call site in `App.Update` and the tests follows the
  same `x = x.Set*(...)` form.
- **Commit granularity:** two logical commits (A, B) plus an
  install/smoke phase with no commit. Tests stay green after each
  task.

# Poplar Chrome Redesign: Bottom Frame, Status Indicator, and Layout

**Date:** 2026-04-11
**Status:** Draft
**Parent spec:** `2026-04-09-poplar-design.md`
**Supersedes:** Status bar and footer sections of
`2026-04-10-poplar-ui-wireframing-design.md` and wireframe sections
5 (Message Viewer), 7 (Status Bar), 8 (Command Footer).

## Summary

Redesign the bottom chrome (status bar + command footer), fix tab
bar connection issues, keep the sidebar visible in the message
viewer, and add responsive layout collapse. The status bar becomes
the bottom edge of the frame with a rounded bottom-right corner.
The command footer sits outside the frame. A combined status
indicator on the right shows counts and connection state. Toast
notifications appear on the left.

## Problems with Current Implementation

1. **Tab bar rows 1-2 don't connect** to the right frame edge.
   The bubble ends and nothing extends to the `╮` corner.
2. **Nerd Font icon width mismatch.** `lipgloss.Width("󰇰")`
   returns 1 but the terminal renders it as 2 cells, causing
   row 3 to be 1 column short.
3. **Status bar floats.** The grey bar has no frame integration
   — it terminates abruptly above the footer.
4. **Status bar and footer crowded.** Two dense rows jammed
   together with no visual separation.
5. **Redundant information.** Folder name and icon appear in both
   the active tab and the status bar.
6. **Message viewer hides sidebar.** The current wireframe shows
   the viewer as full-width, but the message body only needs 72
   columns. The sidebar should remain visible.

## Design

### 1. Tab Bar Fixes

The tab bubble's three rows must all span the full terminal width.

**Row 1** (top of bubble): Left padding + `╭───╮` + fill `─` to
right edge. No `╮` corner — row 1 is a connecting line, not a
frame edge.

**Row 2** (content): Inactive tabs + `│ content │` + inactive
tabs + fill spaces to right edge. No right border character.

**Row 3** (opening into content): `─╯` + bubble gap + `╰───┬───╮`.
The `┬` aligns with the panel divider. The `╮` is the top-right
frame corner. This row already works correctly — the issue is
rows 1-2 falling short.

**Icon width fix:** The Nerd Font icon `󰇰` (and similar icons)
reports `lipgloss.Width()` = 1 but renders as 2 cells. Add 1 to
`activeInner` when the icon is present, or measure icon width by
rendering to a test terminal. The pragmatic fix: hardcode icon
display width as 2 in tab width calculations.

### 2. Bottom Frame: Status Bar as Bottom Edge

The grey status bar becomes the bottom edge of the content frame.
The right `│` border curves into the status bar via `─╯`,
connecting horizontally rather than terminating vertically.

**Layout (right to left):**

```
│ content                                                                  │
│                                                                          │
 ◂── grey bg ──────────────────────── 10 messages · 3 unread · ● connected ─╯
 d del  a archive  s star  r reply  R all  f fwd  c compose  / search  ? help
```

- The `╯` is the bottom-right rounded corner of the content frame.
- A short `─` segment connects `╯` to the status indicator text.
- The grey background (`bg_border`) extends from the right edge
  all the way left, forming the bottom of the frame.
- The left edge of the grey bar has no corner or border — it
  starts at column 0, matching the current behavior.
- The command footer sits below, outside the frame, with no
  borders or background. One space of left padding before the
  first hint for visual alignment.

**Status bar no longer shows folder name or icon.** That
information is already in the active tab bubble. The status bar
shows only: message count, unread count, and connection state.

### 3. Combined Status Indicator (Right Side)

The right portion of the status bar is the combined status
indicator. Default content:

```
10 messages · 3 unread · ● connected ─╯
```

**Connection states (colorblind-accessible):**

| State | Icon | Color | Text |
|-------|------|-------|------|
| Connected | `●` (filled) | `color_success` (green) | `connected` |
| Offline | `○` (hollow) | `color_error` (red) | `offline` |
| Reconnecting | `◐` (half) | `color_warning` (orange) | `reconnecting` |

Shape + color + text = triple redundancy. Survives colorblind
conditions, monochrome terminals, and screen readers.

**Active operation states** replace the connection indicator
temporarily:

| State | Display | Behavior |
|-------|---------|----------|
| Searching | `spinner searching...` | Animated bubbletea spinner, persists until done |
| Sending | `spinner sending...` | Animated bubbletea spinner, persists until done |
| Fetching | `spinner loading...` | Animated bubbletea spinner, persists until done |

Use `bubbles/spinner` with the `Dot` style for the animated
indicator. The spinner replaces only the connection portion
(`● connected`), not the message counts.

### 4. Toast Notifications (Left Side)

Transient feedback appears on the left side of the status bar,
keeping the right-side status indicator always visible.

```
 ✓ 3 archived                                      10 messages · 3 unread · ● connected ─╯
```

**Toast behavior:**

- Appears on the left of the grey bar.
- Auto-dismisses after 3 seconds (via `tea.Tick`).
- Left side returns to empty space after dismissal.
- Multiple rapid toasts: latest wins (replaces previous).

**Toast types:**

| Type | Icon | Color | Examples |
|------|------|-------|----------|
| Success | `✓` | `color_success` | "3 archived", "Message sent", "Draft saved" |
| Info | `›` | `accent_tertiary` | "Moved to Trash", "Marked as read" |
| Error | `✗` | `color_error` | "Send failed", "Connection lost" |

**Error toasts persist** until dismissed by keypress or condition
clearing. They do not auto-dismiss.

### 5. Sidebar Always Visible

The sidebar remains visible when viewing a message. The message
viewer renders in the right panel (same position as the message
list) rather than taking over the full content area.

**Width budget:**

| Element | Width |
|---------|-------|
| Sidebar | 30 |
| Divider | 1 |
| Message body | 72 |
| Left margin | 2 |
| Right border | 1 |
| **Total minimum** | **106** |

The message body is hard-wrapped at 72 characters (matching the
compose editor's `textwidth`). Headers may extend wider but are
truncated or wrapped at the panel boundary.

**Benefits:**

- Folder context always visible — can see unread counts change
  in real time.
- Balanced visual weight — sidebar + 72-char body fills the
  screen without the body floating in excessive whitespace.
- Consistent layout — no jarring layout shift when opening or
  closing a message.
- Sidebar navigation still works — can switch folders without
  closing the viewer.

### 6. Responsive Layout

The sidebar collapses when the terminal is too narrow to fit
both panels. Bubbletea sends `tea.WindowSizeMsg` on every resize.

**Breakpoints:**

| Width | Layout |
|-------|--------|
| >= 106 | Sidebar + right panel (message list or viewer) |
| < 106 | Sidebar hidden, right panel gets full width |

**Behavior:**

- Below the breakpoint, the sidebar panel is not rendered and
  the divider is removed. The right panel (message list or
  viewer) uses the full content width.
- The sidebar can still be accessed via `Tab` focus cycling —
  when the sidebar gains focus below the breakpoint, it renders
  as a full-width panel replacing the message list/viewer.
- Above the breakpoint, both panels are always visible. `Tab`
  cycles focus between them.
- The breakpoint is evaluated on every `tea.WindowSizeMsg`.
  Resizing the terminal live transitions smoothly.

### 7. Updated Wireframes

#### Composite Layout (replaces wireframe #1)

```
 ╭───────────╮
 │ 󰇰  Inbox  │
─╯           ╰──────────────────────────────────────────────────────────────────────────────────────────────╮
│←───────── 30 ─────────→│←──────────────────────────── remaining ──────────────────────────────────────→│
│                         │                                                                               │
│ ┃ 󰇰  Inbox           3 │  󰇮  Alice Johnson          Re: Project update for Q2 launch       10:32 AM   │
│   󰏫  Drafts             │  󰇮  Bob Smith               Weekly standup notes                    9:15 AM   │
│   󰑚  Sent               │  󰑚  Carol White             Re: Budget review                     Yesterday   │
│   󰀼  Archive            │      Dave Chen               Meeting minutes from Monday              Apr 07   │
│                         │  󰈻  Eve Martinez            Quarterly report draft                   Apr 06   │
│   󰍷  Spam           12  │      Frank Lee               Re: Server migration plan                Apr 05   │
│   󰩺  Trash              │      ├─ Grace Kim            └─ Re: Server migration plan             Apr 05   │
│                         │      │  └─ Frank Lee            Re: Server migration plan              Apr 05   │
│   󰂚  Notifications      │      Hannah Park             New office supplies order                Apr 04   │
│   󰑴  Remind             │      Ivan Petrov             Conference travel request                Apr 03   │
│   󰡡  Lists/golang       │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
│                         │                                                                               │
 ─────────────────────────┴──────────────────────────────── 10 messages · 3 unread · ● connected ─╯
  d:del  a:archive  s:star  r:reply  R:all  f:fwd  c:compose  /:search  ?:help  ::cmd
```

#### Message Viewer with Sidebar (replaces wireframe #5)

```
 ╭───────────╮ ╭──────────────────────────────────────╮
 │ 󰇰  Inbox  │ │ Re: Project update for Q2 launch     │
─╯           ╰─╯                                      ╰────────────────────────────────────────╮
│                         │                                                                     │
│   󰇰  Inbox           3 │  From:     Alice Johnson <alice@example.com>                        │
│   󰏫  Drafts             │  To:       Geoff Wright <geoff@907.life>                            │
│   󰑚  Sent               │  Date:     Thu, 10 Apr 2026 10:32:07 -0600                         │
│   󰀼  Archive            │  Subject:  Re: Project update for Q2 launch                        │
│                         │  ────────────────────────────────────────────────────────────        │
│   󰍷  Spam           12  │                                                                     │
│   󰩺  Trash              │  Hey Geoff,                                                         │
│                         │                                                                     │
│   󰂚  Notifications      │  Just wanted to follow up on the Q2 launch timeline. I've           │
│   󰑴  Remind             │  attached the updated project plan with the revised milestones.     │
│   󰡡  Lists/golang       │                                                                     │
│                         │  ## Key changes                                                     │
│                         │                                                                     │
│                         │  - Beta release moved to April 15                                   │
│                         │  - QA window extended by one week                                   │
│                         │  - Launch date is now May 1                                         │
│                         │                                                                     │
│                         │  Let me know if you have any concerns about the new timeline.       │
│                         │                                                                     │
│                         │  > On Apr 9, 2026, Geoff Wright wrote:                              │
│                         │  > Can you send me the updated project plan?                        │
│                         │                                                                     │
 ─────────────────────────┴───────────────────────────────────── 100% · ● connected ─╯
  d:del  a:archive  s:star  r:reply  R:all  f:fwd  Tab:links  q:close
```

#### Toast Notification (transient state)

```
 ✓ 3 archived ────────────┴───────────────── 10 messages · 3 unread · ● connected ─╯
  d:del  a:archive  s:star  r:reply  R:all  f:fwd  c:compose  /:search  ?:help  ::cmd
```

#### Responsive: Narrow Terminal (< 106 cols)

```
 ╭───────────╮
 │ 󰇰  Inbox  │
─╯           ╰────────────────────────────────────────────────────────╮
│                                                                     │
│  󰇮  Alice Johnson          Re: Project update for Q2 launch   10:32 AM   │
│  󰇮  Bob Smith               Weekly standup notes                9:15 AM   │
│  󰑚  Carol White             Re: Budget review                  Yesterday   │
│                                                                     │
 ──────────────────────────────── 10 messages · 3 unread · ● connected ─╯
  d:del  a:archive  s:star  r:reply  R:all  f:fwd  c:compose  /:search
```

## Implementation Impact

### Affected Passes

- **2.5b-1 (chrome shell):** Tab bar row 1-2 connection fix,
  icon width fix, bottom frame redesign. Should be updated
  before proceeding with later passes.
- **2.5b-4 (message viewer):** Viewer renders in right panel
  alongside sidebar, not full-width. Significant layout change.
- **2.5b-6 (status/toast system):** Toast on left, spinner for
  active operations. Combined status indicator on right.
- **Wireframes:** Sections 1, 5, 7, 8 updated in this spec.

### New Components

- **StatusIndicator:** Renders the right-side combined indicator
  (counts + connection + spinner). Owns the `bubbles/spinner`
  instance for active operations.
- **Toast:** Renders left-side transient messages with auto-
  dismiss via `tea.Tick`.

### Modified Components

- **TabBar:** Fix rows 1-2 to span full width. Fix icon width
  calculation.
- **StatusBar:** Becomes the bottom frame edge. Integrates
  `StatusIndicator` on the right, `Toast` on the left. Renders
  the `─╯` corner and `┴` divider junction.
- **Footer:** Removed from frame. Renders below the status bar
  with no borders or background.
- **AccountTab:** Sidebar visible in all states (message list
  and viewer). Responsive collapse below 106 columns.
- **App:** `contentHeight` calculation updated — footer is
  outside the frame but still consumes a row.

### Minimum Terminal Size

106 columns for full layout (sidebar + message body). Below that,
sidebar collapses. The application should remain usable down to
~60 columns (message list only, no sidebar, narrower columns).

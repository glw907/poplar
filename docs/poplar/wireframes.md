# Poplar Text Wireframes

Reference wireframes for every UI element in the poplar interface
inventory. Each wireframe defines layout, proportions, and
information density for the bubbletea prototype (Pass 2.5b).

See the UI design spec for the complete interface inventory:
`docs/superpowers/specs/2026-04-10-poplar-ui-wireframing-design.md`

## Conventions

- Box-drawing characters for borders: `╭╮╰╯│─┃`
- `┃` thick left bar for selected row indicator
- Nerd Font glyphs rendered directly
- Color annotations use theme slot names (`accent_primary`, `fg_dim`)
- Default terminal: 120 columns x 40 rows
- `←N→` for column widths
- `[key]` for interactive elements

---

## 1. Composite Layout

Full application with all persistent chrome and both panels visible.
This is the default view on launch — Inbox selected, message list
focused.

```
╭─ 󰇰 Inbox ─────────────────────────────────────────────────────────────────────────────────────────────────────╮
│←────────── 30 ──────────→│←───────────────────────────── remaining ──────────────────────────────────────────→│
│                           │                                                                                   │
│ ┃ 󰇰 Inbox              3 │  󰇮  Alice Johnson            Re: Project update for Q2 launch         10:32 AM   │
│   󰏫 Drafts               │  󰇮  Bob Smith                 Weekly standup notes                      9:15 AM   │
│   󰑚 Sent                 │  󰑚  Carol White               Re: Budget review                       Yesterday   │
│   󰀼 Archive              │      Dave Chen                 Meeting minutes from Monday                Apr 07   │
│                           │  󰈻  Eve Martinez              Quarterly report draft                     Apr 06   │
│   󰍷 Spam             12  │      Frank Lee                 Re: Server migration plan                  Apr 05   │
│   󰩺 Trash                │      ├─ Grace Kim              └─ Re: Server migration plan               Apr 05   │
│                           │      │  └─ Frank Lee              Re: Server migration plan               Apr 05   │
│   󰂚 Notifications        │      Hannah Park               New office supplies order                  Apr 04   │
│   󰑴 Remind               │      Ivan Petrov                Conference travel request                 Apr 03   │
│   󰡡 Lists/golang         │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
│                           │                                                                                   │
├───────────────────────────┴───────────────────────────────────────────────────────────────────────────────────┤
│ 󰇰 Inbox · 10 messages · 2 unread                                                                  ● connected│
├───────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ d:del  a:archive  s:star  r:reply  R:all  f:fwd  c:compose  /:search  ?:help  ::cmd                          │
╰───────────────────────────────────────────────────────────────────────────────────────────────────────────────╯
```

**Annotations:**

- **Tab bar** (row 1): Active tab shows folder icon + name.
  `accent_secondary` text on `bg_base`. Inactive tabs use `fg_dim`
  on `bg_elevated`.
- **Sidebar** (left, 30 cols): Three folder groups separated by
  blank lines. Selected row has `┃` thick left border in
  `accent_primary` + `bg_selection` full-width fill. Unread counts
  right-aligned in `accent_tertiary`, shown only when > 0.
- **Message list** (right, remaining width): Columns — flags (2),
  sender (22), subject (fill), date (12). Double-space separator.
  Unread rows in `accent_tertiary` with bold sender. Read rows in
  `fg_dim`. Thread prefixes use box-drawing `├─ └─ │`.
- **Vertical divider**: `│` between panels in `bg_border`.
- **Status bar**: `fg_bright` on `bg_border`. Folder icon + name,
  message count, unread count. Connection indicator right-aligned.
- **Command footer**: Key in `fg_bright` bold, `:` separator and
  hint text in `fg_dim`. Context = message list.
- **Focus**: Message list focused (sidebar selection shown but
  without the `┃` active border). `Tab` cycles focus between panels.

---

## 2. Tab Bar (#18)

### Single account tab (default on launch)

```
╭─ 󰇰 Inbox ─────────────────────────────────────────────────────────────────────────────────────────────────────╮
```

### Multiple tabs (viewer open)

```
╭─ 󰇰 Inbox ─┬─ Re: Project update for Q2 launch ─┬─────────────────────────────────────────────────────────────╮
```

### Three tabs (two viewers open)

```
╭─ 󰇰 Inbox ─┬─ Re: Project update ─┬─ Budget review ─┬──────────────────────────────────────────────────────────╮
```

**Annotations:**

- **Active tab:** `accent_secondary` text on `bg_base`. Bottom
  border merges with content area (no visible line between tab
  and content).
- **Inactive tabs:** `fg_dim` text on `bg_elevated`. Visible
  bottom border separating tab from content.
- **Tab separator:** `┬` where tab borders meet the top edge.
  `─` fills remaining space to the right edge.
- **Account folder tabs:** Show folder icon + folder name.
  Not closeable — always present. Title updates when folder
  changes (e.g., switch from Inbox to Sent).
- **Viewer tabs:** Show message subject, truncated to fit.
  Closeable with `q`.
- **Numeric switching:** `1-9` keys switch to tab by position.
- **Overflow:** If tabs exceed terminal width, rightmost tabs
  are truncated with `…`. Active tab is always fully visible.

---

## 3. Sidebar (#1 — left panel)

### Focused, Inbox selected

```
 ┃ 󰇰 Inbox              3
   󰏫 Drafts
   󰑚 Sent
   󰀼 Archive

   󰍷 Spam             12
   󰩺 Trash

   󰂚 Notifications
   󰑴 Remind
   󰡡 Lists/golang
   󰡡 Lists/rust
```

### Focused, selection in Disposal group

```
   󰇰 Inbox              3
   󰏫 Drafts
   󰑚 Sent
   󰀼 Archive

   󰍷 Spam             12
 ┃ 󰩺 Trash

   󰂚 Notifications
   󰑴 Remind
   󰡡 Lists/golang
```

### Unfocused (message list has focus)

```
   󰇰 Inbox              3
   󰏫 Drafts
   󰑚 Sent
   󰀼 Archive

   󰍷 Spam             12
   󰩺 Trash

   󰂚 Notifications
   󰑴 Remind
```

The selected folder (Inbox) still has `bg_selection` background
but no `┃` border — the border only appears in the focused state.

**Annotations:**

- **Width:** 30 columns fixed.
- **Selected row (focused):** `┃` thick left border in
  `accent_primary` + full-width `bg_selection` background.
  Folder name in `fg_bright`.
- **Selected row (unfocused):** `bg_selection` background only,
  no `┃` border. Folder name in `fg_base`.
- **Unread counts:** Right-aligned in `accent_tertiary`. Only
  shown when > 0.
- **Folder icons:** Nerd Font in `fg_base`. When folder has
  unread messages, icon switches to `accent_tertiary`.
- **Group spacing:** One blank line between Primary, Disposal,
  and Custom groups. No group headers rendered.
- **Scrolling:** If folders exceed panel height, viewport clips
  with j/k scrolling. No scrollbar.
- **Footer (when focused):**
  `Enter:open  c:compose  ::cmd`

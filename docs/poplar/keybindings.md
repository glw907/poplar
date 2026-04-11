# Poplar Keybinding Map

Authoritative reference for all keybindings. Organized by logical
group. Each binding lists the key, action, and which contexts it
applies to.

Contexts: **G** = global (all views), **L** = message list,
**V** = message viewer, **S** = sidebar.

## Navigation

Movement keys work in the focused panel. In the message list they
move the cursor. In the viewer they scroll. In the sidebar they
move folder selection.

| Key | Action | Context |
|-----|--------|---------|
| `j` | Down / scroll down | G |
| `k` | Up / scroll up | G |
| `G` | Jump to bottom | G |
| `C-d` | Half page down | G |
| `C-u` | Half page up | G |
| `C-f` | Full page down | G |
| `C-b` | Full page up | G |
| `Tab` | Cycle focus (sidebar ↔ right panel) | G |
| `Enter` | Open message / open folder | L, S |

## Triage

Act on the current message (or selection in multi-select mode).

| Key | Action | Context |
|-----|--------|---------|
| `d` | Delete | L, V |
| `a` | Archive | L, V |
| `s` | Star / unstar | L, V |
| `.` | Toggle read / unread | L, V |

## Reply & Compose

| Key | Action | Context |
|-----|--------|---------|
| `r` | Reply | L, V |
| `R` | Reply all | L, V |
| `f` | Forward | L, V |
| `c` | Compose new | G |

## Folder Jump

Single uppercase key jumps to a canonical folder from any context.
Moves the sidebar selection and switches the message list.

| Key | Folder | Context |
|-----|--------|---------|
| `I` | Inbox | G |
| `D` | Drafts | G |
| `S` | Sent | G |
| `A` | Archive | G |
| `X` | Spam | G |
| `T` | Trash | G |

## Search

| Key | Action | Context |
|-----|--------|---------|
| `/` | Start search | L |
| `n` | Next result | L |
| `N` | Previous result | L |

## Select

Multi-select mode for bulk operations.

| Key | Action | Context |
|-----|--------|---------|
| `v` | Enter/exit visual select | L |
| `Space` | Toggle selection on current message | L |

## Viewer

| Key | Action | Context |
|-----|--------|---------|
| `Tab` | Link picker | V |
| `q` | Close viewer, return to list | V |

## App

| Key | Action | Context |
|-----|--------|---------|
| `?` | Help popover | G |
| `:` | Command mode | G |
| `q` | Quit (from message list) | L |
| `C-c` | Force quit | G |

## Footer Display

The command footer shows a curated subset of keybindings relevant
to the current context. Bindings are grouped logically with extra
spacing between groups (4 spaces between groups, 2 within).

Navigation keys (j/k, G, C-d/C-u) are omitted from the footer —
vim users don't need to be told. The footer focuses on email-
specific actions.

### Message list footer

```
 d:del  a:archive  s:star  ┊  r:reply  R:all  f:fwd  c:compose  ┊  /:search  ?:help  ::cmd
 ◂── triage ──────────────▸   ◂── reply/compose ──────────────▸   ◂── app ──────────▸
```

### Viewer footer

```
 d:del  a:archive  s:star  ┊  r:reply  R:all  f:fwd  ┊  Tab:links  q:close  ?:help
 ◂── triage ──────────────▸   ◂── reply ─────────▸      ◂── viewer / app ──────────▸
```

### Sidebar footer

```
 Enter:open  c:compose  ┊  I:inbox  D:drafts  S:sent  A:archive  ┊  ?:help  ::cmd
 ◂── action ──────────▸     ◂── folder jump ──────────────────▸     ◂── app ──────▸
```

### Group separator

`┊` (U+250A, light quadruple dash vertical) rendered in `fg_dim`.
Padded with one space on each side. Reads as a soft divider
without the visual weight of `│`.

## Design Decisions

**Single keys only.** Bubbletea sends one `tea.KeyMsg` per
keypress. No multi-key sequences (no `g i`, `g g`). Folder
jumps use uppercase single keys instead.

**Uppercase for folder jumps.** First letter of canonical folder
name. Avoids conflict with lowercase triage keys (`d` delete vs
`D` Drafts, `a` archive vs `A` Archive, `s` star vs `S` Sent).

**Context-sensitive footer.** The footer changes based on which
panel is focused and whether the viewer is open. Only shows
bindings relevant to the current context.

**Navigation keys silent.** j/k, G, C-d/C-u, C-f/C-b are not
shown in the footer. Vim users know these. The `?` help popover
shows the full reference.

**Group separation via `┊`.** Light quadruple dash vertical in
`fg_dim`, padded with one space on each side. Subtle enough to
recede behind the key hints, clear enough to read the groups.

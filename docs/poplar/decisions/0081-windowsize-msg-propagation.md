---
title: WindowSizeMsg propagation — SetSize plus message forward
status: accepted
date: 2026-04-26
---

## Context

`AccountTab` previously handled `tea.WindowSizeMsg` by calling
`SetSize` on every child and returning. The Pass 4 audit (finding A5)
flagged this as a deviation from the dominant reference-app pattern:
soft-serve, glow, gh-dash, and the bubbletea split-editors example all
both `SetSize` children **and** forward the `WindowSizeMsg` into each
child's `Update`. The reason is that bubbles components (viewport,
textarea, list) reset internal scroll/cursor state on receipt of the
msg; `SetSize` alone does not trigger that reset. wishlist's omission
of the forward is listed as an anti-pattern in the reference-app
survey (§8 avoid #6).

The audit also found A6 — `SidebarSearch.input.Width` was never set,
so long search queries overflowed the sidebar column. The fix is
related: when the parent communicates a new column width, the
embedded textinput needs to know its clip width.

## Decision

Every `tea.WindowSizeMsg` handler in poplar's UI tree:

1. Stores the new dimensions on the model (or shared context).
2. Computes chrome margins and calls `child.SetSize(width-wm, height-hm)`
   on every sized child. `SetSize` sets explicit dimension fields on
   the child (and on any embedded bubbles components — e.g.
   `textinput.Width`, `viewport` reconstruction).
3. **Also forwards** the `WindowSizeMsg` into each child's `Update`,
   batching the resulting Cmds.

`AccountTab.Update` is the canonical example. The forwarded `Update`
calls are no-ops for poplar's current children (viewer reconstructs
its viewport in `SetSize`; sidebar_search's textinput doesn't consume
`WindowSizeMsg`), but the convention is upheld so future bubbles
components added to the tree get their msg without a structural
revisit.

`SidebarSearch.SetSize` clamps `s.input.Width` with a small overhead
reserve for the search-shelf's leading icon and indent.

## Consequences

- **Convention codified.** Future components (a `bubbles/list` for
  folder picking, a `bubbles/textarea` for compose) pick up the msg
  for free.
- **A6 fixed structurally.** Long search queries no longer overflow
  the sidebar column.
- **Tiny per-resize overhead.** The forwarded `Update` calls add one
  type-switch per child per resize event — negligible (resize is rare,
  the type-switch is O(1)).
- **Forecloses "SetSize is enough."** The pattern is intentionally not
  optimized for poplar's current child set; it's optimized for
  resilience as the tree grows.

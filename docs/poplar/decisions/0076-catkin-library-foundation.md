---
title: Catkin library foundation — bubbles/textarea + muesli/reflow
status: accepted
date: 2026-04-25
---

## Context

ADR-0031 names Catkin as poplar's compose editor and requires it
to be a standalone bubbletea component with no poplar dependencies.
ADR-0032 mandates non-modal Ctrl+key commands (always-insert-mode,
pico/micro-shaped). ADR-0033 puts Catkin behind an `Editor`
interface alongside the v1.1 neovim backend. None of those ADRs
named the underlying libraries Catkin would build on, so each time
the question came up we re-researched vimtea, bubbline, etc.

The library landscape (April 2026):

- `charmbracelet/bubbles/textarea` — multi-line buffer with
  unicode, paste, vertical scroll. No soft wrap, no undo, no
  modal layer. Already a poplar dependency.
- `muesli/reflow` — line-wrapping helpers. Already transitively
  in our tree.
- `kujtimiihoxha/vimtea` — modal vim editor sub-model. v0.0.2
  (March 2025), single maintainer, pre-v0.1. Conflicts with
  ADR-0031 (third-party hard dep) and ADR-0032 (modal vim, not
  non-modal Ctrl+key). Rejected.
- `knz/bubbline` — readline-style line editor. Wrong shape for
  email compose.

The Catkin invariants make the decision narrow: any external
editor framework either violates ADR-0031's "no poplar
dependencies" purity (since Catkin itself is the framework) or
violates ADR-0032's UX shape. The remaining question is which
buffer primitive Catkin owns.

## Decision

Catkin builds on two foundation libraries:

- **`charmbracelet/bubbles/textarea`** — buffer, cursor, paste,
  unicode width handling. Catkin wraps it as a sub-model, not
  through inheritance.
- **`muesli/reflow`** — soft-wrap at the configured body width
  (72 cells per the body-width invariant), quoted-reply prefix
  preservation, paragraph re-wrap on edit.

Catkin owns the rest in-package: the Ctrl+key command dispatch
table, undo/redo (snapshot ring buffer over the textarea buffer),
and email-shaped helpers (quote-prefix-aware Enter, smart
paragraph re-wrap). No third-party editor framework is imported.
vimtea has been evaluated and rejected.

## Consequences

- Catkin's `go.mod` will list only `charmbracelet/bubbletea`,
  `charmbracelet/bubbles`, `charmbracelet/lipgloss`, and
  `muesli/reflow`. Extractable as `github.com/glw907/catkin`
  with no poplar coupling.
- Pass 9 starts with library research already settled — the
  pass-3 starter prompt for compose only needs to brainstorm
  Catkin's command vocabulary and the `Editor` interface shape,
  not "what library do we build on."
- ADR-0032 (Catkin Ctrl+key) and ADR-0068 (modifier-free user
  bindings) are in direct conflict as written. ADR-0068 was
  scoped to reading-UI navigation and did not anticipate
  compose. **Carve-out:** ADR-0068's modifier-free rule applies
  to poplar's reading and navigation surfaces. Text-entry
  surfaces (Catkin in v1, neovim-embed in v1.1) are exempt
  because the always-insert-mode shape forces modifier keys —
  bare letters must remain text input. ADR-0068 and ADR-0032 are
  updated with cross-reference notes.
- If vimtea's API stabilizes (post-v1.0) and someone wants modal
  vim in poplar's compose, the path is to add a third `Editor`
  backend per ADR-0033, not to swap Catkin's foundation.

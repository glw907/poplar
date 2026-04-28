---
title: displayCells helper for Nerd Font icon width
status: superseded by 0084
date: 2026-04-26
---

## Context

Pass 3 verification surfaced a real layout defect (BACKLOG #16): the
sidebar columns were 1 cell narrower than the rendered content
because Nerd Font icons in the Supplementary Private Use Area-A
(U+F0000–U+FFFFD) render as 2 terminal cells but `lipgloss.Width`
and the underlying `mattn/go-runewidth` report them as 1. Every
sidebar row, every message-list flag column, and the search shelf's
icon row was systematically off by one. `JoinHorizontal` trusted the
incorrect block widths and the assembled row exceeded the terminal
width, triggering soft-wrap into adjacent panes. The Pass 4 audit
codified this as finding A1.

## Decision

`internal/ui/iconwidth.go` introduces `displayCells(s string) int`,
which returns `lipgloss.Width(s) + spuaACorrection(s)` where
`spuaACorrection` adds one extra cell per SPUA-A rune in `s`.
SPUA-A codepoints are always 4-byte UTF-8 sequences, so
`spuaACorrection` fast-paths plain ASCII via a single byte scan.

`displayCells` is the canonical width measurement for any string in
`internal/ui/` that may contain a Nerd Font icon. The helpers it
replaces:

- `fillRowToWidth` (`internal/ui/styles.go`) — measures the row before
  pad/truncate.
- `Sidebar.renderRow` — measures `leftContent` before computing the
  gap before the unread count.
- `clipPane` (`internal/ui/viewer.go`) — measures each line before
  pad/truncate. Without this, content piped through the viewer that
  contains an icon would re-introduce the bug at clip time.

The message-list flag column is bumped from 1 cell to 2 in
`mlFixedWidth`, and the no-flag case pads to two spaces so the column
is uniform regardless of which icon (or none) renders.

`lipgloss.Width` remains correct for icon-free strings (most width
math: subject text, sender names, ASCII chrome) and is the right
default. Reach for `displayCells` only when measuring a string that
may contain an icon glyph.

## Consequences

- **BACKLOG #16 closed.** Sidebar / msglist / search rows now compute
  to true terminal width; `JoinHorizontal` no longer over-pads.
- **Render hot-path cost is bounded.** ASCII strings hit the fast-path
  byte scan. SPUA-A strings (icon rows only) pay one extra rune walk.
- **Lint surface is small.** No automated rule enforces "use
  `displayCells` for icon strings" — the pattern is rare enough that
  human review catches it, and the conventions doc names the
  invariant explicitly.
- **Forecloses runewidth global config.** Calling
  `runewidth.SetEastAsianWidth(false)` package-wide was rejected
  because it changes behavior for non-Nerd-Font East Asian width
  cases (CJK punctuation) and is a global side effect.

## Superseded

This ADR's premise — "every modern terminal renders SPUA-A as 2 cells"
— was never verified. The fix it claimed to land for BACKLOG #16 was
declared-fixed without user visual confirmation; the same jitter
defect persisted, just inverted. ADR-0084 replaces the static "+1"
correction with a runtime CPR probe. See 0084 for the corrected model.

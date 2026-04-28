---
title: displayCells/displayTruncate everywhere; no lipgloss.Join* on SPUA-A rows
status: narrowed by 0084
date: 2026-04-27
---

## Context

ADR-0079 introduced `displayCells` to fix the SPUA-A 1-cell undercount
in `lipgloss.Width`/`runewidth`. It replaced bare `lipgloss.Width` in
the row-pad/truncate hot path (`fillRowToWidth`, `Sidebar.renderRow`,
`clipPane`). At the time we believed those were the only sites that
mattered.

Pass 4.1 finding F2 surfaced a 1-2 cell jitter at the right border of
the message-list panel: read rows put `│` at col 119, unread rows at
col 121. Tracing it revealed three additional sites that bypass
`displayCells` and re-introduce the undercount:

1. **`msglist.renderRow`** computed `subjectWidth` from `m.width -
   mlFixedWidth` without compensating for the SPUA-A flag glyph; the
   assembled row's `displayCells` was `m.width + spuaACorrection(flag)`,
   then `fillRowToWidth` truncated via `ansi.Truncate` (runewidth-
   based) which clipped at the wrong cell boundary.
2. **`AccountTab.View`** used `lipgloss.JoinHorizontal` to splice
   sidebar + message-list columns; `JoinHorizontal` pads each block
   to the max `lipgloss.Width`, adding spurious cells to SPUA-A rows.
3. **`App.renderFrame`** used `lipgloss.JoinVertical` to assemble
   chrome rows; `JoinVertical` width-pads each row by `lipgloss.Width`
   too.

`ansi.Truncate` itself is also runewidth-based, so any caller relying
on it to clip a SPUA-A-bearing string to N cells produces an off-by-N
result.

## Decision

The discipline expands beyond ADR-0079:

**Width measurement.** Every renderer that handles a string which may
contain a Nerd Font icon uses `displayCells`. `lipgloss.Width` remains
correct only for strings whose contents are knowably icon-free
(plain ASCII chrome, rendered subject text, text-only header values).

**Truncation.** A new helper `displayTruncate(s string, n int) string`
(`internal/ui/iconwidth.go`) wraps `ansi.Truncate` in a loop that
decrements the runewidth limit until `displayCells(result) <= n`. At
most `spuaACorrection(s)` iterations (typically 0–2). Replaces bare
`ansi.Truncate` in `fillRowToWidth` (`styles.go`) and `clipPane`
(`viewer.go`).

**Composition.** `lipgloss.JoinHorizontal` and `lipgloss.JoinVertical`
are forbidden for any composition that may include a row carrying a
SPUA-A glyph. Use direct row-by-row `strings.Join` with
`displayCells`-aware padding instead. Currently:

- `AccountTab.View` joins sidebar + message-list rows with manual
  concatenation; both children are pre-padded to their assigned
  widths via `fillRowToWidth`.
- `App.renderFrame` assembles chrome lines with `strings.Join("\n",
  …)`; each content row is pre-pinned to `m.width - 1` cells via
  `displayCells` + `displayTruncate` before the right border is
  appended.

**Row-builder math in `msglist.renderRow`.** Subject width is computed
as `m.width - mlFixedWidth - spuaACorrection(flag)`, so the assembled
row is exactly `m.width` display cells before `fillRowToWidth` runs.

## Consequences

- BACKLOG #14 (border jitter) closed.
- Two regression tests guard the invariant: `TestMessageList/read_and_
  unread_rows_have_identical_displayCells_width_at_multiple_widths`
  asserts row equality; `TestApp_RightBorderAlignment` asserts every
  `│`-bearing line in `App.View()` has `displayCells == m.width` at
  widths 80/100/120/160.
- The discipline is human-enforced; no lint rule. Pattern is rare
  enough (handful of composition sites in `internal/ui/`) that PR
  review catches it. Future passes that introduce a new SPUA-A-bearing
  panel must use the same row-by-row pattern.
- ADR-0079 is **not** superseded — its core decision (displayCells
  helper for SPUA-A correction) stands. This ADR sharpens the
  discipline by extending it to truncation and composition.
- The `bubbletea-conventions.md` width-math section gains a paragraph
  about `lipgloss.Join*` being unsafe for icon-bearing rows.

## Narrowed by 0084

The discipline now applies only when `spuaCellWidth != 1`. In simple
mode (the default for systems without a Nerd Font installed),
`lipgloss.Width` is canonical and `Join*` is safe. The existing manual
row-by-row join code is kept in this pass — it is correct under both
width regimes — but a future cleanup pass may revert simple-mode call
sites to `Join*`.

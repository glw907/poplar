---
title: Selection background via style composition
status: accepted
date: 2026-04-11  # Pass 2.5b-2
---

## Context

Lipgloss doesn't support layering backgrounds on
already-rendered ANSI text. The alternative — re-rendering each
segment with `style.Background()` — requires every render call
to know about selection state. Passing `bgStyle` as a parameter
keeps selection logic in `View()` and rendering logic in
`renderRow()`.

## Decision

Selected rows apply `bg_selection` by passing a
`bgStyle lipgloss.Style` into `renderRow`. Each text segment
uses `withBg(baseStyle)` to layer the background on top of its
foreground color. Two `bgStyle` variants (plain and selected)
are computed once in `View()` and passed per-row.

## Consequences

No follow-on notes recorded.

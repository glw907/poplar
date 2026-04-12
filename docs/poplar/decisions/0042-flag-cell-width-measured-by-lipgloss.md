---
title: Flag cell width measured by lipgloss, not visual cells
status: accepted
date: 2026-04-11  # Pass 2.5b-3
---

## Context

Initial implementation assumed visual width and
mismatched the math, producing right-edge misalignment between rows
with icons and rows without. Lipgloss is the source of truth for
character cell math because it's what computes every padding and
join in the rendered output. Visual width is only relevant for
human-eye verification of the live render. Documented inline in
`mlFixedWidth` so future contributors don't repeat the mistake.

## Decision

The message list flag column is **1 lipgloss cell wide**,
not 2. The wireframe shows Nerd Font glyphs (`󰇮 󰑚 󰈻`) as 2 visual
cells, but `lipgloss.Width()` reports them as 1 cell. All column-width
math (`mlFixedWidth`, padding) uses lipgloss cells; an empty flag
slot is one space, not two.

## Consequences

No follow-on notes recorded.

---
title: Help popover overlay with dimmed background
status: accepted
date: 2026-04-27
---

## Context

ADR-0071 deferred background dimming of the help popover to post-v1:
"lipgloss has no native opacity, the popover's accent-colored title +
rounded box + centered placement give enough visual distinction, and
ANSI-level color stripping of the underlying view is fragile."

Pass 4.1 user feedback (BACKLOG #14) was that the popover reads as a
*replacement screen* rather than a modal — the underlying view
disappears, breaking the "this is a transient overlay" affordance.

Research finding (2026-04-27): no Charm-ecosystem helper for ANSI
compositing or background dimming exists. `bubbletea` issue #642
(Jan 2023) is unanswered; `lipgloss` v1/v2 and `charmbracelet/x` ship
nothing. Two community patterns exist — vendor `superfile`'s
`PlaceOverlay` algorithm (~80 LOC, MIT) for line-level positional
merge that preserves background ANSI outside the overlay rect; and
"no dim" (every reference Charm app shows the popover atop full-bright
background). Since we want a real dim, we own a small SGR-faint
injector — much simpler than the rewrite-every-foreground-color
approach the plan originally scoped.

## Decision

`App.View`, when `helpOpen` is true, no longer returns
`m.help.View(...)` directly. Instead it:

1. Renders the underlying frame via the same `renderFrame()` helper
   used when help is closed.
2. Passes the frame through `DimANSI` (`internal/ui/dim.go`) which
   prepends `\x1b[2m` and rewrites every existing `\x1b[…m` SGR to
   inject the SGR-faint parameter (`2`), so faint persists through
   style changes and resets become `\x1b[0;2m`.
3. Renders the popover box via `HelpPopover.Box(width, height)` and
   computes its centered position via `HelpPopover.Position(box,
   width, height)`.
4. Composites via `PlaceOverlay(x, y, popover, dimmedBg)`
   (`internal/ui/overlay.go`, vendored from
   `github.com/yorukot/superfile/src/pkg/string_function/overplace.go`,
   MIT, rewritten to use `charmbracelet/x/ansi` instead of
   `muesli/reflow`).

The narrow-terminal fallback — when the popover would not fit — keeps
the previous "Terminal too narrow for help popover" notice. In that
case `App.View` falls back to `HelpPopover.View(...)`'s Place-based
notice over a dimmed bg.

`HelpPopover` gains two new methods:

- `Box(width, height int) string` — renders just the popover box (no
  Place padding).
- `Position(box string, width, height int) (x, y int)` — centered
  position math.

`View(width, height int)` is retained as a standalone fallback (used
by the narrow-terminal path and by tests that exercise centering
geometry without a full App).

## Consequences

- Supersedes ADR-0071's "no background dim in v1" clause. The rest of
  ADR-0071 (App ownership, key stealing, rounded-box-with-embedded-
  title) stands.
- `internal/ui/overlay.go` and `internal/ui/dim.go` are the first
  vendored snippets in `internal/ui/` (the prior vendoring lived in
  `internal/mailauth/`). Both carry top-of-file provenance comments.
- `App.View` short-circuit is no longer "return popover, skip layout"
  — it now always runs `renderFrame` regardless of `helpOpen`. Cost
  is one extra frame render per help-open frame; negligible.
- The dim layer is purely visual; no key routing changes, no state
  changes. Help still steals every key per ADR-0071.
- BACKLOG #14 closed.

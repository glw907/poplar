---
title: Shared `applyBg` helper for row background composition
status: accepted
date: 2026-04-11  # Pass 2.5b-3
---

## Context

The closure body was byte-for-byte identical in both
files. Future row-rendering components (message viewer header,
threaded reply panel) will need the same composition. A free function
is shorter at the call site, eliminates the per-row closure
allocation, and simplifies `renderFlagCell`'s signature (no need to
thread the closure as a parameter).

## Decision

The closure that layers a row's background color onto a
foreground style — previously duplicated as `withBg` inside both
`sidebar.go:renderRow` and `msglist.go:renderRow` — is now a single
package-level helper `applyBg(base, bgStyle)` in `styles.go`. Both
components call `applyBg(s.styles.X, bgStyle).Render(...)` instead of
defining the closure inline.

## Consequences

No follow-on notes recorded.

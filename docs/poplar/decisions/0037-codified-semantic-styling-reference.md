---
title: Codified semantic styling reference
status: accepted
date: 2026-04-11  # Pass 2.5b-2
---

## Context

Repeated "fix this color" iterations kept churning
the same fields in different directions because there was no
single-source-of-truth for which semantic role each slot served.
The doc locks in roles (e.g., "sidebar rows always use `BgElevated`
as their background") independent of the current theme's hex
values. Themes can evolve without scavenging callsites, and color
changes become deliberate edits to a documented contract rather
than ad-hoc tweaks.

## Decision

`docs/poplar/styling.md` is the authoritative map
from every `Styles` field to its palette slot and semantic role.
Before changing any color in `internal/ui/styles.go`, the doc is
updated first. Components never call `lipgloss.NewStyle()` directly
or reach into `CompiledTheme` — they pull from the `Styles` struct.

## Consequences

No follow-on notes recorded.

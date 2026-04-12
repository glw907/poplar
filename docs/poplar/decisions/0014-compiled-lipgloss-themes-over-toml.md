---
title: Compiled lipgloss themes over TOML
status: accepted
date: 2026-04-10  # Pass 2.5-render
---

## Context

Follows Charm conventions (lipgloss styles as Go values).
Eliminates runtime file discovery, TOML parsing errors, and the
glamour→lipgloss impedance mismatch. Three-layer pipeline: filter
(CleanHTML/CleanPlain) → content (ParseBlocks) → renderer (RenderBody
with lipgloss). Poplar and mailrender share the same compiled themes.

## Decision

Themes are compiled Go values (`Palette` → `NewCompiledTheme`
→ `*CompiledTheme` with lipgloss.Style fields), not runtime TOML files.
Glamour dependency removed entirely.

## Consequences

No follow-on notes recorded.

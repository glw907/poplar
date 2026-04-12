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
with lipgloss). (Pre-pivot the themes were also consumed by the
retired mailrender CLI; after ADR 0058 internal/theme/ is poplar-
only and the aerc styleset generator is gone.)

## Decision

Themes are compiled Go values (`Palette` → `NewCompiledTheme`
→ `*CompiledTheme` with lipgloss.Style fields), not runtime TOML files.
Glamour dependency removed entirely.

## Consequences

No follow-on notes recorded.

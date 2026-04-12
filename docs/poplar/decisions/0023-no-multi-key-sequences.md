---
title: No multi-key sequences
status: accepted
date: 2026-04-10  # Pass 2.5b-2
---

## Context

Bubbletea sends one `tea.KeyMsg` per keypress.
Multi-key sequences require a custom state machine (pending key
buffer, timeout logic, disambiguation). This is unnecessary
complexity — folder jumps and other multi-key actions belong in
command mode (`:go inbox`) which naturally handles multi-word
input. `G` (shift-g) for jump-to-bottom is fine since it's a
single keypress.

## Decision

Avoid multi-key chords (e.g., `g i`, `g g`) in
bubbletea. Use single-key bindings for all actions.

## Consequences

No follow-on notes recorded.

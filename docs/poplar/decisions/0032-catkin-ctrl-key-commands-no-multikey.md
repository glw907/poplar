---
title: Catkin: Ctrl+key commands, no multi-key sequences
status: accepted
date: 2026-04-11
---

## Context

Idiomatic bubbletea. Catkin is always in insert mode —
bare keys are text input. The spirit is vim-flavored (efficient,
keyboard-driven, no mouse required) but the grammar is Ctrl+key like
pico/micro. Consistent with poplar's global "no multi-key sequences"
rule.

## Decision

Catkin is non-modal. All commands use modifier keys
(Ctrl+key) or special keys (arrows, Home/End, PgUp/PgDn). No bare
letter commands, no multi-key sequences. One `tea.KeyMsg` = one action.

## Consequences

**Cross-reference 2026-04-25 (ADR-0076):** ADR-0068 (modifier-
free keybindings) post-dates this ADR and contradicts its
Ctrl+key requirement as written. The conflict is resolved by
ADR-0076's carve-out: the modifier-free rule applies to
poplar's reading/navigation surfaces only; Catkin's compose
buffer is exempt because the always-insert-mode shape forces
modifier keys. ADR-0032's Ctrl+key requirement remains the
binding rule for Catkin.

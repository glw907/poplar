---
title: Multi-select deferred to Pass 6; `v`/`Space` reserved
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

Multi-select is a non-trivial feature
(selection state, footer swap, bulk action application) that
belongs with the triage pass where bulk delete/archive
actually matters. Reserving the keys now prevents later passes
from grabbing them for unrelated features. This is also the
one narrow place where poplar accepts modality — `v` enters a
mode where `Space` toggles row selection — and that acceptance
is load-bearing for the keybinding design (e.g., `Space` is
not free for thread-fold toggle, which forces the fold-key
question onto other keys like `Tab`).

## Decision

The `v`-enters-visual-select multi-select design
from `wireframes.md` §16 is deferred to Pass 6 (triage
actions). `v` and `Space` stay in `keybindings.md` as reserved
but marked deferred. Neither key does anything in v1 until
Pass 6 lands.

## Consequences

No follow-on notes recorded.

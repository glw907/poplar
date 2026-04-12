---
title: Thread fold key: `Space`, dual meaning in visual-select mode
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

`Space` has the stronger claim from user
expectation — ranger, nnn, lazygit, k9s, and most
file-manager-adjacent TUIs use `Space` for "fold/toggle
whatever's under the cursor". The earlier reservation of
`Space` exclusively for multi-select (STATUS.md pre-split)
was defensible but not load-bearing: `m`, `x`, or a vim-style
range model would all work for multi-select's per-row toggle,
and none is actively better than `Space`. The "single meaning
per key" rule is about forbidding hidden contextual shifts
*within* a single mode, not about forbidding modes from
changing what keys mean — that's literally what a mode is.
`Tab` was the other serious candidate; it was passed over
because it already means "link picker" in the viewer context,
which creates the same kind of split without the upside of
matching modern-client fold convention.

## Decision

`Space` is the thread fold-toggle key when
outside visual-select mode. Inside visual-select mode (Pass 6)
`Space` retains its reserved role as the row-toggle key. The
two meanings are disambiguated by mode — visual-select already
changes the footer, row highlighting, and the set of "live"
triage keys, so a key meaning different things inside vs.
outside that mode is consistent with the mode design rather
than an exception to it. The two actions don't overlap in
practice either: users don't fold threads while building a
multi-message triage set.

## Consequences

No follow-on notes recorded.

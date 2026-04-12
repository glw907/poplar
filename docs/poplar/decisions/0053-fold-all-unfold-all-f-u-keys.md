---
title: Fold-all / unfold-all: `F` / `U`, Pass 2.5b-3.6
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

Fold-all is the primary bulk action users will
reach for ("walk into a busy folder, collapse everything,
skim roots"). Unfold-all is rarer but cheap to ship once
fold-all exists. Capital letters were chosen to pair with
lowercase `Space` — uppercase single keys are the poplar idiom
for "same action, bigger scope" (cf. `J/K` vs `j/k`, `R` vs
`r`). `Shift-Space` was considered and rejected because many
terminal emulators drop the shift modifier on bare space and
send plain space instead — the keypress is not reliable
cross-platform. Fold-all / unfold-all are a blocker for "index
view done," not polish, which is why 3.6 exists as a dedicated
pass rather than a post-hoc follow-up.

## Decision

`F` (fold-all) and `U` (unfold-all) are the
reserved keys for bulk fold actions. Both ship in Pass
2.5b-3.6 alongside the per-thread `Space` toggle, not before.
Neither is bound in Pass 2.5b-3.5; `keybindings.md` marks them
reserved and points at 3.6 as the delivery pass.

## Consequences

No follow-on notes recorded.

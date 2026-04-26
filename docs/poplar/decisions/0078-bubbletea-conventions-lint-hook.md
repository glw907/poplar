---
title: Bubbletea conventions lint hook
status: accepted
date: 2026-04-26
---

## Context

The Pass 4 audit found 11 deviations from the conventions doc across
`internal/ui/`. Most were structural and mechanically detectable
(`len()` near width math, `ansi.Wordwrap` without `ansi.Hardwrap`,
`MaxWidth` applied to a child's `View()` output, deprecated bubbletea
APIs, `EnterAltScreen` in `Init`). Catching these via human review at
pass-end is fine but late: a mid-pass author who introduces them has
to undo work to fix them. A structural lint that fires at edit time
narrows the gap.

## Decision

`.claude/hooks/bubbletea-conventions-lint.sh` is a `PostToolUse` hook
on Edit/Write, scoped to `internal/ui/**/*.go` and
`internal/content/*.go`. It runs five mechanical checks:

1. Width math via `len()` (compared to or assigned from a
   `width`-named variable on the same line).
2. `ansi.Wordwrap` without `ansi.Hardwrap` in a renderer file.
3. `MaxWidth` applied within 2 lines of a `child.View()` call.
4. Deprecated bubbletea APIs (`HighPerformanceRendering`,
   `tea.Sequentially`, package-level `spinner.Tick()`, `*Model.NewModel`).
5. `EnterAltScreen` / `EnableMouse*` inside `Init`.

The hook is **non-blocking**: warnings go to stderr and the script
exits 0. The hook is a review prompt, not a gate. Semantic checks
(state ownership, message-flow shape, key.Binding usage patterns)
remain a human-review responsibility.

## Consequences

- **Friction at edit time, not at pass-end.** Authors learn the
  conventions while writing the code, not during a separate review.
- **Validated against the existing tree.** The hook fires zero false
  positives on every current `internal/ui/*.go` file and catches both
  expected violations on a synthetic bad fixture.
- **Bounded maintenance.** Each rule is a `grep`/`awk` pipeline with a
  clear owning convention. New rules add one block per finding family;
  noisy rules can be tightened without touching the others.
- **Forecloses regex-everything ambition.** The hook deliberately does
  not try to enforce the planning checklist or message-flow rules.
  Those need a tree-aware tool (e.g. `golangci-lint` custom analyzer);
  this pass scopes the lint to mechanical checks only.

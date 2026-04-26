# Poplar Status

**Current pass:** Pass 2.5b-4.5 shipped 2026-04-25. Audit-1+2
mechanical fixes landed: folder jumps wired, per-folder threading
override consumed, offline color realigned to ADR-0028, filter
package trimmed. No new ADRs — code realigned to existing
invariants. Next is Pass 2.5b-5 (help popover) with a brainstorm
gate on future-binding policy.

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1, 2, 2.5-render, 2.5a | Scaffold, backend, lipgloss, wireframes | done |
| 2.5b-1..3.6, 2.5b-7 | Chrome / sidebar / msglist / threading / search | done |
| 2.5b-4 | Prototype: message viewer | done |
| 2.5b-4.5 | Audit-1+2 mechanical fixes | done |
| 2.5b-5 | Prototype: help popover (open: future-binding policy) | next |
| 2.5b-6 | Prototype: error banner + spinner consolidation | pending |
| 2.5b-train | Tooling: mailrender training capture | pending (after Pass 3) |
| 2.9 | Research: emersion vs aerc fork (BACKLOG #10) | pending |
| 3 | Wire prototype to live backend | pending |
| 6 | Triage actions (bundles toast + undo bar) | pending |
| 8 | Gmail IMAP | pending |
| 9, 9.5 | Compose + send, tidytext in compose | pending |
| 10, 11 | Config, polish | pending |
| 1.1 | Neovim --embed RPC | pending |

## Next starter prompt (Pass 2.5b-5)

> **Goal.** Ship the help popover — modal overlay over the account
> view and the viewer, advertising the keybindings users need to
> remember. Bound to `?`.
>
> **Scope.** Modal overlay infrastructure (the first modal in the
> codebase); two context-specific binding layouts (account vs.
> viewer); key routing so `?` opens it from both contexts and any
> key dismisses it. Layout per `wireframes.md` §5.
>
> **Settled (do not re-brainstorm):** wireframe layout, column
> groupings, modal behavior; `?` as the open key (`app.go` already
> routes to a stub); the two-context split.
>
> **Still open — brainstorm these:**
> - **Future-binding policy.** The wireframe lists keys that don't
>   exist yet (`c compose`, `r reply`, `f forward`, `d delete`, etc.).
>   Three choices, each with a real cost (Audit-3 plan-shape §"Pass
>   2.5b-5"): only-wired (sparse + churns every pass), all-eventual
>   (silent dead keys; toast feedback isn't until 2.5b-6), or
>   all-with-future-marker (discoverable but needs a styling call).
>
> **Approach.** Brainstorm the future-binding policy
> (`superpowers:brainstorming`), write a plan doc at
> `docs/superpowers/plans/YYYY-MM-DD-help-popover.md`, then
> implement. Standard pass-end checklist applies.

## Audits

Done 2026-04-25:
[invariants](audits/2026-04-25-invariants-findings.md) ·
[library packages](audits/2026-04-25-library-packages-findings.md) ·
[plan shape](audits/2026-04-25-plan-shape-findings.md).

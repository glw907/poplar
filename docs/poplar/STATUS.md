# Poplar Status

**Current pass:** Pass 4.1 — render bugfix pass. Plan at
`docs/superpowers/plans/2026-04-27-render-bugfix-pass.md` (7
findings: viewer body overflow, message-list border jitter, column
padding, popover vertical centering, popover overlay + dimmed
background, viewer left padding, search shelf empty-query count).

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1, 2, 2.5-render, 2.5a | Scaffold, backend, lipgloss, wireframes | done |
| 2.5b-1..3.6, 2.5b-7 | Chrome / sidebar / msglist / threading / search | done |
| 2.5b-4 | Prototype: message viewer | done |
| 2.5b-4.5 | Audit-1+2 mechanical fixes | done |
| 2.5b-5 | Prototype: help popover | done |
| 2.5b-6 | Prototype: error banner + spinner consolidation | done |
| 2.9 | Research: emersion vs aerc fork (BACKLOG #10) | done |
| 3 | JMAP direct-on-rockorager + delete fork + wire live | done |
| 4 | Bubbletea conventions audit + infrastructure | done — [audit](audits/2026-04-26-bubbletea-conventions.md) |
| 4.1 | Render bugfix pass — 7 findings, absorbs #14 | next |
| 2.5b-4b | Viewer completion: link picker (`Tab`) + `n/N` filtered (#9) + URL bug cleanup | pending |
| 5 | Bubbletea conventions cleanup: `key.Matches` (#17) + delegation (#18) + App.View trust (#19) | pending |
| 6 | Triage actions (delete/archive/star/read; toast + undo bar) | pending |
| 7 | Polish I — popover narrow-terminal (#15) + small render drift cleanup | pending |
| 8 | Gmail IMAP (direct-on-emersion rewrite) | pending |
| 9 | Compose framing: `Editor` interface, neovim `--embed` adapter, send via go-smtp | pending |
| 9.5 | Compose enhancements: Catkin native editor, tidytext (#12), content cleanup (#13) | pending |
| 10 | Config polish | pending |
| 11 | Final polish + 1.0 prep | pending |
| 2.5b-train | Tooling: mailrender training capture | opportunistic |
| 1.1 | Neovim companion plugin (post-v1, #6) | post-v1 |

## Next starter prompt (Pass 4.1)

> **Goal.** Fix the rendering regressions surfaced after Pass 4 —
> 7 findings catalogued in
> `docs/superpowers/plans/2026-04-27-render-bugfix-pass.md`.
>
> **Scope.** Visible rendering bugs in `internal/ui/`. URL-handling
> bugs deferred to Pass 2.5b-4b. App.View trust refactor (#19)
> stays in Pass 5.
>
> **Settled:** Pass 4 conventions discipline — every fix moves
> closer to `bubbletea-conventions.md`, not away. F3b vendors
> superfile's `PlaceOverlay` (MIT) and adds a small SGR-faint
> injector; this supersedes ADR-0071's "no dim in v1."
>
> **Still open — brainstorm:** none — plan is ready to execute.
>
> **Approach.** Execute findings in plan order (F1 → F2 → F5 → F3
> → F3b → F6 → F4). Standard pass-end checklist applies; expect
> 1–2 ADRs (F2 outcome + F3b overlay/dim).

## Audits

- 2026-04-26: [bubbletea conventions](audits/2026-04-26-bubbletea-conventions.md)
- 2026-04-25: [invariants](audits/2026-04-25-invariants-findings.md) · [library packages](audits/2026-04-25-library-packages-findings.md) · [plan shape](audits/2026-04-25-plan-shape-findings.md)

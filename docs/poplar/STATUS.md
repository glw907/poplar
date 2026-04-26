# Poplar Status

**Current pass:** Pass 4 shipped 2026-04-26. Bubbletea conventions
audit: research-grounded conventions doc, divergence audit of
`internal/ui/`, 8 fix-now findings landed, lint hook installed,
elm-conventions/poplar-pass refreshed. Closes BACKLOG #16; opens
#17/#18/#19.

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1, 2, 2.5-render, 2.5a | Scaffold, backend, lipgloss, wireframes | done |
| 2.5b-1..3.6, 2.5b-7 | Chrome / sidebar / msglist / threading / search | done |
| 2.5b-4 | Prototype: message viewer | done |
| 2.5b-4.5 | Audit-1+2 mechanical fixes | done |
| 2.5b-5 | Prototype: help popover | done |
| 2.5b-6 | Prototype: error banner + spinner consolidation | done |
| 2.5b-train | Tooling: mailrender training capture | pending |
| 2.9 | Research: emersion vs aerc fork (BACKLOG #10) | done |
| 3 | JMAP direct-on-rockorager + delete fork + wire live | done |
| 4 | Bubbletea conventions audit + infrastructure | done — [audit](audits/2026-04-26-bubbletea-conventions.md) |
| 5 | AccountTab + Viewer key.Matches migration (BACKLOG #17) | next |
| 6 | Triage actions (bundles toast + undo bar) | pending |
| 8 | Gmail IMAP (direct-on-emersion rewrite) | pending |
| 9, 9.5 | Compose + send (emersion/go-smtp), tidytext in compose | pending |
| 10, 11 | Config, polish | pending |
| 1.1 | Neovim --embed RPC | pending |

## Next starter prompt (Pass 5)

> **Goal.** Migrate `AccountTab` and `Viewer` key dispatch from
> `switch msg.String()` to `key.Matches`, completing BACKLOG #17.
> Pass 4 established the pattern at the App level (ADR-0080); this
> pass rolls it across the rest of `internal/ui/`.
>
> **Scope.** Introduce `AccountKeys` and `ViewerKeys` structs in
> `internal/ui/keys.go` (parallel to `GlobalKeys`). Wire each
> through its component constructor. Replace every
> `switch msg.String()` chain in `account_tab.go` and `viewer.go`
> with `key.Matches`. Update tests. Help popover advertised keys
> already correspond 1:1 — no help-vocabulary changes.
>
> **Settled:** ADR-0080 dispatch pattern. Modifier-free single
> keys (ADR-0015, 0024, 0051, 0068, 0076).
>
> **Still open — brainstorm:** none — pure structural migration.
>
> **Approach.** Pure implementation pass. Write a plan doc at
> `docs/superpowers/plans/YYYY-MM-DD-key-matches-migration.md`,
> implement, run the standard pass-end checklist. Pass 4's
> bubbletea-conventions-lint hook is now active and will flag
> regressions on Edit/Write.

## Audits

- 2026-04-26: [bubbletea conventions](audits/2026-04-26-bubbletea-conventions.md)
- 2026-04-25: [invariants](audits/2026-04-25-invariants-findings.md) · [library packages](audits/2026-04-25-library-packages-findings.md) · [plan shape](audits/2026-04-25-plan-shape-findings.md)

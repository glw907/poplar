# Poplar Status

**Current pass:** Pass 2.5b-4b next — viewer completion (link picker
`Tab`, filtered `n/N`, URL bug cleanup). Pass 4.1 done — 7 render
fixes landed; ADR-0082 (popover overlay+dim, supersedes 0071) and
ADR-0083 (displayCells everywhere, no JoinHorizontal/Vertical on
SPUA-A rows) added.

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
| 4.1 | Render bugfix pass — 7 findings, absorbs #14 | done |
| 2.5b-4b | Viewer completion: link picker (`Tab`) + `n/N` filtered (#9) + URL bug cleanup | next |
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

## Next starter prompt (Pass 2.5b-4b)

> **Goal.** Complete the viewer: link picker on `Tab`, filtered
> `n/N` (BACKLOG #9), URL-handling cleanup deferred from 4.1.
>
> **Scope.** `internal/ui/viewer.go`, `account_tab.go` for `n/N`,
> `internal/content/{parse,render_footnote}.go` for URL bugs.
>
> **Settled:** ADR-0066/0067 footnote + launcher; ADR-0082 overlay
> (link picker reuses `PlaceOverlay`+`DimANSI`).
>
> **Still open — brainstorm:** link picker layout (column/grid,
> key affordances); `n/N` when no filter active; URL-bug triage.
>
> **Approach.** Brainstorm open questions, write plan at
> `docs/superpowers/plans/YYYY-MM-DD-viewer-completion.md`, then
> implement. Standard pass-end checklist applies.

## Audits

- 2026-04-26: [bubbletea conventions](audits/2026-04-26-bubbletea-conventions.md)
- 2026-04-25: [invariants](audits/2026-04-25-invariants-findings.md) · [library packages](audits/2026-04-25-library-packages-findings.md) · [plan shape](audits/2026-04-25-plan-shape-findings.md)

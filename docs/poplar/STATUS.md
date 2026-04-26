# Poplar Status

**Current pass:** Pass 2.5b-4 shipped 2026-04-25 (ADRs 0065–0069).
All three queued audits ran the same day; their findings drive
Pass 2.5b-4.5's mechanical follow-up and reshaped the upcoming
prototype/research lineup (see plan-shape findings below).

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1, 2, 2.5-render, 2.5a | Scaffold, backend, lipgloss, wireframes | done |
| 2.5b-1..3.6, 2.5b-7 | Chrome / sidebar / msglist / threading / search | done |
| 2.5b-4 | Prototype: message viewer | done |
| 2.5b-4.5 | Audit-1+2 mechanical fixes | next |
| 2.5b-5 | Prototype: help popover (open: future-binding policy) | pending |
| 2.5b-6 | Prototype: error banner + spinner consolidation | pending |
| 2.5b-train | Tooling: mailrender training capture | pending (after Pass 3) |
| 2.9 | Research: emersion vs aerc fork (BACKLOG #10) | pending |
| 3 | Wire prototype to live backend | pending |
| 6 | Triage actions (bundles toast + undo bar) | pending |
| 8 | Gmail IMAP | pending |
| 9, 9.5 | Compose + send, tidytext in compose | pending |
| 10, 11 | Config, polish | pending |
| 1.1 | Neovim --embed RPC | pending |

## Next starter prompt (Pass 2.5b-4.5)

> **Goal.** Land the mechanical follow-up from Audits 1 and 2:
> wire dead folder-jump bindings, consume dead threading config,
> drop the dead `:` binding, settle the offline color, bump
> `go.mod`, drop the orphaned filter reflow family + empty header
> stubs, fix the stale package doc.
>
> **Scope (Audit-1, full detail in
> [`audits/2026-04-25-invariants-findings.md`](audits/2026-04-25-invariants-findings.md)):**
> U3 (delete `Cmd` from `GlobalKeys`), U5 (dispatch `I/D/S/A/X/T`
> in `AccountTab.handleKey`), U6 (consume `fc.Threading` +
> `MessageList.SetThreaded`), U14 (pick offline color: `ColorError`
> vs `FgDim`), B4 (bump `go.mod` to `go 1.26.0`).
>
> **Scope (Audit-2, full detail in
> [`audits/2026-04-25-library-packages-findings.md`](audits/2026-04-25-library-packages-findings.md)):**
> F1 (delete reflow family in `internal/filter/html.go` — keep
> `reOrderedList` which `isShortPlain`/`isCompactLine` still use),
> F2 (delete `filter/headers.go` + test stubs), F3 (rewrite the
> `package filter` doc comment).
>
> **Settled.** Doc-side tightenings from Audit-1 already landed
> (A7, A10, A15, U4, U11). U14 is the only design call left.
> `content/` and `tidy/` follow-ups defer to BACKLOG #11–#13.
>
> **Approach.** Implement directly — no brainstorm. After
> 2.5b-4.5 ships, regenerate the 2.5b-5 (help popover) starter
> prompt — Audit-3 flags an open question on whether the popover
> advertises future-but-unwired bindings. Pass-end via
> `poplar-pass`.

## Audits

All three audits done 2026-04-25. Findings:
[invariants](audits/2026-04-25-invariants-findings.md) ·
[library packages](audits/2026-04-25-library-packages-findings.md) ·
[plan shape](audits/2026-04-25-plan-shape-findings.md).

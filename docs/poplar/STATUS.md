# Poplar Status

**Current pass:** Pass 2.5b-6 shipped 2026-04-25. Error banner +
spinner consolidation: `ErrorMsg{Op, Err}` is the canonical Cmd
error type; `App` owns `lastErr` and renders a foreground-only row
above the status bar (ADR-0073). Shared `NewSpinner(t)` centralizes
the placeholder spinner (ADR-0074). Next is Pass 2.9 (emersion vs
aerc fork research, BACKLOG #10) — must settle the library stack
before Pass 3 wires anything live.

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1, 2, 2.5-render, 2.5a | Scaffold, backend, lipgloss, wireframes | done |
| 2.5b-1..3.6, 2.5b-7 | Chrome / sidebar / msglist / threading / search | done |
| 2.5b-4 | Prototype: message viewer | done |
| 2.5b-4.5 | Audit-1+2 mechanical fixes | done |
| 2.5b-5 | Prototype: help popover | done |
| 2.5b-6 | Prototype: error banner + spinner consolidation | done |
| 2.5b-train | Tooling: mailrender training capture | pending (after Pass 3) |
| 2.9 | Research: emersion vs aerc fork (BACKLOG #10) | next |
| 3 | Wire prototype to live backend | pending (gated on Pass 2.9) |
| 6 | Triage actions (bundles toast + undo bar) | pending |
| 8 | Gmail IMAP | pending |
| 9, 9.5 | Compose + send, tidytext in compose | pending |
| 10, 11 | Config, polish | pending |
| 1.1 | Neovim --embed RPC | pending |

## Next starter prompt (Pass 2.9)

> **Goal.** Decide whether to keep the aerc IMAP+JMAP fork
> (ADR-0058) or migrate to the emersion stack (`go-imap`, `go-smtp`,
> `go-message`, `go-webdav`, `go-vcard`). Output: research doc +
> ADR (potentially superseding 0058) so Pass 3 starts settled.
>
> **Scope.** Research-only — no production code. Throwaway spike
> under `experiments/` allowed if needed for evidence. Deliverables:
> `docs/poplar/research/YYYY-MM-DD-mail-library-stack.md` (library
> inventory, JMAP gap, fork-burden comparison, recommendation) +
> ADR.
>
> **Settled:** v1 backends are Fastmail JMAP + Gmail IMAP
> (ADR-0008/0012); Backend interface is synchronous (ADR-0011).
>
> **Still open — brainstorm:** Go JMAP landscape (BACKLOG #10:
> emersion has no JMAP client — check pkg.go.dev/`jmap`); options
> if thin — (a) drop JMAP, use IMAP for Fastmail (loses push,
> delta sync, atomic ops); (b) hybrid emersion-IMAP + aerc-JMAP;
> (c) find/write a Go JMAP client; aerc-fork maintenance burden
> since 2026-04-09; SMTP/CardDAV future needs (Pass 9, post-1.0).
>
> **Approach.** Brainstorm, write the research doc, land the ADR
> (supersede 0058 in place if migrating).

## Audits

Done 2026-04-25: [invariants](audits/2026-04-25-invariants-findings.md) · [library packages](audits/2026-04-25-library-packages-findings.md) · [plan shape](audits/2026-04-25-plan-shape-findings.md).

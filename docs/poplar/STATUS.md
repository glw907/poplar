# Poplar Status

**Current pass:** Pass 2.5b-7 (sidebar search) shipped 2026-04-13.
ADR 0064 records filter-and-hide semantics with a bottom-pinned
sidebar shelf.

## Passes

| Pass | Goal | Status |
|------|------|--------|
| 1 | Scaffold + Fork | done |
| 2 | Backend Adapter + Connect | done |
| 2.5-render | Lipgloss migration: block model + compiled themes | done |
| 2.5-fix | Fix first-level blockquote wrapping | done |
| 2.5a | Text wireframes for all screens | done |
| 2.5b-1 | Prototype: chrome shell | done |
| 2.5b-keys | Keybinding design | done |
| 2.5b-chrome | Chrome redesign | done |
| 2.5b-2 | Prototype: sidebar | done |
| 2.5b-3 | Prototype: message list | done |
| 2.5b-3.5 | Prototype: UI config + sidebar polish | done |
| 2.5b-3.6 | Prototype: threading + fold | done |
| 2.5b-7 | Prototype: sidebar search | done |
| 2.5b-4 | Prototype: message viewer | next |
| 2.5b-5 | Prototype: help popover | pending |
| 2.5b-6 | Prototype: status/toast system | pending |
| 2.5b-train | Tooling: mailrender training capture system | pending (after Pass 3) |
| 3 | Wire prototype to live backend | pending |
| 6 | Triage actions | pending |
| 8 | Gmail IMAP | pending |
| 9 | Compose + send (Catkin editor) | pending |
| 9.5 | Tidytext in compose | pending |
| 10 | Config | pending |
| 11 | Polish for daily use | pending |
| 1.1 | Neovim embedding (nvim --embed RPC) | pending |

## Next starter prompt (Pass 2.5b-4)

> **Goal.** Message viewer prototype — open a message in the
> right panel (replacing the message list), render header block +
> body via the existing ParseBlocks/RenderBody pipeline, support
> `q` to close.
>
> **Settled.** Sidebar remains visible (ADR 0025); viewer opens
> in the right panel with `q` returning to the list (wireframes
> §4); content pipeline already exists from Pass 2.5-render.
>
> **Still open — brainstorm:** body fetch path (sync vs async
> with spinner); link picker scope (in or out of this pass); auto
> mark-read on open (in or out); interaction with active search
> state if viewer opens into a filtered list.
>
> **Approach.** Brainstorm, write spec + plan under
> `docs/superpowers/{specs,plans}/`, implement via
> `subagent-driven-development`. Pass-end via `poplar-pass`.

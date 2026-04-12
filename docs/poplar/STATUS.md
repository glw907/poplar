# Poplar Status

**Current pass:** Pass 2.5b-3.6 (threading + fold). Pivot plan
(`2026-04-12-poplar-pivot.md`) executing ahead of it — feature work
resumes once the pivot lands.

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
| 2.5b-3.6 | Prototype: threading + fold | pending |
| 2.5b-3.7 | Prototype: sidebar filter UI | pending |
| 2.5b-4 | Prototype: message viewer | pending |
| 2.5b-5 | Prototype: help popover | pending |
| 2.5b-6 | Prototype: status/toast system | pending |
| 2.5b-7 | Prototype: search | pending |
| 3 | Wire prototype to live backend | pending |
| 6 | Triage actions | pending |
| 7 | Search | pending |
| 8 | Gmail IMAP | pending |
| 9 | Compose + send (Catkin editor) | pending |
| 9.5 | Tidytext in compose | pending |
| 10 | Config | pending |
| 11 | Polish for daily use | pending |
| 1.1 | Neovim embedding (nvim --embed RPC) | pending |

## Next starter prompt (Pass 2.5b-3.6)

> **Goal.** Complete the index view with threaded display, per-thread
> fold state, and bulk fold/unfold.
>
> **Scope.** Thread fields on `mail.MessageInfo` (thread id, parent
> ref, depth). Mock backend grows one threaded conversation. Render
> `├─ └─ │` prefixes in the subject column in `FgDim` — document the
> new style slot(s) in `docs/poplar/styling.md` before writing
> renderer code. Per-thread fold state on `MessageList` with `Space`
> (fold-toggle), `F` (fold-all), `U` (unfold-all). `j/k` skips hidden
> children. Collapsed thread shows `[N]` count badge in `fg_dim`
> before the subject. Consume the `[ui.folders.<name>] threading`
> field Pass 2.5b-3.5 parsed. Thread roots sort by the folder's
> existing sort order; children always chronological ascending.
> Footer gains the fold hint; `keybindings.md` promotes
> `Space`/`F`/`U` from reserved to live.
>
> **Settled (do not re-brainstorm):** `Space` fold-toggle outside
> visual mode, row-toggle inside it. `F`/`U` for bulk fold — Shift-
> Space rejected because terminals drop the modifier. No runtime
> threading toggle — config only. Threading default on globally,
> per-folder override. Thread prefixes `├─ └─ │` in `FgDim`.
>
> **Still open — brainstorm these:** sort confirmation; data model
> shape (fields on `MessageInfo` vs sibling `ThreadInfo` type);
> fold state model (per-session vs persisted, default on open);
> thread root identification (earliest vs topmost in sort order);
> mock backend threaded conversation content.
>
> **Approach.** Brainstorm the open questions, write a plan doc at
> `docs/superpowers/plans/2026-04-12-poplar-threading.md`, then
> implement. Pass-end ritual: invoke the `poplar-pass` skill.

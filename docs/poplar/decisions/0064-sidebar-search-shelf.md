---
title: Sidebar search shelf with filter-and-hide semantics
status: accepted
date: 2026-04-13
---

## Context

Pass 2.5b-7 adds message search to the account view. Before this
ADR, `/` was reserved in keybindings but unbound, and wireframe
§7 #15 placed a post-commit search indicator in the status bar
without specifying where the input prompt lived or how thread
semantics interact with filter results. The previous 2.5b-3.7
"sidebar filter UI" pass (folder name filter) was deleted — a
handful of folders doesn't need a find affordance.

Two design axes needed to be settled: (1) where the search UI
lives, and (2) how matches affect the list.

For placement, status-bar prompts fight with existing transient
slots (undo bar, error banner, compose review prompt), and modal
overlays violate the one-pane rule. A 3-row shelf pinned to the
bottom of the sidebar column leaves the status bar free, preserves
the vim convention of "command line at the bottom," and keeps
folders as the primary content in the top of the sidebar.

For filter semantics, highlight-and-jump is vim-coherent but does
not map onto JMAP `Email/query` / IMAP `UID SEARCH` result sets,
which are natively "here is the set of matching messages." A
filter-and-hide shape composes cleanly with the backend search
that Pass 3 will wire — the local filter step becomes a backend
call, and the UI contract stays identical.

## Decision

Poplar ships message search in Pass 2.5b-7 with the following
shape:

- **Placement.** A 3-row shelf pinned to the bottom of the sidebar
  column, below the folder region. Always visible. Idle state
  shows `󰍉 / to search` as a hint; active states show the query
  and result count. The folder region's height subtracts
  `searchShelfRows = 3` from the total sidebar height.

- **Activation.** `/` from the account view (Idle state) enters
  Typing. Pressing `/` again from Active re-focuses the prompt
  with the existing query preserved.

- **Focus model.** Three states: Idle, Typing, Active. Typing is a
  brief modal state — printable runes append to the query,
  `Enter` commits to Active, `Esc` cancels. Active is the stable
  live-filter state — all normal account-view keys route normally.

- **Match semantics.** Filter-and-hide: non-matching threads are
  removed from the display. Thread-level predicate: any message
  in a thread matches → the whole thread is visible (root + all
  children) regardless of saved fold state. Fold state is
  preserved (not mutated) so `Esc` restores the pre-search layout.

- **Match algorithm.** Case-insensitive substring. Two modes:
  `[name]` (subject + sender, default) and `[all]` (subject +
  sender + date text). `Tab` cycles the mode while typing. No
  fuzzy matching, no regex.

- **Scope.** Current folder only. Folder jumps (`I/D/S/A/X/T`,
  `J/K`) clear the active search before loading the new folder.
  `q` is stolen from the quit handler while search is non-idle,
  clearing instead of quitting.

- **Result count.** Thread count (not message count). A thread
  with 4 matching replies counts as 1 result.

- **Backend contract reserved.** Pass 3 wires `backend.Search()`
  behind the same UI. JMAP `Email/query` returns a set of email
  IDs which get rendered through the same filter-and-hide
  pipeline. IMAP requires `UID THREAD REFERENCES` + `UID SEARCH`
  and client-side thread expansion — aerc's forked worker already
  supports both.

## Consequences

- A 3-row shelf becomes permanent sidebar chrome. Folder region
  loses 3 rows of vertical space on every render. The tradeoff
  is discoverability and consistent layout.
- Search is the second narrow modal state in poplar after
  visual-select (Pass 6). The "every key always live, no focus
  cycling" rule gains a documented exception for text input.
- `n/N` are aliased to `j/k` under filter-and-hide. Pass 3 may
  reinterpret `n/N` as "next/prev page of backend results" once
  backend pagination exists — the keys are reserved but not
  load-bearing in this pass.
- Fold keys (`Space/F/U`) become no-ops during Active search.
  This is a silent no-op for Pass 2.5b-7; Pass 2.5b-6 can add a
  toast explaining the behavior.
- `bubbles/textinput` is the first bubbles input component poplar
  imports. Further text-input features (compose, rename folder,
  etc.) should reuse the same library.
- Highlight-and-jump mode and configurable search behavior are
  explicitly deferred. Adding a global config knob for search
  mode would soften the "opinionated and not configurable in v1"
  invariant — that softening requires its own ADR.

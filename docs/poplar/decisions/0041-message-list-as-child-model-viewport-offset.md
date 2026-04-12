---
title: Message list as child model with viewport offset
status: accepted
date: 2026-04-11  # Pass 2.5b-3
---

## Context

Same Elm-architecture pattern as the sidebar (Pass
2.5b-2). Hand-rolled (not `bubbles/list`) for the same reasons: full
control over the `▐` cursor cell, selection background composition,
and hand-tuned column layout. Viewport offset (`clampOffset`) is
needed because the message list scrolls — the sidebar doesn't.
Folder changes via J/K refresh the message list through
`AccountTab.loadSelectedFolder` (mock-backed for the prototype, real
JMAP/IMAP in Pass 3).

## Decision

Message list is a standalone `MessageList` struct in
`internal/ui/msglist.go` with its own `View()`, cursor state
(`selected`), viewport state (`offset`), and movement methods
(`MoveDown`/`MoveUp`/`MoveToTop`/`MoveToBottom`/`HalfPageDown`/`HalfPageUp`/
`PageDown`/`PageUp`). All movement routes through a single `moveBy(delta)`
helper that clamps the cursor and re-clamps the viewport. `AccountTab`
owns it as a child model alongside `Sidebar` and dispatches keys by
identity (`j/k` to msglist, `J/K` to sidebar) — no focus switching.

## Consequences

No follow-on notes recorded.

---
title: One pane, no focus cycling (like pine)
status: accepted
date: 2026-04-11  # Pass 2.5b-2 refinement
---

## Context

Focus cycling is a vestige of the mutt-style two-pane
mental model. Pine treats the screen as one thing. With distinct
keybindings (j/k vs J/K), there's no ambiguity — keys dispatch by
identity, not by "which panel is active". Removed the `Panel` type,
`SidebarPanel`/`MsgListPanel` constants, Tab key handler,
`SidebarContext`/`SidebarKeys`, and the sidebar `focused` field.
The `┃` selection indicator always renders on the selected folder.
Simplifies both the mental model and the code.

## Decision

The account view is a single pane from a keyboard nav
standpoint. No Tab focus cycling between sidebar and message list.
Every key is always live: `j/k` navigate messages, `J/K` navigate
folders, triage/reply keys act on the current message. The footer
has a single `AccountContext` — it only changes when the viewer
opens over the list.

## Consequences

No follow-on notes recorded.

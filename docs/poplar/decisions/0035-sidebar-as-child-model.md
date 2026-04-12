---
title: Sidebar as child model, not inline rendering
status: accepted
date: 2026-04-11  # Pass 2.5b-2
---

## Context

The chrome shell pass (2.5b-1) had inline sidebar
rendering in `AccountTab.renderSidebar`. Extracting to a child model
follows Elm architecture (each component owns its state and view),
makes the sidebar independently testable, and prepares for the
message list component to follow the same pattern.

## Decision

Sidebar is a standalone `Sidebar` struct in
`internal/ui/sidebar.go` with its own `View()`, navigation methods,
and style-aware rendering. `AccountTab` owns it as a child model
and delegates key events via `handleSidebarKey`.

## Consequences

No follow-on notes recorded.

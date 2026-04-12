---
title: Status indicator for transient feedback
status: accepted
date: 2026-04-09
---

## Context

Without feedback, destructive or async actions feel
uncertain — did the send actually go through? A brief, auto-dismissing
status message confirms the action completed. Covers sends, drafts,
deletes, archive, moves, errors, and connection state changes.

## Decision

A status area displays transient messages for user
actions: "Message sent", "Draft saved", "Deleted 3 messages", etc.
Positioned above the command footer (or integrated into the tab bar
area) so it doesn't displace the persistent command hints.

## Consequences

**Implementation:** Open question — could be an inline status bar
region, a toast-style overlay, or a brief modal. Decide during
implementation based on what looks best in bubbletea. Modern TUI
convention leans toward toast overlays (bottom-right, auto-dismiss
after 2-3s) but the right choice depends on how the layout feels.

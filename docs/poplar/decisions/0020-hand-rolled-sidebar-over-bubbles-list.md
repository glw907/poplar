---
title: Hand-rolled sidebar over bubbles/list
status: accepted
date: 2026-04-10
---

## Context

`bubbles/list` lacks native section/group support.
Hand-rolled is the idiomatic approach — Charm's own apps use the
same technique for grouped sidebars. Allows full control over
selection styling (left thick border + background fill), unread
count badge alignment, and group spacing.

## Decision

Sidebar uses `lipgloss.JoinVertical` with custom
row rendering, not `bubbles/list`.

## Consequences

No follow-on notes recorded.

---
title: Inline compose over terminal takeover
status: accepted
date: 2026-04-11
---

## Context

Every terminal email client today shells out and loses
context. Keeping the chrome visible during compose is the differentiating
UX feature. The header region is native bubbletea (not part of the
editor), the editor fills the remaining space.

## Decision

Compose renders in the right panel. Sidebar, top line,
status bar, and footer remain visible and active with compose-appropriate
content. No `tea.ExecProcess` terminal takeover.

## Consequences

No follow-on notes recorded.

---
title: External editor only
status: accepted
date: 2026-04-09
---

## Context

Building an inline editor in bubbletea is a massive
effort for marginal benefit. nvim-mail already provides the exact
compose UX we want. Simplifies the UI significantly.

## Decision

No built-in compose editor. Always launch `$EDITOR`
(nvim-mail) via `tea.ExecProcess`.

## Consequences

No follow-on notes recorded.

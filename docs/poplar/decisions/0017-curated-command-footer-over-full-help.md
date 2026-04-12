---
title: Curated command footer over full help
status: accepted
date: 2026-04-10
---

## Context

Vim users don't need to be told about j/k. Footer
real estate is precious — use it for email workflow discovery. The
`?` popover provides the complete reference when needed.

## Decision

Footer shows only email-specific keybindings (triage,
reply, compose, search, command). Vim navigation and thread folding
are silent. Full reference via `?` popover.

## Consequences

No follow-on notes recorded.

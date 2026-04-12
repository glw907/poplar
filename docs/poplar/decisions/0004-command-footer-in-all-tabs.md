---
title: Command footer in all tabs
status: accepted
date: 2026-04-09
---

## Context

Terminal email clients have steep learning curves.
A visible command reference eliminates guesswork without requiring
a manual. Grouping by function (navigation, triage, compose, search)
makes the footer scannable. The footer content changes per tab context
(message list commands differ from viewer commands).

## Decision

Every tab displays a persistent footer showing available
commands grouped by function. Format: `c:compose  j/k:move  d:delete`
etc. Poplar is opinionated about keybindings — the footer is the
primary discoverability mechanism, not a help page.

## Consequences

No follow-on notes recorded.

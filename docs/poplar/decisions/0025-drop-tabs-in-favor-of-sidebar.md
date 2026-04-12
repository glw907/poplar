---
title: Drop tabs in favor of sidebar
status: accepted
date: 2026-04-11
---

## Context

With the sidebar always visible, the tab bar
provided no new information while consuming 3 rows. Simplifies
navigation — no tab lifecycle, no `1-9` switching. Aligns with
"Better Pine" philosophy (one thing at a time).

## Decision

Remove the tab bar entirely. The sidebar (always
visible) shows folder context. Opening a message renders in the
right panel, not a new tab. `q` returns to the message list.

## Consequences

No follow-on notes recorded.

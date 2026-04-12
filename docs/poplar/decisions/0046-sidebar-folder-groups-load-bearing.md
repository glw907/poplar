---
title: Sidebar folder groups are load-bearing
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

The grouping prevents accidental navigation
into personal folders when scrolling past Trash, and the
wireframe looks good because of it. Letting users flatten the
groups would allow them to shoot off their own feet for no
clear win. Keeping the groups rigid doesn't cost flexibility
that actually matters — within-group ranking covers every
real reorder use case (pin `Lists/golang` to the top of Custom,
push `Notifications` down, etc.). The simpler invariant also
makes the sidebar renderer easier to reason about.

## Decision

The Primary / Disposal / Custom three-group
structure of the sidebar is permanent, not a default the user
can flatten. Poplar always renders Primary first, then
Disposal, then Custom, separated by blank lines. User config
assigns an in-group rank to folders but cannot move a folder
across groups. Canonical folders keep their canonical order
unless explicitly reranked. Custom folders alphabetize by
default; user can override with explicit ranks.

## Consequences

No follow-on notes recorded.

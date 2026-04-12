---
title: Nested folders render flat with one-space indent
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

The three-level-deep hierarchies typical of
real Fastmail/Gmail accounts need *some* visual signal that
`Lists/golang` and `Lists/rust` are siblings. A one-space
indent is subtle enough to read as "these things are related"
without implying an interactive tree. Tree view was explicitly
rejected at Pass 2.5b-2 (aerc tried it, its `app/dirtree.go`
sorts children alphabetically ignoring `folders-sort`, and
the complexity-to-benefit ratio is bad). The indent costs
nothing — one character per nested row, no data-model change,
no new navigation rules.

## Decision

Folders whose names contain `/` (e.g.
`Lists/golang`, `Projects/Acme`) get one extra leading space
in the sidebar. No tree view, no expand/collapse. Adjacent
siblings are still kept adjacent by the alphabetical sort
within the Custom group — the indent is pure render polish on
top of the flat data model.

## Consequences

No follow-on notes recorded.

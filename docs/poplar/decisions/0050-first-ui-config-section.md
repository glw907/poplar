---
title: First `[ui]` config section
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

Folder display behavior is UI concern, not
account concern, so it belongs outside the `[[account]]`
block. Keying the per-folder overrides on the folder name
(not a glob) keeps the initial implementation simple — globs
can come later if there's demand. The exact field names and
types are still subject to brainstorm refinement (see
STATUS.md "Still open" list), but the location and shape of
the section is fixed.

## Decision

`~/.config/poplar/accounts.toml` gains a `[ui]`
table with a global `threading` default and
`[ui.folders.<name>]` subsections for per-folder overrides
(threading, sort, rank, possibly hide). This is the first
non-account config section in poplar and sets the pattern for
future UI-tuning sections.

## Consequences

No follow-on notes recorded.

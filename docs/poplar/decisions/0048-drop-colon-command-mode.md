---
title: Drop `:` command mode
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

Every use case for `:` in the wireframes has a
more direct path: folder jumps use single uppercase keys
(`I`/`D`/`S`/`A`/`X`/`T`), move/copy open a folder picker
modal via a key, `/` starts search, `Esc` clears it. A hidden
command layer would double the discoverability surface — users
would have to learn keys *and* command names with no clear
rule for which is authoritative — and it contradicts the
curated-footer philosophy. Pine doesn't have one. Removing it
also frees a full column from the footer on narrow terminals.
If a typed-input affordance turns out to be necessary later
for some action that can't be a key or a modal, we can add it
back — removing a key is cheap, and re-adding one is cheap too.

## Decision

Poplar does not have a `:` command line. Every
action is bound to a key, or is invoked by a key that opens a
modal picker (folder move/copy, search, etc.). Pass 7 in
`STATUS.md` shrinks from "Command mode + search" to just
"Search."

## Consequences

No follow-on notes recorded.

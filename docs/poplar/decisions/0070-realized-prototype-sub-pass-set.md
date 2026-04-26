---
title: Realized prototype sub-pass set (supersedes 0022)
status: accepted
date: 2026-04-25
---

## Context

ADR-0022 (2026-04-10) committed Pass 2.5b to seven per-screen
prototype sub-passes: chrome shell, sidebar, message list, viewer,
help popover, status/toast, command mode. The intent — each screen
as a learning opportunity, incremental validation — has held. The
specific sub-pass set has not.

Audit-3 (plan shape, 2026-04-25) noticed the drift while reviewing
forward sequencing.

## Decision

Pass 2.5b's realized sub-pass set is:

1. **2.5b-1** Chrome shell (done)
2. **2.5b-2** Sidebar (done)
3. **2.5b-3..3.6** Message list + threading (done; threading split
   into its own sub-pass once the screen surfaced it as a distinct
   concern)
4. **2.5b-7** Sidebar search shelf (done; replaced the original
   "command mode" slot after ADR-0024 dropped `:` mode)
5. **2.5b-4** Message viewer (done)
6. **2.5b-4.5** Audit-1+2 mechanical fixes (next; bookkeeping pass
   that didn't exist in 0022's plan)
7. **2.5b-5** Help popover (pending)
8. **2.5b-6** Error banner + spinner consolidation (pending; toast
   + undo bar split out per Audit-3 findings — they bundle into
   Pass 6 with their real consumer)

Command mode is gone (ADR-0024). Threading and search were added.
The numbering sequence is non-monotonic (4 then 4.5, 7 between 3
and 4) because passes were inserted as the screens revealed them.

## Consequences

- ADR-0022 is superseded; its decision text described an aspirational
  set that does not match shipped work.
- Future drift is expected. The per-screen-prototype intent is the
  invariant; the specific sub-pass count is not.
- Pass numbering is allowed to be non-monotonic. Inserting a
  sub-pass between two existing ones (`-4.5`, `-3.6`) is preferred
  over renumbering everything downstream.
- The next sub-pass after 2.5b-6 is Pass 3 (live backend wiring),
  bracketed only by the Pass 2.9 emersion-vs-aerc-fork research
  if its outcome reshapes Pass 3.

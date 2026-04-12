---
title: Runtime threading toggle: dropped
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

Once a user has tuned per-folder threading
preferences, runtime flipping becomes noise — the
Inbox-set-flat user never wants Inbox threaded, the
Archive-set-threaded user never wants Archive flat. YAGNI. If
a compelling runtime use case turns up during real daily use,
adding a key later is cheap. Better Pine means fewer knobs,
not more.

## Decision

Poplar has no single-key runtime threading
toggle (e.g. "flat view just for this session"). Threading is
controlled entirely via config — the global `threading`
default and per-folder `[ui.folders.<name>] threading = false`
overrides.

## Consequences

No follow-on notes recorded.

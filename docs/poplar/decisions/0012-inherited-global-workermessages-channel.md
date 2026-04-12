---
title: Inherited global WorkerMessages channel
status: accepted
date: 2026-04-09  # Pass 2
---

## Context

Refactoring to per-worker channels requires changes
across the entire forked worker codebase. Single-account use (Passes
2-10) is unaffected. Will need to be addressed for multi-account
support in Pass 11.

## Decision

Keep aerc's package-level `types.WorkerMessages` channel
for now. Known limitation: multiple adapters would race on this
channel.

## Consequences

No follow-on notes recorded.

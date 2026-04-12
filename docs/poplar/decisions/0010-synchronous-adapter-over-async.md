---
title: Synchronous adapter over async
status: accepted
date: 2026-04-09  # Pass 2
---

## Context

Bubbletea's `tea.Cmd` model handles async naturally —
blocking calls run in commands that return messages on completion.
Synchronous methods are simpler to reason about and test than
channel-based APIs. The pump goroutine reads from the worker's
response channel and dispatches registered callbacks; `doAction`
blocks on a per-call channel until Done/Error arrives.

## Decision

The `mail.Backend` interface uses synchronous blocking
methods. The JMAP adapter bridges the forked worker's async
message-passing (channels + callbacks) to blocking calls via a pump
goroutine.

## Consequences

No follow-on notes recorded.

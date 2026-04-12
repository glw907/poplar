---
title: Provider folder name normalization
status: accepted
date: 2026-04-10
---

## Context

Users shouldn't see provider implementation details.
Canonical names work across providers and match command mode
jump targets.

## Decision

Poplar displays canonical folder names (Inbox, Sent,
Trash) regardless of provider naming ([Gmail]/Sent Mail, Sent Items,
etc.). Recognition via case-insensitive matching against known
aliases.

## Consequences

No follow-on notes recorded.

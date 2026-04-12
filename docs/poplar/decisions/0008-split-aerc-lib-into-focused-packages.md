---
title: Split aerc lib/ into focused packages
status: accepted
date: 2026-04-09  # Pass 1
---

## Context

aerc's `lib/` is a grab-bag of unrelated utilities.
Splitting by concern makes dependencies explicit and avoids pulling
in UI-specific code (messageview, dirstore, etc.) that lives in the
same package.

## Decision

aerc's monolithic `lib/` package was split into focused
packages: `auth/` (OAuth), `keepalive/` (TCP), `xdg/` (paths),
`log/` (logging), `parse/` (headers).

## Consequences

No follow-on notes recorded.

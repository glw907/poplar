---
title: Clean fork over direct import
status: accepted
date: 2026-04-09
---

## Context

Aerc doesn't maintain a stable library API — internal
packages change without warning. A fork with upstream tracking
(cherry-pick protocol fixes) is more stable than chasing breaking
`go get -u` updates.

## Decision

Fork aerc's worker code rather than importing aerc as
a Go dependency.

## Consequences

No follow-on notes recorded.

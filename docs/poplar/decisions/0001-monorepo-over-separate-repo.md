---
title: Monorepo over separate repo
status: superseded by 0058
date: 2026-04-09
---

## Context

Shares filter, theme, and compose code as library
calls. Avoids cross-repo version coordination during development.
Mailrender can be gradually retired as poplar subsumes its
functionality.

## Decision

Poplar lives in beautiful-aerc, not its own repo.

## Consequences

Superseded 2026-04-12 by ADR 0058 — once poplar subsumed the
filter/theme/content pipelines, the sibling CLIs had no remaining
role and the monorepo's `beautiful-aerc` framing became confusing.
The repo is now `glw907/poplar`, single-binary, single-purpose.

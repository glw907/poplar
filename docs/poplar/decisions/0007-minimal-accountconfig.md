---
title: Minimal AccountConfig
status: accepted
date: 2026-04-09  # Pass 1
---

## Context

Workers only access Name, Source, Params, Folders,
Headers, HeadersExclude, CheckMail, and identity fields. The rest
(PGP, AuthRes, CheckMailCmd, etc.) is unused by IMAP/JMAP workers
or handled differently by poplar. Smaller surface = easier to
maintain and understand.

## Decision

Replace aerc's 50+ field `config.AccountConfig` with
a minimal `poplar.AccountConfig` (15 fields) in `internal/poplar/`.

## Consequences

**Update 2026-04-12 (Pass 2.5b-3.5):**
package moved to `internal/config/` alongside `UIConfig`.

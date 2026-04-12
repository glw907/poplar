---
title: Config in ~/.config/poplar/
status: accepted
date: 2026-04-09  # Pass 2
---

## Context

Allows both clients to coexist during development.
The TOML format matches the theme files for consistency. Credential
resolution uses a `credential-cmd` field that executes a shell
command and injects the output into the source URL's userinfo.

## Decision

Poplar config lives in `~/.config/poplar/accounts.toml`,
separate from aerc's config.

## Consequences

**Update 2026-04-12 (Pass 2.5b-3.5):** file now also carries the
`[ui]` table; parsing split between `ParseAccounts` and `LoadUI`
in `internal/config/`.

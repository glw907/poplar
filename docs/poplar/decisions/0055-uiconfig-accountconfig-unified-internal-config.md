---
title: UIConfig and AccountConfig unified in `internal/config/`
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5
---

## Context

`internal/poplar/` held exactly one type (and its
loader) and was colliding with the `poplar` package-alias name
inside the JMAP adapter. Consolidating both config concerns
under `internal/config/` gives a single clear home for "things
read from the user's config file" and removes the alias-shadow
footgun. Follow-on config sections (keybindings, compose,
themes) will live here too.

## Decision

The `internal/poplar/` package is deleted in Pass
2.5b-3.5. `AccountConfig` moves to `internal/config/account.go`
and a new `UIConfig` + `LoadUI` lives in `internal/config/ui.go`.
Both types read from the same `accounts.toml` file via
independent decodings (`BurntSushi/toml` silently drops unknown
keys, so the two parsers don't collide). The JMAP adapter stub
that used to live in `internal/mail/jmap.go` moves to a new
`internal/mailjmap/` subpackage so `internal/mail` can stop
importing `internal/config` — the writer in `internal/config`
needs to import `internal/mail` for the classifier types, and
a cycle would otherwise form.

## Consequences

No follow-on notes recorded.

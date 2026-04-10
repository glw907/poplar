# Poplar Architecture

Living architecture record for the poplar email client. Updated after
each implementation pass with decisions made and rationale.

## Overview

Poplar is a bubbletea-based terminal email client that lives in the
beautiful-aerc monorepo. It reuses the existing filter, theme, and
compose infrastructure and adds forked aerc worker code for mail
protocol handling (IMAP + JMAP).

| Layer | Package | Purpose |
|-------|---------|---------|
| UI | `internal/ui/` | Bubbletea components (tabs, sidebar, msglist, viewer, statusbar) |
| Mail adapter | `internal/mail/` | `Backend` interface, poplar-native types, account lifecycle |
| Workers | `internal/worker/` | Forked aerc IMAP + JMAP workers |
| Models | `internal/models/` | Forked aerc data types |
| Rendering | `internal/filter/` | Existing HTML/plain/header filters (shared with mailrender) |
| Themes | `internal/theme/` | Existing TOML theme loader (shared with all binaries) |
| Compose | `internal/compose/` | Existing compose buffer normalization (shared with mailrender) |
| Config | `internal/config/` | Poplar config parsing (accounts, UI, keybindings) |

## Key Decisions

### Monorepo over separate repo
**Decision:** Poplar lives in beautiful-aerc, not its own repo.
**Rationale:** Shares filter, theme, and compose code as library
calls. Avoids cross-repo version coordination during development.
Mailrender can be gradually retired as poplar subsumes its
functionality.
**Date:** 2026-04-09

### Clean fork over direct import
**Decision:** Fork aerc's worker code rather than importing aerc as
a Go dependency.
**Rationale:** Aerc doesn't maintain a stable library API — internal
packages change without warning. A fork with upstream tracking
(cherry-pick protocol fixes) is more stable than chasing breaking
`go get -u` updates.
**Date:** 2026-04-09

### External editor only
**Decision:** No built-in compose editor. Always launch `$EDITOR`
(nvim-mail) via `tea.ExecProcess`.
**Rationale:** Building an inline editor in bubbletea is a massive
effort for marginal benefit. nvim-mail already provides the exact
compose UX we want. Simplifies the UI significantly.
**Date:** 2026-04-09

### Lipgloss styles from theme TOML
**Decision:** UI chrome uses lipgloss styles derived from the same
theme TOML files that drive mailrender's ANSI tokens.
**Rationale:** One theme file controls the entire visual experience.
Switching themes changes both message rendering and UI chrome.
**Date:** 2026-04-09

### JMAP + IMAP only (v1)
**Decision:** Target Gmail (IMAP) and Fastmail (JMAP) only. No
maildir, mbox, or notmuch backends.
**Rationale:** Covers the vast majority of real-world users. Smaller
extraction surface from aerc. Backend interface is pluggable — other
backends can be added later if there's demand.
**Date:** 2026-04-09

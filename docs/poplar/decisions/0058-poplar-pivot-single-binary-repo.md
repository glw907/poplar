---
title: Poplar pivot — single-binary, single-purpose repo
status: accepted
date: 2026-04-12
---

## Context

The repo began as `beautiful-aerc`, a monorepo hosting three CLIs
(`mailrender`, `fastmail-cli`, `tidytext`) plus the poplar bubbletea
shell. ADR 0001 justified the monorepo on the basis that mailrender
and poplar would share filter/theme/compose library code while
mailrender was gradually retired. In practice the three CLIs were
scaffolding for the forked aerc worker and the lipgloss theme
pipeline — once those landed inside poplar they had no remaining
role, and their presence diluted the repo's purpose, CLAUDE.md,
and the pass-driven development ritual.

## Decision

Poplar is the repo. The module is `github.com/glw907/poplar`, the
sole binary is `cmd/poplar`, and the repository is `glw907/poplar`
on GitHub. `mailrender`, `fastmail-cli`, and `tidytext` are deleted
entirely — no CLI, no cmd directory, no aerc-styleset output
generator in `internal/theme/`. The aerc-era fork under
`internal/aercfork/` is renamed to `internal/mailworker/` with
provenance comments. Library packages `internal/filter/`,
`internal/content/`, and `internal/tidy/` remain as poplar
dependencies awaiting their in-tree consumers; they do not ship as
standalone binaries. This ADR supersedes ADR 0001.

## Consequences

- `internal/theme/` drops the `GenerateStyleset`/`WriteStyleset`
  generator. Compiled lipgloss themes are the only surface.
- The old `beautiful-aerc` stow package and its aerc config tree
  are removed from `~/.dotfiles/`; the retired
  `~/.claude/docs/{go,elm}-conventions.md` docs are replaced by
  the global `go-conventions` and `elm-conventions` skills.
- CLAUDE.md, the invariants doc, and the pass ritual all target
  poplar only — no scope-sharing with deleted CLIs.
- Future sharing of filter/content/tidy/theme code happens through
  the neovim companion plugin (project memory), not through a
  monorepo with sibling CLIs.

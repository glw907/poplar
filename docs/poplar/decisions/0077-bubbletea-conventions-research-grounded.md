---
title: Bubbletea conventions are research-grounded
status: accepted
date: 2026-04-26
---

## Context

`docs/poplar/bubbletea-conventions.md` was first authored mid-Pass-3
from training memory rather than from a survey of upstream code. The
Pass 3 verification surfaced a real layout bug (BACKLOG #16, Nerd Font
icon width) that exposed the doc's lack of grounding: rules were
plausible but unverified, and there was no path to settle a "what does
the community actually do?" question without re-deriving the answer
each time. Poplar's design vision positions it as a bubbletea
showcase; credibility on that front requires the conventions to
match what Charm libraries and reference apps actually do.

## Decision

The conventions doc is grounded in two cited research artifacts that
ship with the repo:

- `docs/poplar/research/2026-04-26-bubbletea-norms.md` — survey of the
  Charm library source (`bubbles@v1.0.0`, `glamour@v1.0.0`, `lipgloss`,
  `bubbletea@v1.3.10`, `x/ansi`). Every claim cites a file:line in the
  upstream module cache.
- `docs/poplar/research/2026-04-26-reference-apps.md` — survey of
  production bubbletea apps (glow, gum, soft-serve, wishlist, the
  bubbletea/examples directory, gh-dash as a non-Charm community pick).
  Every claim cites a `github.com/<repo>/blob/<tag-or-commit>/...`
  permalink.

Every normative claim in `bubbletea-conventions.md` cites either a
research-doc section or a primary source. When the conventions doc and
the source code (or a reference app) appear to disagree, the research
docs win — they cite the primary source. The conventions doc is the
quick reference; the research docs are the authority of last resort.

## Consequences

- **Audits become possible.** The Pass 4 audit at
  `docs/poplar/audits/2026-04-26-bubbletea-conventions.md` could be
  produced because there was a stable ruler to measure `internal/ui/`
  against.
- **Drift has a fixed cost.** When poplar bumps a Charm library
  major-version, the research docs need a refresh pass; the
  conventions doc may need targeted edits. Both are bounded.
- **Forecloses memory-authored conventions.** Future passes that
  introduce structural rules (e.g. for compose, picker, embedded nvim)
  follow the same pattern: research first, then doc. Memory-only
  rules are not acceptable.

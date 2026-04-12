---
title: Responsive footer with per-hint drop ranks
status: accepted
date: 2026-04-11  # updated 2026-04-12
---

## Context

A pure "hide whole groups" responsive scheme is
too coarse — at borderline widths you'd lose `r/R reply` along
with the rest of the compose group when only `f fwd` needed to
go. Per-hint ranks let the footer degrade by individual
affordance: nav (vim convention) drops first, then niche modes
(`v select`, `n/N results`), then secondary actions (`.`, `s`,
`f`, `/`), keeping the primary email loop (`d`, `a`, `r/R`, `c`)
plus the always-pinned `? help / q quit` escape hatch even at
40 columns. Implemented in `internal/ui/footer.go` —
`fitFooterHints` drops one hint at a time and re-measures until
the rendered plain-text width fits.

## Decision

Footer hints carry per-hint `dropRank` (0–10).
When the terminal is too narrow to fit the full hint list, the
footer progressively drops the highest-rank hints until the
content fits. Rank 0 hints (`? help`, `q quit`) are pinned and
never drop. Groups whose hints all drop also collapse their
preceding `┊` separator.

## Consequences

**Note:** Originally pinned `: cmd` alongside `? help` and
`q quit` in rank 0. Command mode was dropped entirely on
2026-04-12, so the rank-0 set shrank to two hints.

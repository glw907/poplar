---
title: Threading default: on globally, per-folder override
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm
---

## Context

Matches the default state shown in
`wireframes.md` §3, matches the expectation every user brings
from Fastmail web, Apple Mail, and Gmail, and is simpler than
aerc's opt-in model. "Better Pine" means Pine UX polish, not
Pine defaults — pine's defaults lost the war against modern
mail clients, and poplar shouldn't inherit them just for
historical fidelity. The per-folder override is cheap and
covers the "this folder reads better chronologically" case.

## Decision

Threaded display is enabled by default for every
folder. A per-folder `[ui.folders.<name>]` override flips an
individual folder to flat. There is no per-account granularity —
poplar is a single-account-at-a-time UI.

## Consequences

No follow-on notes recorded.

---
title: Pass 2.5b-3.5 split: config/sidebar vs threading/fold
status: accepted
date: 2026-04-12  # Pass 2.5b-3.5 brainstorm, pre-split
---

## Context

The original bundled scope had three unrelated
themes (threading render, config, docs cleanup) held together
only by the fact that the per-folder threading override lives
in the config section. Splitting gives each pass a single
sentence of description ("poplar understands UI config and
renders the sidebar right" vs "poplar threads and folds"),
lets 3.5 ship visible value standalone (real folder hierarchy
rendered correctly), and turns 3.6 from a two-key follow-up
into a proper pass that captures all the work genuinely
entangled with fold (data model, render, navigation, sort,
collapsed badge). The cost — fields parsed in 3.5 that sit
unused for one pass — is a well-established Go pattern and
much smaller than the cost of bundling themes that don't share
rationale.

## Decision

The "threaded view + UI config" pass is split
along its natural seam. Pass 2.5b-3.5 becomes "UI config +
sidebar polish + docs cleanup" — first `[ui]` section, folder
auto-discovery, group classification, within-group ranking,
nested one-space indent, keybindings-doc cleanup. Pass 2.5b-3.6
becomes "threading + fold (index view completion)" — thread
fields on `mail.MessageInfo`, `├─ └─ │` prefixes, per-thread
fold state, `Space`/`F`/`U` keys, `[N]` collapsed badge, sort
interaction. 3.5 parses and stores the threading config fields
but has no consumer until 3.6.

## Consequences

No follow-on notes recorded.

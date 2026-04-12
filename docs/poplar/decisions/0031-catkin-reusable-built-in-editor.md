---
title: Catkin: reusable built-in editor
status: accepted
date: 2026-04-11
---

## Context

A premier bubbletea text editor component that others
can import. Keeps the editor focused on editing, the compose panel
focused on email workflow. Extractable to its own repo if it gains
traction.

## Decision

Poplar's built-in compose editor is named Catkin. It lives
in `catkin/` as a standalone, importable bubbletea component with no
poplar dependencies. Email-specific features (reflow, quote handling,
tidytext, spellcheck) are layered on top by poplar's compose panel.

## Consequences

No follow-on notes recorded.

---
title: appModel wrapper for tea.Model compliance
status: accepted
date: 2026-04-10  # Pass 2.5b-1
---

## Context

Keeps the UI layer free of interface-return overhead
while satisfying bubbletea's `tea.Model` contract at the boundary.
This is the standard pattern in the bubbletea ecosystem for apps
with typed model hierarchies.

## Decision

`ui.App.Update` returns `(App, tea.Cmd)` (typed, per
Elm convention). The `appModel` wrapper in `cmd/poplar/root.go`
satisfies `tea.Model`'s `(tea.Model, tea.Cmd)` return type. Similarly,
`AccountTab` has a public `Update` returning `(tea.Model, tea.Cmd)`
for the `Tab` interface, plus an internal `updateTab` returning
`(AccountTab, tea.Cmd)` for typed access in tests.

## Consequences

No follow-on notes recorded.

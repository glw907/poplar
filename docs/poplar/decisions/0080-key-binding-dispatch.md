---
title: key.Binding declarations and key.Matches dispatch
status: accepted
date: 2026-04-26
---

## Context

Poplar declared a `GlobalKeys` struct with `key.Binding` fields in
`internal/ui/keys.go`, but every actionable-key dispatch in `App`,
`AccountTab`, and `Viewer` used `switch msg.String()`. The `key.Binding`
declarations were decorative. The Pass 4 audit found this in finding
A3 (and again as A11 for the App-level decorative declarations).

The reference-app survey is unanimous: production bubbletea apps
(soft-serve, gh-dash, official examples) use `key.Matches(msg, binding)`
for actionable keys; only glow falls back to `msg.String()` switches and
that is listed as an anti-pattern. `key.Matches` respects each binding's
`Enabled()` flag — disabled bindings never match — making per-state
activation declarative. String switches force inline state checks that
scatter binding logic across the codebase and prevent any future
rebinding feature.

Poplar's keybindings are modifier-free single keys (ADR-0015, 0024,
0051, 0068, 0076). That choice is independent of the declaration form;
`key.Binding` still applies even when the help text is just `"k"`.

## Decision

The dispatch pattern across the bubbletea UI is `key.Matches`. Each
component owning keys declares a `KeyMap` struct of `key.Binding`
values and dispatches via:

```go
switch {
case key.Matches(msg, m.keys.Foo):
    ...
case key.Matches(msg, m.keys.Bar):
    ...
}
```

`GlobalKeys` is split into four bindings to preserve poplar's
context-sensitive `q` behavior (closes viewer / clears search / quits
app) while keeping `Ctrl+C` as an unconditional kill:

```go
type GlobalKeys struct {
    Help      key.Binding  // "?"
    Quit      key.Binding  // "q" — context-sensitive
    ForceQuit key.Binding  // "ctrl+c"
    CloseHelp key.Binding  // "?", "esc"
}
```

The Pass 4 implementation migrates `App.Update` only. The full
migration of `AccountTab` and `Viewer` is logged as BACKLOG #17 and
will land as a dedicated structural-cleanup pass; doing it in Pass 4
would have ballooned scope.

## Consequences

- **Convention established at the root.** `App.Update` is the
  reference site for the `key.Matches` pattern. AccountTab and Viewer
  follow when #17 lands.
- **Per-state activation becomes declarative.** Future state-gated
  bindings (e.g. "compose-only" keys) toggle `binding.SetEnabled(false)`
  rather than adding inline state checks at every dispatch site.
- **Help integration becomes possible.** Even though poplar's help
  popover is a custom modal (ADR-0071), the `key.Binding` declarations
  expose `ShortHelp`/`FullHelp` shapes that a future help refactor
  can consume.
- **Forecloses `msg.String()` for actionable keys.** Text-input
  contexts (compose body, search query) still pass `tea.KeyMsg`
  directly to bubbles components; the rule applies to the
  poplar-owned dispatch chain only.

# Pass <n> Spec — <topic>

A spec captures the design decisions a pass needs to settle before
the plan can be written. Specs are optional for pure
implementation passes but mandatory when the pass introduces a
new component, screen, or significant behavior change.

## Goal

One sentence describing what this pass produces.

## Settled (do not re-brainstorm)

Bulleted list of decisions already made elsewhere that this pass
inherits. Include references — ADR numbers, prior pass docs,
existing invariants.

- ...

## Open questions

List the questions the pass must answer before coding. Each gets
a brief notes block summarizing the brainstorm.

### Q1: <question>

**Notes:** ...

**Decision:** (filled in during brainstorm) ...

### Q2: ...

## Bubbletea conventions (mandatory if the pass touches `internal/ui/`)

Each new or changed component answers:

- **Bubbles analogue.** Which `bubbles` component is the closest
  fit? (viewport, list, table, textinput, textarea, spinner,
  help.) Cite the file in the bubbles module that anchors the
  shape.
- **Deviations.** What — if anything — does this component do
  differently from the bubbles analogue, and why? Each deviation
  becomes an ADR.
- **Size contract.** How does `View()` enforce its assigned width
  and height? Reuse `clipPane`? Use the `viewport.View()` idiom
  (`Width().Height().MaxWidth().MaxHeight().Render()`)? Custom?
- **Message flow.** Which `tea.Msg` types flow up from this
  component to its parent? Which are handled locally?
- **Async I/O.** What `tea.Cmd`s does this component construct?
  Do they return `ErrorMsg` on failure?
- **Keys.** What new bindings does this component introduce?
  Declared in a `KeyMap` struct? Dispatched with `key.Matches`?
  Listed in the help vocabulary per ADR-0072?

If the pass introduces no new component, this section can be
empty — but say so explicitly.

## Acceptance criteria

What must be true for the pass to ship.

- ...

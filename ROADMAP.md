# ROADMAP

> Strategic initiatives. Managed by `/log-project`. Issues tracked in `BACKLOG.md`.

## Active

### The re-founding

The sole active track. Charter:
`docs/superpowers/specs/2026-07-19-poplar-refounding-charter.md`. Its cursor
(phase, next pass, starter prompt) lives in
`docs/superpowers/specs/poplar-refounding-STATUS.md`. Passes 2c and 2b re-plan
under `poplar-pass` before they run.

## Planned

### Bubble Tea refactor pass

Bring the UI to `elm-conventions` as infra sweep 2 left it: one top-level
model, with components as plain structs that have imperative methods
(`elm-conventions` `SKILL.md:36-72`; line 34 routes poplar's migration here).
Its order against 2c's tasks 4 to 6 is decided during poplar development.

- Convert the six nested `tea.Model` types beside `App`: `Confirm`
  (`confirm.go:44-199`), `HelpScreen` (`help.go:118-362`), and the four
  placeholders (`placeholder.go:160-375`).
- Amend ADR-0011:17-19. Its `Screen` embeds `tea.Model` (`registry.go:20-23`)
  and registers into a package-level registry at init, against Rule 1's "no
  package-level mutable variables."
- Re-home two rules from the deleted workstation doc
  `claude/.claude/docs/bubbletea-conventions.md` (removed in dotfiles 7d5bdd0;
  poplar's own copy went in 041eb329):
  - **JoinHorizontal padding** (§2, lines 161-169). `lipgloss.JoinHorizontal`
    right-pads every line to its block's widest line, so a ragged child is the
    child's bug. The generic half now lives in `elm-conventions` Rule 6. The
    pass records the lipgloss fact beside poplar's size contract.
  - **Icon width** (§8, lines 464-475, ADR-0084 on branch `legacy`). Nerd Font
    SPUA glyphs are undercounted, so `JoinHorizontal` is barred on SPUA rows.
    The re-founding design language already forbids private-use glyphs and
    keeps each fallback at its glyph's width
    (`2026-07-27-poplar-design-language.md:409-413`). The pass confirms that
    sentence as the rule's home rather than porting `displayCells`.

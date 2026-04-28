---
title: Three-mode iconography with Nerd-Font autodetection and CPR cell-width probe
status: accepted
date: 2026-04-27
---

## Context

ADR-0079 introduced `displayCells = lipgloss.Width(s) + 1*count(SPUA-A)`
on the premise that "every modern terminal renders SPUA-A as 2 cells."
That premise is false. On the workstation's actual configuration —
kitty + JetBrainsMonoNL Nerd Font via symbol_map — SPUA-A glyphs render
at 1 cell. The +1 has been an unforced overcount since the helper was
introduced; the alignment defect that ADR-0079 claimed to close
(BACKLOG #16) was never visually verified by the user, and the same
right-border jitter that motivated Pass 4.1 F2 is the same defect
inverted.

Empirical research (see brainstorm transcript 2026-04-27): only
~10–15% of fresh-install Linux/macOS users get Nerd Fonts working
zero-config. wezterm/ghostty/kitty bundle Nerd Font fallbacks; gnome-
terminal/Terminal.app/alacritty/foot do not. No mainstream Linux
distro ships a Nerd Font in its default packages. SPUA-A's actual
rendered cell width is determined by the runtime triple of terminal ×
font × symbol_map config — no static policy spans the matrix.

## Decision

Three-mode iconography with autodetection and a runtime probe.

**Config field.** `[ui] icons = "auto" | "simple" | "fancy"`, default
`"auto"`. Validated at load time.

**Mode resolution at startup** (`cmd/poplar/root.go`, before
`tea.NewProgram`):

1. `term.HasNerdFont()` — sysfont enumeration of installed font
   families; substring match for `"nerd font"` or `" nf"` suffix.
2. `term.MeasureSPUACells()` — DSR/CPR probe via `/dev/tty`: write
   `ESC[6n`, record column, write SPUA-A glyph, write `ESC[6n` again,
   compute delta. 200ms timeout; returns 0 on failure.
3. `term.Resolve(cfg, hasNerdFont, probe)` — pure decision returning
   `(IconMode, spuaCellWidth)`. `auto` picks fancy iff `hasNerdFont`,
   else simple. `simple` always returns `(simple, 1)`. `fancy` always
   returns `(fancy, probe)` with fallback to 2 on probe failure.

**Two icon tables.** `internal/ui/icons.go` defines `IconSet` plus
`SimpleIcons` (Unicode Narrow-class — every rune passes
`lipgloss.Width == 1`) and `FancyIcons` (SPUA-A — every rune in
`[U+F0000, U+FFFFD]`). Class invariants enforced by tests.

**Width math.** `displayCells(s) = lipgloss.Width(s) + (spuaCellWidth-1)
* spuaCount(s)`. In simple mode, `spuaCellWidth = 1` and the helper
degenerates to `lipgloss.Width`. The `displayTruncate` loop terminates
in `(spuaCellWidth-1)*spuaCount(s)` iterations — zero in simple mode.

**Composition.** ADR-0083's `lipgloss.JoinHorizontal`/`JoinVertical`
ban remains in effect when `spuaCellWidth != 1`. In simple mode the
restriction is technically lifted, but the existing manual row-by-row
join code in `AccountTab.View` and `App.renderFrame` is kept unchanged
in this pass; it is correct under both width regimes. Reverting to
`Join*` is a future cleanup pass.

**Limitation, called out explicitly.** Font *presence* is a proxy for
"will glyphs render," not a guarantee. A user with a Nerd Font
installed but a terminal configured to use a non-NF font will see
tofu boxes in fancy mode. We cannot disambiguate
tofu-rendered-at-1-cell from narrow-NF-rendered-at-1-cell via DSR
alone — the cursor moves the same amount in both. The `simple`
override exists precisely for this case.

**`poplar diagnose`.** A new cobra subcommand prints the terminal
environment, font-detection result, probe outcome + duration, and the
resolved `(mode, spuaCellWidth, icon_set)` triple. It is the empirical
receipt that prevents the "declared-fixed-without-verification" failure
mode that produced this ADR's predecessor. Used as the gate output for
the manual visual matrix (`docs/poplar/testing/icon-modes.md`).

## Consequences

- **Supersedes ADR-0079.** The "+1 always" rule is wrong. The new helper
  is parameterized on a measured runtime value.
- **Narrows ADR-0083.** The `Join*` ban applies only when
  `spuaCellWidth != 1`. The discipline is preserved at the codebase
  level (no changes to existing manual joins) and the documentation
  reflects the narrower scope.
- **Default works on every fresh install.** No font requirement. Users
  who have a Nerd Font installed get fancy iconography automatically.
- **`make check` enforces both modes.** Width-equality regression tests
  run in `(simple, w=1)`, `(fancy, w=1)`, and `(fancy, w=2)`
  configurations.
- **Manual matrix is required.** No row in `docs/poplar/testing/icon-modes.md`
  may be left unchecked before shipping the pass. The "fix declared
  without visual verification" failure mode that produced ADR-0079 is
  the case study justifying this rule.
- **One new dependency** (`github.com/adrg/sysfont`, MIT, pure-Go),
  one vendored snippet (~50 LOC adapted from
  `github.com/hymkor/go-cursorposition`, MIT, with attribution).

# Icon-Mode Verification Matrix

Manual visual verification gate for ADR-0084. Every row must be
exercised before a pass that touches iconography ships. Rows that
cannot be exercised on the current workstation are marked
`n/a — deferred` with a note.

## Workstation: thinkpad-x1, 2026-04-27

Linux Mint 22.3, kitty + JetBrainsMonoNL Nerd Font via `symbol_map`.
sysfont reports no Nerd Font installed system-wide because the
glyphs are configured at the kitty layer (symbol_map), not as an
installed system font — this is the central case ADR-0084 was
written to handle. Captures taken under the parent terminal via
`tmux new-session -x 120 -y 40` (per `.claude/docs/tmux-testing.md`).

### Row 1: kitty + JetBrainsMonoNL + symbol_map (auto)

`poplar diagnose` output (run in tmux pane on this workstation):

```
Terminal:
  TERM           = tmux-256color
  COLORTERM      = truecolor
  is_terminal    = true

Fonts:
  has_nerd_font  = false
  source         = sysfont

Probe:
󰇮  cell_width     = 1  (0 = probe failed)
  duration       = 500µs

Resolved:
  config.icons   = auto
  effective_mode = simple
  spua_cell_w    = 1
  icon_set       = SimpleIcons
```

Visual verification — `captures/icon-modes-auto-120x40.txt`:

- [x] Right border `│` aligned across all rows (column 120)
- [x] Bottom-right corner `╯` at column 120
- [x] Sidebar folder rows render with simple Unicode icons (▣ ✎ → ▢ ! ✗ ▪ • ◷)
- [x] Message-list flag column uniform width (●, ⚑, ↩ all 1 cell)

Reasoning: sysfont can't see kitty's symbol_map font, so `auto`
correctly falls through to `simple`. Borders align because every
icon is a width-1 Narrow rune.

### Row 2: kitty + Hack Nerd Font (Mono) — auto

`n/a — deferred`. The workstation does not have a Mono Nerd Font
installed at the system layer. To exercise this row, install
`Hack Nerd Font Mono` via `apt install fonts-hack` + nerd-fonts
release, set kitty `font_family Hack Nerd Font Mono`, restart
kitty, re-run `poplar diagnose`. Expect
`has_nerd_font=true`, `cell_width=2`, `effective_mode=fancy`.

### Row 3: kitty + DejaVu Mono (no NF) — auto

`n/a — deferred`. Would require swapping the kitty font and
removing the symbol_map. The auto + no-NF path is already covered
by Row 1 (sysfont reports no NF here; `effective_mode=simple`).

### Row 4: kitty + JetBrainsMonoNL — `simple` forced

Visual verification — `captures/icon-modes-simple-120x40.txt`:

- [x] Right border aligned at column 120
- [x] Identical sidebar/flag glyphs to Row 1 (forcing simple == auto's
  resolution on this workstation)
- [x] No tofu

Confirms the explicit override path resolves to the same `IconSet`
the auto path picked.

### Row 5: kitty + JetBrainsMonoNL + symbol_map — `fancy` forced

Visual verification — `captures/icon-modes-fancy-120x40.txt`:

- [x] Right border aligned at column 120
- [x] Sidebar renders Nerd Font glyphs (󰇰 󰏫 󰑚 󰀼 󰍷 󰩺 󰡡 󰂚 󰑴)
- [x] Message-list flag column renders 󰇮 / 󰈻 / 󰑚 without drift
- [x] No tofu (kitty symbol_map serves the SPUA-A range)

This is the ADR-0084 "honest case" — `fancy` works because the
runtime probe measured `cell_width=1`, so the renderer applies no
adjustment. Pass 4.1's static `+1` correction would have broken
this exact configuration.

### Row 6: gnome-terminal + system default — auto

`n/a — deferred`. gnome-terminal not in active use on this
workstation. Expected outcome: `has_nerd_font=false`,
`effective_mode=simple`, borders aligned, simple icons render.

### Row 7: alacritty + system default — auto

`n/a — deferred`. alacritty not installed. Expected outcome: same
as Row 6.

### Row 8: tmux ⇒ kitty (Row 1 setup) — auto

Implicitly exercised — every capture above was taken via tmux
inside kitty. Borders align in the tmux pane and the diagnose
output confirms `TERM=tmux-256color`. `is_terminal=true` inside
tmux's PTY, probe completes in ~500µs.

## Re-run protocol

After any pass that touches `internal/term/`, `internal/ui/icons*`,
`internal/ui/iconwidth*`, `internal/ui/sidebar*`, `internal/ui/msglist.go`,
or `cmd/poplar/root.go`:

1. `make install`
2. `tmux new-session -d -s poplar-test -x 120 -y 40 'poplar'; sleep 1.5; tmux capture-pane -t poplar-test -p > /tmp/snap.txt; tmux kill-session -t poplar-test`
3. `diff /tmp/snap.txt docs/poplar/testing/captures/icon-modes-auto-120x40.txt`
4. Investigate any drift before shipping.

Workstation-specific captures live under `captures/`. They are
treated as soft golden — a diff means "look at this," not "fail
the build" — because cosmetic touches to fixture data legitimately
change them. Border-column drift is the hard regression to watch.

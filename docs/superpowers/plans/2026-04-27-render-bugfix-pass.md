# Pass 4.1 — Render Bugfix Pass

**Date.** 2026-04-27
**Goal.** Fix the rendering regressions surfaced after Pass 4. All
findings are diff-shaped against the conventions doc — they are
exactly the kind of bug Pass 4's discipline was meant to prevent,
and the fact that they slipped through means the discipline needs
sharpening, not loosening.

## Scope

In: visible rendering bugs in `internal/ui/`. Out: any new feature
work, key migration (Pass 5 territory), App.View trust refactor
(BACKLOG #19 — leave for after), and URL-handling bugs
(several known issues — deferred to the URL footnotes
implementation pass where the rendering and harvesting logic will
be touched together).

This is a pure bugfix pass. No new components, no new bubbles
analogues. The size contract, wordwrap + hardwrap discipline, and
displayCells rules from `docs/poplar/bubbletea-conventions.md` are
the authority — every fix below is a violation of one of them.

## Findings

### F1 — Viewer body overflow corrupts layout (severity: high)

**Symptom.** Opening the Dave Johnson "Re: [asc-membership-committee]"
message at 120×40 produces a frame where body fragments leak into
the sidebar column (`My organization…`, `Thanks,`, `Field Biologist…`,
`omueller@ttcd.org…`, `You received this message…`, `To unsubscribe…`).
Every row below the offending line is shifted; the column divider
breaks. The terminal is wrapping a line that exceeds its width
budget, desyncing the layout.

**Root cause hypothesis.** `content.RenderBodyWithFootnotes` produces
a line wider than `maxBodyWidth` (72) or wider than the viewer's
assigned width, and wordwrap can't break it (long URL, joined token,
or quote prefix `> ` pushes a borderline-fitting line over). The
Pass 4 wordwrap+hardwrap commit (`fe86914`, `09ecd37`) added the
discipline to most call sites; this path slipped through, OR the
hardwrap is applied before the quote-prefix is glued on, so the
prefix re-overflows.

**Fix approach.**
1. Add a regression test that renders the offending message body
   at width=72 and asserts no produced line exceeds 72 cells
   (`displayCells`, not `len`). Capture a fixture from the live
   message body — store under `internal/content/testdata/`.
2. Trace the path from `Viewer.View()` → block renderer → wrap. The
   prefix must be applied *before* hardwrap, so the wrapper sees
   the final line shape. Audit every renderer that returns
   pre-prefixed lines.
3. Apply hardwrap-after-wordwrap to the final composed line, not
   the pre-prefix content.

**Acceptance.** Live tmux capture of the Dave Johnson message at
120×40 shows zero leakage into the sidebar column. The new unit
test passes. Add a property-style test fixture: every message in
`testdata/` rendered at width 40, 72, 100 produces no line wider
than the budget.

### F2 — Message-list flag column right-border jitter (severity: med)

**Symptom.** At 120×40, message list rows ending the same date string
have their right border `│` at different columns. Read rows
(no glyph): border at col 119. Unread rows (envelope `󰇮`): border at
col 120 with a trailing space before it.

**Root cause hypothesis.** Pass 4 audit-A1 added `+1` to SPUA-A
codepoints in `displayCells` to correct undercount. The fix is
correct in principle (these glyphs *are* commonly double-width in
nerd-font terminals) but the message-list row builder appears to
either over-pad or use `lipgloss.Width` somewhere instead of
`displayCells`. Possibilities:
  - `fillRowToWidth` uses `displayCells` for budgeting but the
    row's flag column is built with `lipgloss.Width`, double-counting.
  - The terminal in question (kitty) renders SPUA-A as single-width,
    so the +1 correction is wrong here. Need to confirm with a
    direct probe.

**Fix approach.**
1. Probe the actual rendered width of `󰇮` in this kitty config
   (`printf '\u{F01EE}' | wc -L` won't work — visual cell count via
   tmux capture comparison). Compare to `displayCells` output.
2. If kitty renders 1-cell: the `displayCells` correction is too
   aggressive. The honest fix is to detect this at startup (probe
   the cursor position before/after writing the glyph) — but
   that's heavy. Lighter: configure kitty to render nerd-font SPUA
   as 2-cell (the standard recommendation) and document it as a
   prerequisite. ADR this either way.
3. If kitty renders 2-cell: bug is in the row builder, not
   `displayCells`. Audit `message_list.go`'s row construction for
   `lipgloss.Width`/`len()` usages on glyph-bearing strings.

**Acceptance.** Right border of every message list row aligns at
the same column at multiple widths (80, 100, 120, 160).

### F3 — Help popover not vertically centered (severity: low)

**Symptom.** With viewer closed and `?` pressed at 120×40, the
popover renders horizontally centered but vertically pinned ~1 row
from the top instead of ~10 rows from the top.

**Root cause hypothesis.** Per ADR-0072, centering is `lipgloss.Place`.
Either (a) `App.View` returns `m.help.View(width, height)` with
height too small, (b) `HelpPopover.View` hardcodes a top-anchor, or
(c) `lipgloss.Place` is called with `lipgloss.Top` instead of
`lipgloss.Center` for the vertical axis.

**Fix approach.** Read `internal/ui/help.go` and `app.go`. Single
spot fix likely. Add a render test that exercises help at 120×40
and asserts top margin ≈ bottom margin (±1).

**Acceptance.** Tmux capture at 120×40 with `?` shows popover
roughly centered (top blank rows within ±1 of bottom blank rows).

### F3b — Help popover overlay + dimmed background (severity: med, scope expansion)

**Request.** The popover should appear *over* the context it was
called from — the underlying account view (or viewer) stays visible
but dimmed, so the popover reads as a modal overlay rather than a
replacement screen. Closes BACKLOG #14; supersedes ADR-0071's
"no dim in v1" decision.

**Research finding (2026-04-27).** Neither bubbletea, lipgloss v1,
lipgloss v2, nor `charmbracelet/x` ships a modal-overlay or
background-dim helper. bubbletea issue #642 (Jan 2023) requesting
this is still unanswered. No reference Charm app dims a modal
background. The community converged on two patterns:
  1. **Overlay (compositing only):** vendor superfile's
     `PlaceOverlay` algorithm — line-level positional merge that
     preserves background ANSI codes outside the overlay rect and
     substitutes foreground content inside. Used by superfile,
     `rmhubbert/bubbletea-overlay`, `quickphosphat/bubbletea-overlay`.
     ~80 LOC, MIT.
  2. **No dim** — every reference app (glow, soft-serve, gum) shows
     the popover atop the live background at full brightness.

Since we want a real dim and the ecosystem has no helper for it, we
own a small SGR-faint injector. The transform is much simpler than
the "rewrite every foreground color" approach my first draft scoped:
inject the faint parameter (SGR `2`) into every existing
`\x1b[…m` sequence and patch resets `\x1b[0m` → `\x1b[0;2m` so faint
survives. ~20 LOC, regex-driven; no SGR parser, no color awareness.

**Current state.** Per ADR-0071 and the invariant `App.View`
short-circuit, when `helpOpen` is true `App.View` returns
`m.help.View(width, height)` *directly* — the underlying layout is
skipped entirely. We replace that with: render the underlying
frame, dim it, composite the popover atop.

**Fix approach (split into two sub-fixes).**

**F3b.1 — Vendor `PlaceOverlay`.** New file
`internal/ui/overlay.go`, vendored from
`github.com/yorukot/superfile/src/pkg/string_function/overplace.go`
with a top-of-file provenance comment naming source repo, commit,
and license (MIT). Width math uses `displayCells` for icon-bearing
strings (per ADR-0079); the upstream uses `charmansi.StringWidth`,
which we can keep for non-icon ANSI text. Add a unit test covering
plain text overlay, ANSI-styled overlay, and an off-edge case.

**F3b.2 — SGR faint injector.** New file `internal/ui/dim.go` with
`DimANSI(s string) string`. Implementation:
- Compile a regex matching `\x1b\[([0-9;]*)m`.
- For each match: if params are empty or `0`, output
  `\x1b[0;2m`. Otherwise output `\x1b[2;<params>m`. (The order
  doesn't matter to terminals — both forms set faint plus the
  other attributes.)
- Prepend `\x1b[2m` at string start so unstyled leading content
  is also dim.
Unit test against representative lipgloss-rendered fixtures
(256-color, truecolor, nested styles, reset sequences).

**F3b.3 — Wire into App.View.** When `helpOpen`:
1. Render the underlying frame (the same path used when help is
   closed).
2. Pass through `DimANSI`.
3. Render the popover with `m.help.View(...)`.
4. Composite via `PlaceOverlay`. Use the popover's natural width
   and `lipgloss.Place`-style centering offsets.

**ADR + invariants.** Write a new ADR (likely 0082 or 0084
depending on F2 outcome) titled "Help popover overlay with dimmed
background." Mark ADR-0071 `status: superseded by NNNN`. Update the
invariant currently reading "No background dim in v1" to describe
the new behavior (vendored `PlaceOverlay` + SGR faint injector).
The system-map and bubbletea-conventions docs gain a brief mention
of the `internal/ui/overlay.go` and `internal/ui/dim.go` files.

**Acceptance.** Live tmux capture at 120×40: pressing `?` shows
the popover atop a clearly dimmed account view (folder names,
message rows, chrome all visibly dimmer than the popover). Closing
help restores full brightness with no flicker. Same behavior when
`?` is pressed with the viewer open (popover sits over the dimmed
viewer + sidebar layout).

### F6 — Viewer left padding (severity: low, UX tightening)

**Request.** When viewing a message, add one column of padding
between the sidebar/viewer column divider (`│`) and the start of
viewer content. Currently `From: Sam Raife` (and every body line)
butts directly against the divider.

**Fix approach.** The viewer's effective render width drops by 1
to make room for a leading space on every rendered line. Apply at
the `Viewer.View()` boundary, not at `App.View` (children own
their width contract per the conventions doc — App should not
post-pad child output). The viewer's wordwrap+hardwrap budgets
must use the reduced width so wrapping is still correct.

**Acceptance.** Tmux capture at 120×40 with viewer open shows a
single blank column between the `│` divider and viewer content
(headers, body, horizontal rule, footnotes). The horizontal rule
under the headers still ends at the right border. No regression
in F1's overflow fix.

### F4 — Search shelf shows "0 results" with empty query (severity: low)

**Symptom.** Pressing `/` opens the search shelf showing
`[name]    0 results` before any character is typed. Empty query
should match every visible thread (33 results in the test inbox)
or suppress the count entirely until first keystroke.

**Root cause hypothesis.** Filter logic treats empty query as "no
match" rather than "no filter applied." Likely a one-line fix in
the sidebar/search shelf result counter.

**Fix approach.** Suppress the count line until query length > 0.
That's the cleanest UX (the count is only informative once filtering
is active).

**Acceptance.** `/` shows the prompt with no count line. Typing one
character reveals the count.

### F5 — Message-list column padding (not a bug — UX tightening)

**Request.** Two spaces of padding between sender name and subject,
and between subject and date, in the message list. Currently spacing
appears uneven / single-space in places.

**Fix approach.** Update `message_list.go` row construction to use
fixed 2-cell gaps. Re-measure column budgets so the date column
still right-aligns at the row's right edge.

**Acceptance.** Visual inspection at 120×40 — exactly 2 spaces
between name column and subject, and between subject and date.

## Convention deviations

None planned. Every fix moves the code *closer* to the
`bubbletea-conventions.md` contract, not away from it. If F2's
investigation reveals that the `displayCells` SPUA-A correction needs
softening (terminal-dependent), an ADR will document the change.

## Order of operations

1. F1 (viewer overflow) — highest impact, hardest, do first while
   fresh.
2. F2 (border jitter) — investigate width of `󰇮` first; the fix
   path branches on that finding.
3. F5 (column padding) — small, paired naturally with F2 since
   both are message_list.go.
4. F3 (popover centering) — small isolated fix.
5. F3b (popover overlay + dim) — pairs with F3 (same file), but
   larger. Vendored overlay + new SGR-faint module; supersedes an ADR.
6. F6 (viewer left padding) — small, isolated to viewer width math.
7. F4 (search "0 results") — small isolated fix.

## Pass-end ritual

ADRs only for findings that change a binding fact:
- F2 will produce ADR-0082 if the displayCells policy changes (or
  if a terminal-config prerequisite is added to invariants).
- F1 may produce ADR-0083 if the wordwrap+hardwrap call-site
  contract is sharpened (e.g. "hardwrap is the *outermost* step
  before lines hit the parent").

F3b will produce an ADR superseding ADR-0071 (background dim
decision reversed; ANSI-rewrite approach codified) and an
invariants edit replacing the "No background dim in v1" line.

F3, F4, F5 are pure bug fixes — no ADRs needed unless investigation
surfaces something binding.

Standard pass-end checklist: /simplify, idiomatic-bubbletea review
(§10 of conventions doc), invariants update if any ADR landed,
STATUS.md, archive plan, make check, commit + push + install.

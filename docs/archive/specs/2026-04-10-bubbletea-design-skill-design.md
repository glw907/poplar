# bubbletea-design Skill Design Spec

A Claude Code skill for designing and verifying bubbletea terminal UIs
with headless analysis. General-purpose — works on any bubbletea
project, not poplar-specific.

## Skill Identity

- **Name:** `bubbletea-design`
- **Skill type:** Pattern (way of thinking about TUI design and
  verification, not a mechanical procedure)
- **Delivery:** Single SKILL.md file, installed as a personal Claude
  Code skill at `~/.claude/skills/bubbletea-design/SKILL.md`

### Frontmatter

```yaml
name: bubbletea-design
description: >-
  Use when building or modifying bubbletea terminal UI components,
  diagnosing TUI rendering or alignment issues, or reviewing TUI
  layout code for visual correctness. Also use when choosing icons,
  spacing, box-drawing characters, or theme-driven color decisions
  for terminal interfaces.
```

The description contains **triggers only** — no workflow summary.
This prevents Claude from shortcutting the skill body by following
a description-level summary instead of loading the full SKILL.md.
(See writing-skills CSO guidance for rationale.)

### Token Budget

Target: **under 400 lines** for the SKILL.md body. The skill
teaches design thinking and a verification mindset — Claude already
knows bubbletea, lipgloss, and Go. Don't explain what these are;
explain how to think about design and verification with them.

## Section 1: TUI Design Thinking

This section is the terminal equivalent of frontend-design's
aesthetic guidelines. Voice should be direct and opinionated, like
frontend-design's "commit to a BOLD aesthetic direction."

### Theme-Driven Design

All color decisions reference semantic theme slots (`accent_primary`,
`fg_dim`, `color_error`), never hex values. Components must survive
a theme swap — if two elements differ only by hue, they collapse
when the palette changes. Differentiate by semantic role, not color.

The skill should prompt: "does this distinction survive switching
from Nord to Solarized?" If not, the design relies on a specific
palette, not the semantic system.

### Nerd Font Iconography

Nerd Font families (`nf-md-*` Material Design, `nf-dev-*` devicons,
`nf-cod-*` codicons, `nf-pl-*` Powerline) have different visual
weight and — critically — different cell widths. Some icons render
as double-width despite being a single codepoint.

Key rules:
- Always verify icon width with `lipgloss.Width()`, not `len()` or
  rune count
- Match icon visual weight to element importance — a heavy icon on
  a de-emphasized element creates noise
- Material Design (`nf-md-*`) has the broadest coverage for
  application UIs

### Character Cell Realities

Every glyph is 1 or 2 cells. Width assumptions are the #1 source
of alignment bugs. Box-drawing characters are always single-width.
CJK, some emoji, and certain Nerd Font icons are double-width.

**The rule:** all padding and alignment uses display width
(`lipgloss.Width`), never string length or rune count. A column
that looks aligned in code may be ragged on screen.

### Box-Drawing Vocabulary

| Family | Characters | Use |
|--------|-----------|-----|
| Single-line | `│ ─ ┬ ┴ ├ ┤ ┼` | Structural borders, panel dividers |
| Rounded | `╭ ╮ ╰ ╯` | Floating elements (tab bubbles, cards) |
| Double-line | `║ ═ ╔ ╗ ╚ ╝` | Heavy emphasis, dialog borders |

Mix families intentionally. Junction connections are the
highest-priority verification target — a `─` meeting a `│` must
use the correct T-junction or corner piece. When mixing families
(rounded corners meeting single-line frame), verify that Unicode
has the transition glyph.

### Information Density and Hierarchy

The grid is finite — every cell earns its place. Create visual
layers through theme contrast: `fg_bright` for primary,
`fg_base` for secondary, `fg_dim` for tertiary. Bold and underline
as texture, not emphasis (emphasis comes from color role).

Blank lines cost a full row — use them deliberately as group
separators, not as filler.

### Spatial Composition

- `lipgloss.JoinHorizontal`/`JoinVertical` for panel layout
- Dividers connect to borders (verified by junction checks)
- Floating elements break the grid intentionally
- Alignment columns create vertical rhythm — a divider position
  should be consistent from header through content to footer

### Aesthetic Intentionality

Same principle as frontend-design: commit to a direction and execute
precisely. The skill doesn't prescribe an aesthetic — it demands
that one is chosen and followed through. "Better Pine", brutalist
terminal, maximalist dashboard, minimal zen — the choice shapes
every spacing, icon, and density decision.

## Section 2: Headless Analysis Protocol

Three modes, used in combination, starting from the fastest.
**This is the core differentiator** — the defined method for
inspecting TUI state that prevents "guess and ship."

### Mode 1: Render-and-Inspect (automated)

Structural correctness — connections, alignment, borders.

```
1. Instantiate model with known dimensions (e.g., 80x24)
2. model.View() → strip ANSI (regexp \x1b\[[0-9;]*[a-zA-Z])
3. Split into lines, iterate as rune slices
4. Positional checks:
   - Junction: ┬ on row N aligns with │ on rows N+1..M
   - Border continuity: every content row ends with same char
   - Column alignment: elements line up across rows
   - Width: lipgloss.Width(line) == expected terminal width
```

Write as Go test assertions. This is the regression safety net.

**When:** After any layout code change.

### Mode 2: Render-and-Snapshot (visual review)

Aesthetic judgment — spacing, density, feel.

Same render + strip pipeline. Paste the stripped plain-text grid
(or relevant rows) into the conversation. Scan for ragged alignment,
inconsistent padding, visual weight imbalance.

**When:** Building a new component, user reports a visual issue,
or positional checks pass but something looks off.

### Mode 3: Live Snapshot (ground truth)

Run binary in a terminal, capture screenshot, read the image.
Compare against headless render — discrepancy reveals a wrong
rendering assumption.

**When:** Sparingly. Only when Modes 1-2 aren't sufficient, or for
final sign-off on a major visual change.

### Priority

Always start Mode 1. Escalate to Mode 2 to see the grid. Mode 3
only when Modes 1-2 diverge from reality.

## Section 3: Iteration Loop

```
Identify → Analyze → Fix → Verify → Report
```

**Analyze before touching code.** Know exactly what character is at
what position and why. "Row 3 has `─` at column 30 where `┬` is
expected because `dividerCol` is off by one" — not "the lines look
wrong."

**Verify with the same mode that found the issue.** Mode 1 found
it → run the test again. Mode 2 found it → paste the new grid.
Never skip verification.

**Report matched to context:**
- Verification summary (terse) when confirming a fix
- Visual snippet when the user needs to see the layout
- Both when closing out a visual issue

### Red Flags — STOP

| Thought | Reality |
|---------|---------|
| "I'll just change this character" | Render first. Know what's there now. |
| "It compiles, layout must be right" | Layout correctness requires visual verification. |
| "Looks fine at 80 columns" | Also check 120 and minimum viable (40). |
| "The test passes so we're done" | Positional checks catch structure, not aesthetics. Use Mode 2. |
| "I can see from the code it'll work" | The chrome shell took 3 attempts. Render it. |

## Section 4: TUI Quality Checklist

Quick reference — verify on any component:

- **Junction connections** — every `─` meeting `│` has correct piece
- **Border continuity** — consistent edge chars across all rows
- **Column alignment** — divider consistent header through footer
- **Width budget** — `lipgloss.Width` matches terminal width
- **Theme survival** — distinctions hold across palette swaps
- **Icon width** — `lipgloss.Width`, not rune count
- **Minimum viable width** — no crash or garble at 40 columns
- **Nerd Font weight** — icons match element importance
- **Whitespace budget** — padding and blank rows are deliberate

## Implementation Notes

### File Location

Personal skill: `~/.claude/skills/bubbletea-design/SKILL.md`

Not project-specific — this skill applies to any bubbletea project.

### Relationship to Other Skills

- **frontend-design** — web equivalent. bubbletea-design fills the
  same role for terminal UIs. They don't overlap — different medium,
  different constraints, different verification methods.
- **superpowers:writing-skills** — use this to create the actual
  SKILL.md, following the TDD-for-docs methodology (RED: baseline
  without skill, GREEN: write skill, REFACTOR: close loopholes).
- **elm-conventions** — project-specific Elm architecture rules.
  bubbletea-design assumes Elm architecture but doesn't enforce it
  (that's elm-conventions' job).

### Testing the Skill

Per writing-skills methodology, test as a Pattern skill:

- **Recognition:** Does Claude invoke the skill when building TUI
  components? When diagnosing alignment issues?
- **Application:** Does Claude use Mode 1 before Mode 3? Does it
  render before fixing? Does it verify after fixing?
- **Counter-examples:** Does Claude skip the skill for non-UI Go
  code? For backend/mail adapter work?

Pressure scenario: "Just fix the border character, it's obviously
wrong." The skill should prevent blind fixes — Claude should render
first, even when the fix seems obvious.

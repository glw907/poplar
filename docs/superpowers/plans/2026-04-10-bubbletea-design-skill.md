# bubbletea-design Skill Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a personal Claude Code skill that teaches TUI design thinking and headless verification for bubbletea projects.

**Architecture:** Single SKILL.md file at `~/.claude/skills/bubbletea-design/SKILL.md`. Pattern skill — teaches a way of thinking, not a procedure. Modeled on the frontend-design skill's voice and structure, adapted for terminal constraints.

**Tech Stack:** Markdown (YAML frontmatter), Claude Code skill format

**Spec:** `docs/superpowers/specs/2026-04-10-bubbletea-design-skill-design.md`

---

### File Structure

```
~/.claude/skills/bubbletea-design/
  SKILL.md    # Everything inline, no supporting files
```

Single file, under 400 lines. No references/ or scripts/ — the skill is pure guidance.

---

### Task 1: Write SKILL.md

**Files:**
- Create: `~/.claude/skills/bubbletea-design/SKILL.md`

- [ ] **Step 1: Create the skill directory**

```bash
mkdir -p ~/.claude/skills/bubbletea-design
```

- [ ] **Step 2: Write the SKILL.md**

Write the complete file with this structure. The spec at `docs/superpowers/specs/2026-04-10-bubbletea-design-skill-design.md` has the full content for each section — transform it into the skill's voice (direct, opinionated, concise).

**Frontmatter:**

```yaml
---
name: bubbletea-design
description: >-
  Use when building or modifying bubbletea terminal UI components,
  diagnosing TUI rendering or alignment issues, or reviewing TUI
  layout code for visual correctness. Also use when choosing icons,
  spacing, box-drawing characters, or theme-driven color decisions
  for terminal interfaces.
---
```

**Body structure** (each section maps to a spec section):

```markdown
# Bubbletea TUI Design

[1-2 sentence core principle: design thinking + headless
verification for terminal UIs]

## TUI Design Language

### Theme-Driven Design
[From spec Section 1: semantic slots, theme survival test]

### Nerd Font Iconography
[From spec: families, width verification, visual weight]

### Character Cell Realities
[From spec: display width vs string length, lipgloss.Width]

### Box-Drawing Vocabulary
[From spec: three families table, junction priority, mixing rules]

### Information Density
[From spec: fg_bright/fg_base/fg_dim hierarchy, whitespace cost]

### Spatial Composition
[From spec: JoinHorizontal/Vertical, divider connections,
alignment columns]

### Aesthetic Intentionality
[From spec: commit to a direction, don't prescribe one]

## Headless Analysis Protocol

### Mode 1: Render-and-Inspect
[From spec: View() → strip ANSI → positional checks. Show the
strip regex. List the four check types. When to use.]

### Mode 2: Render-and-Snapshot
[From spec: paste stripped grid. When to use.]

### Mode 3: Live Snapshot
[From spec: terminal screenshot. When to use.]

### Priority
[From spec: 1 → 2 → 3 escalation]

## Iteration Loop

[From spec: Identify → Analyze → Fix → Verify → Report.
Analyze before touching code. Verify with same mode. Report
format rules.]

### Red Flags — STOP

[From spec: the 5-row rationalization table]

## Quality Checklist

[From spec: the 9 verification items as a scannable list]
```

**Voice guidelines** (follow these while writing):
- Direct and opinionated, like frontend-design
- Assume Claude knows bubbletea, lipgloss, and Go — don't explain what they are
- Explain *how to think*, not *what to do*
- Use "you" sparingly — prefer imperative or declarative
- Concrete examples over abstract principles (the chrome shell `┬` story is a good anchor for why verification matters, but don't tell the story — state the rule)

- [ ] **Step 3: Check line count**

```bash
wc -l ~/.claude/skills/bubbletea-design/SKILL.md
```

Expected: under 400 lines. If over, trim — look for explanations of things Claude already knows (what lipgloss is, what bubbletea does) and cut them.

- [ ] **Step 4: Commit**

```bash
cd ~/Projects/beautiful-aerc
git add -f ~/.claude/skills/bubbletea-design/SKILL.md
git commit -m "Add bubbletea-design skill for TUI design and verification"
```

Note: `~/.claude/skills/` is outside the repo, so this file lives on the local filesystem only, not in the project repo. If git add fails because it's outside the worktree, that's expected — skip the commit. The skill is a personal tool, not a project artifact.

---

### Task 2: Verify Skill Discovery

**Files:**
- Read: `~/.claude/skills/bubbletea-design/SKILL.md`

- [ ] **Step 1: Verify the skill appears in Claude Code**

Start a new Claude Code session (or use `/clear`) and check that the skill appears in the available skills list. The description should trigger on TUI-related prompts.

- [ ] **Step 2: Verify the frontmatter parses correctly**

Check that the YAML frontmatter has no syntax errors by confirming the skill loads without warnings. The `>-` folded scalar in the description must render as a single line with no trailing newline.

- [ ] **Step 3: Test trigger recognition**

In a fresh session, try prompts that should trigger the skill:
- "Build a sidebar component for the bubbletea app"
- "The box-drawing characters aren't connecting properly"
- "Review the tab bar layout code"

And prompts that should NOT trigger:
- "Fix the JMAP authentication error"
- "Add a new cobra subcommand"
- "Write a unit test for the HTML filter"

Document which prompts triggered correctly and which didn't.

---

### Task 3: Smoke Test with Poplar

**Files:**
- Read: `internal/ui/tab_bar.go`
- Read: `internal/ui/app.go`
- Read: `internal/ui/app_test.go`

- [ ] **Step 1: Use the skill on an existing component**

In a fresh Claude Code session, ask: "Review the tab bar rendering in internal/ui/tab_bar.go for visual correctness."

The skill should guide Claude to:
1. Use Mode 1 (render and inspect) — instantiate the model, call View(), strip ANSI, check junctions
2. Report findings with positional detail ("row 3, column 30")
3. Reference the quality checklist items

- [ ] **Step 2: Verify the Red Flags table works**

Ask: "Just change the ╮ to ╯ on row 3, it's obviously wrong."

The skill should prevent a blind fix — Claude should render first, analyze the current state, and explain what's actually at that position before changing anything.

- [ ] **Step 3: Document any skill adjustments needed**

If the smoke test reveals gaps (Claude skipped verification, didn't use the right mode, ignored the checklist), note specific wording changes needed in the SKILL.md. Apply fixes immediately.

- [ ] **Step 4: Commit any fixes**

If SKILL.md was updated:

```bash
git add -f ~/.claude/skills/bubbletea-design/SKILL.md
git commit -m "Refine bubbletea-design skill after smoke testing"
```

(Same note about personal skill location applies — skip commit if outside worktree.)

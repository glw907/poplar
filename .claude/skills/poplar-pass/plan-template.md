# Pass <n> — <topic>

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> `superpowers:subagent-driven-development` (recommended) or
> `superpowers:executing-plans` to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** One paragraph stating what this pass accomplishes and
what the resulting state looks like.

**Why now:** Brief context — what surfaced the pass, what it
unblocks, why it can't be deferred.

**Architecture:** What this pass produces (deliverables, files,
changes). Bullets, not prose.

**Tech Stack:** Go version, libraries, frameworks. Note any new
runtime deps.

**Required reading before starting:**
- Invoke `go-conventions` skill before any Go file change.
- **If the pass touches `internal/ui/`:** invoke `elm-conventions`
  skill, read `docs/poplar/bubbletea-conventions.md`, and confirm
  the plan cites a bubbles analogue (viewport, list, table,
  textinput, textarea, spinner, help) for each new component.
  Deviations from a bubbles analogue must be named in the plan
  with rationale — "we need a custom list because X" is fine;
  "we just wrote a custom thing" is not.
- Read `docs/poplar/invariants.md` once.
- Read any prior research / audit docs the pass depends on.

**Conventions for this plan:** Each task is one coherent unit of
work that ends with `make check` green and a commit. Subagents are
dispatched per task with self-contained prompts (no shared state).
Commits use imperative mood with `Co-Authored-By: Claude` trailer.

---

## Phase 0 — Research (delete if implementation-only)

Research tracks dispatch as separate subagents and produce written
findings docs that subsequent phases consume. Each research task
ends with a citable deliverable in `docs/poplar/research/`.

### Task 1: <research topic>

**Files:**
- Create: `docs/poplar/research/YYYY-MM-DD-<topic>.md`

- [ ] Step 1: Dispatch research subagent (self-contained prompt)
- [ ] Step 2: Spot-check N citations from the deliverable
- [ ] Step 3: Commit

---

## Phase 1 — Implementation

### Task N: <one task per coherent unit of work>

**Files:**
- Modify: `internal/<package>/<file>.go`
- Create: `internal/<package>/<new-file>.go`

- [ ] Step 1: <prerequisite read / context gathering>
- [ ] Step 2: <core implementation>
- [ ] Step 3: `make check`
- [ ] Step 4: Commit

---

## Phase N — Pass-end consolidation

Standard `poplar-pass` ending ritual. The current ritual lives in
`.claude/skills/poplar-pass/SKILL.md`.

- [ ] `/simplify` on the diff
- [ ] If `internal/ui/` changed: run step 1b idiomatic-bubbletea check
- [ ] Write ADRs for each design decision
- [ ] Update `invariants.md` in place
- [ ] Update `STATUS.md` (mark done, write next starter prompt)
- [ ] Archive plan + spec to `docs/superpowers/archive/`
- [ ] `make check` green
- [ ] Commit, push, install

---

## Out of scope

- <bullet list of things adjacent to this pass that are NOT being
  done here, with a forward pointer if relevant>

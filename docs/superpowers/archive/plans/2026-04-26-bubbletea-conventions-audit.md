# Bubbletea Conventions Audit & Infrastructure Pass

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ground poplar's "idiomatic bubbletea" discipline in
researched community norms, audit the existing `internal/ui/` for
divergences, fix the divergences worth fixing now (log the rest),
and lock in Claude-side infrastructure (hooks, skills, docs,
templates) that keeps poplar on the bubbletea showcase path going
forward.

**Why now:** The Pass 3 verification surfaced a body-text bleed
that turned out to be a Nerd Font icon-width miscount (BACKLOG
#16). Investigating it exposed that the just-written
`docs/poplar/bubbletea-conventions.md` was authored from memory
rather than research, so its rules are plausible but unverified.
Poplar's design vision explicitly positions it as a bubbletea
showcase — credibility on that front requires the conventions to
match what the Charm community actually does, and requires
ongoing structural enforcement so future passes don't drift.

**Architecture:** This pass produces three deliverables:

1. **A researched conventions doc** —
   `docs/poplar/bubbletea-conventions.md` rewritten from primary
   sources (bubbles components, glamour, lipgloss API, bubbletea
   itself) and reference apps (charm examples, glow, gum,
   soft-serve), with every claim citable.
2. **A divergence-fix sprint** — `internal/ui/` audited against
   the doc; high/medium-severity divergences fixed in-pass,
   low-severity logged to BACKLOG with severity + suggested fix.
3. **Forward-looking infrastructure** — a new
   `bubbletea-conventions-lint.sh` hook, refreshed
   `elm-conventions` skill, updated `poplar-pass` ritual, and
   plan/spec templates that bake the discipline into every
   future pass.

**Tech Stack:** Go 1.26.1, bubbletea, bubbles, lipgloss,
charmbracelet/x/ansi (already vendored). No new runtime deps.
Hook scripts are bash + jq.

**Required reading before starting:**
- Invoke `go-conventions` skill before any Go file change.
- Invoke `elm-conventions` skill before any `internal/ui/` change.
- Read `docs/poplar/invariants.md` once.
- Read the current `docs/poplar/bubbletea-conventions.md` once
  (this pass rewrites it — read it as the *starting* draft, not
  as authoritative).
- Read `BACKLOG.md` #16 (the icon-width bug that motivates this).
- Skim the `bubbles` source tree at
  `~/go/pkg/mod/github.com/charmbracelet/bubbles@*` —
  particularly `viewport/`, `list/`, `table/`, `textinput/`,
  `textarea/`, `spinner/`, `help/`, `key/`. The audit needs a
  ground-truth view of what the community considers idiomatic.

**Conventions for this plan:** Each task is one coherent unit of
work that ends with `make check` green and a commit. Research
tasks dispatch subagents and produce a written findings doc as
their commit artifact. Audit tasks produce a divergence report
under `docs/poplar/audits/`. Fix tasks reference the audit
finding by ID. Hook scripts are tested with synthetic stdin
before commit. Commits use imperative mood with Co-Authored-By
trailer.

---

## Phase 0 — Research

Two parallel research tracks. Both produce written, citable
findings in `docs/poplar/research/` that subsequent phases
consume. Dispatch as separate subagents — they share no state.

### Task 1: Bubbles + glamour + lipgloss + bubbletea source survey

**Files:**
- Create: `docs/poplar/research/2026-04-26-bubbletea-norms.md`

- [ ] **Step 1: Dispatch research subagent**

Subagent prompt (self-contained — agent has no prior context):

> You are surveying the Charm libraries to extract idiomatic
> bubbletea patterns. The deliverable is a single markdown
> document at
> `docs/poplar/research/2026-04-26-bubbletea-norms.md` that a
> downstream pass will use to rewrite a project's conventions
> document. Every claim you make must cite a specific file and
> line in the surveyed source — no hand-waving from training
> knowledge.
>
> **Sources to read (in this order):**
>
> 1. **bubbles** at
>    `~/go/pkg/mod/github.com/charmbracelet/bubbles@*` — read every
>    component package. For each (viewport, list, table, textinput,
>    textarea, spinner, help, key, paginator, progress, stopwatch,
>    timer, filepicker), document:
>    - Public API surface (constructor, `SetSize`/`SetWidth`,
>      Update return type, View signature, message types emitted)
>    - How `View()` enforces width and height
>    - How key bindings are declared and dispatched
>    - How focus/blur is managed (if applicable)
>    - Any "gotchas" called out in comments or examples
> 2. **glamour** at `~/go/pkg/mod/github.com/charmbracelet/glamour@*`
>    — focus on the renderer entry point and how it honors width.
> 3. **lipgloss** at
>    `~/go/pkg/mod/github.com/charmbracelet/lipgloss@*` — the
>    `Style` API patterns: when to use `Width` vs `MaxWidth`,
>    `Inline`, `Render` semantics, `JoinHorizontal`/`JoinVertical`
>    padding behavior, `Place` semantics. Read the package
>    examples and tests for canonical usage.
> 4. **bubbletea** at
>    `~/go/pkg/mod/github.com/charmbracelet/bubbletea@*` — the
>    `tea.Program` constructor options (alt screen, mouse modes,
>    focus reporting, output, input), the `tea.Model` interface,
>    `tea.Batch`/`tea.Sequence`/`tea.Tick`, how
>    `tea.WindowSizeMsg` is delivered.
> 5. **charmbracelet/x/ansi** — `Wordwrap`, `Hardwrap`,
>    `Truncate`, `Strip` — semantics and when to use each.
>
> **Output structure** (sections; cite file:line for every
> normative claim):
>
> 1. *Component shape* — the canonical component Update return
>    shape (`(M, tea.Cmd)` vs `(tea.Model, tea.Cmd)`), View
>    signature, what state ownership patterns are common.
> 2. *Sizing* — how components honor width/height, `tea.WindowSizeMsg`
>    handling vs `SetSize` methods, padding/clipping idioms.
> 3. *Key bindings* — `key.Binding` and `bubbles/key` usage,
>    help integration via `bubbles/help`, key declaration vs.
>    string-matching in Update.
> 4. *Layout primitives* — `JoinHorizontal`/`JoinVertical`
>    behavior with mismatched-height/width inputs, `Place`,
>    `Style.Width`/`MaxWidth` differences, when each is right.
> 5. *Text rendering* — wordwrap, hardwrap, truncate; how
>    glamour combines them for arbitrary input.
> 6. *Program-level* — `tea.NewProgram` options most apps use,
>    common `tea.Cmd` patterns (Tick, async I/O, Batch).
> 7. *Anti-patterns called out by the libraries themselves* —
>    anything the source comments explicitly warn against.
>
> Keep claims terse. Citations must be path + line range so the
> downstream consumer can verify. Aim for 400–800 lines total.

- [ ] **Step 2: Verify the deliverable**

Spot-check 5 random citations: open the cited file at the cited
line, confirm the claim. If any fail, dispatch a fix subagent
with the specific failures — don't accept the doc until every
spot-check passes.

- [ ] **Step 3: Commit**

`git add docs/poplar/research/2026-04-26-bubbletea-norms.md`,
commit with message
`Pass <n>: bubbles/glamour/lipgloss/bubbletea source survey`.

### Task 2: Reference-app survey

**Files:**
- Create: `docs/poplar/research/2026-04-26-reference-apps.md`

- [ ] **Step 1: Dispatch research subagent**

Subagent prompt (self-contained):

> You are surveying real-world bubbletea applications to extract
> the patterns the community treats as idiomatic in production
> code (as opposed to library examples). The deliverable is a
> single markdown document at
> `docs/poplar/research/2026-04-26-reference-apps.md`. Every
> claim must cite a specific repo + file path + commit/tag (or
> permalink). No hand-waving from training knowledge.
>
> **Apps to survey** (use WebFetch and/or `git clone` into
> `/tmp/refapps/` for inspection):
>
> 1. **charmbracelet/bubbletea/examples/** — the official
>    examples. Read every example that demonstrates layout,
>    multi-pane composition, modal overlays, help integration,
>    focus management, or window-resize handling.
> 2. **charmbracelet/glow** — markdown reader, multi-pane.
> 3. **charmbracelet/gum** — script primitives; useful for how
>    each command's bubbletea program is structured at the
>    boundary.
> 4. **charmbracelet/soft-serve** — server with TUI; multi-tab,
>    multi-view.
> 5. **charmbracelet/wishlist** — SSH session list; lists +
>    forms + viewports.
> 6. **One non-Charm community app of the agent's choice** that
>    has visible attention from the Charm community (recent
>    Charm star, Charm blog post, Awesome Bubbletea entry).
>    Document why this app was selected.
>
> **Output structure** (sections; cite repo + path + line for
> every normative claim):
>
> 1. *App shape* — single-model vs nested-model trees, where
>    state lives, how programs handle quit/sigint.
> 2. *Layout composition* — how multi-pane apps compose
>    sidebars, panels, modals; what they delegate to lipgloss
>    vs hand-roll.
> 3. *Help and key bindings* — `bubbles/help` adoption,
>    `key.Binding` usage in real apps, single-key vs chord
>    conventions.
> 4. *Window resize* — how each app threads `WindowSizeMsg`
>    through nested components.
> 5. *Async I/O* — patterns for long-running commands, polling
>    loops, server pushes.
> 6. *Theming and styling* — how production apps organize their
>    lipgloss styles (struct vs constants vs theme types).
> 7. *Patterns to emulate* — concrete recommendations
>    extracted from cross-app overlap.
> 8. *Patterns to avoid* — divergences between apps and the
>    library docs, or commonly-cargo-culted bad shapes.
>
> Keep it terse. Aim for 300–600 lines total. Permalinks (e.g.
> `github.com/charmbracelet/glow/blob/v1.5.1/ui/pager.go#L210`)
> are mandatory for every citation.

- [ ] **Step 2: Verify the deliverable**

Spot-check 5 citations via WebFetch on the permalinks. If any
fail, dispatch a fix subagent.

- [ ] **Step 3: Commit**

`git add docs/poplar/research/2026-04-26-reference-apps.md`,
commit `Pass <n>: bubbletea reference-app survey`.

---

## Phase 1 — Rewrite the conventions doc

### Task 3: Rewrite `docs/poplar/bubbletea-conventions.md`

**Files:**
- Modify: `docs/poplar/bubbletea-conventions.md`

This is a single-controller task — the controller (you, with
context fresh from the two research docs) writes the rewrite
inline; no subagent needed. Subagent dispatch is wasteful when
the work is "synthesize two documents into a third" because the
synthesis is exactly the kind of judgment work the controller
has the context for.

- [ ] **Step 1: Read both research docs end-to-end**

Load `docs/poplar/research/2026-04-26-bubbletea-norms.md` and
`docs/poplar/research/2026-04-26-reference-apps.md` fully. Do
not start writing until both are in working memory.

- [ ] **Step 2: Rewrite the conventions doc**

The new doc has the following sections (drop any whose research
returned nothing actionable; add any the research surfaced that
this skeleton missed):

1. *Purpose and scope* — what this doc governs (the structural
   contract between bubbletea components in poplar) and what it
   does not (visual design — that's `bubbletea-design`).
2. *Component shape* — Update return type, View signature,
   constructor pattern, state ownership. **Cite primary
   sources for each rule.**
3. *Sizing contract* — width/height honoring, `WindowSizeMsg`
   vs `SetSize`, the JoinHorizontal trust contract, when to
   self-guard with a clip-pane helper.
4. *Text rendering* — wordwrap + hardwrap + truncate semantics
   from `charmbracelet/x/ansi`; the renderer's contract to
   honor its width arg.
5. *Key bindings and help* — `bubbles/key.Binding`,
   `bubbles/help` integration, single-key vs chord conventions
   (poplar overrides community defaults here per ADR — call out
   the deviation).
6. *Async I/O and update flow* — `tea.Cmd` capture-by-value,
   batching, polling, error message conventions.
7. *Program setup* — `tea.NewProgram` options poplar uses and
   why, what each option costs.
8. *Anti-patterns* — every anti-pattern from research, each
   tied to a specific failure mode and a concrete example.
9. *Planning checklist* — the questions a plan doc must answer
   before any UI code lands. Update from the current draft
   based on what research surfaced.
10. *Review checklist* — the verifications that gate a UI
    diff. Update from current draft.
11. *See also* — pointers to research docs (canonical sources),
    `elm-conventions` skill, `bubbletea-design` skill, tmux
    testing doc, the lint hook (Phase 3).

Every normative claim in the rewrite cites either a research
doc section or a primary source. No memory-based assertions.

- [ ] **Step 3: `make check`**

Doc-only change so this is a smoke test, but run it anyway —
the convention is "every commit ends green."

- [ ] **Step 4: Commit**

`git add docs/poplar/bubbletea-conventions.md`, commit with
message `Pass <n>: rewrite bubbletea conventions from research`.

---

## Phase 2 — Audit poplar against the conventions

### Task 4: Produce divergence report

**Files:**
- Create: `docs/poplar/audits/2026-04-26-bubbletea-conventions.md`

- [ ] **Step 1: Dispatch audit subagent**

Subagent prompt (self-contained):

> You are auditing poplar's `internal/ui/` against the
> conventions documented in
> `docs/poplar/bubbletea-conventions.md`. The deliverable is a
> divergence report at
> `docs/poplar/audits/2026-04-26-bubbletea-conventions.md`.
>
> **Inputs you must read fully before auditing:**
>
> - `docs/poplar/bubbletea-conventions.md` (the just-rewritten
>   doc — this is the ruler)
> - `docs/poplar/research/2026-04-26-bubbletea-norms.md`
> - `docs/poplar/research/2026-04-26-reference-apps.md`
> - Every `.go` file under `internal/ui/`
>
> **Audit dimensions** (per file under `internal/ui/`):
>
> 1. **Component shape** — Update return type, View signature,
>    constructor.
> 2. **State ownership** — package-level mutables, mutations
>    outside Update, blocking I/O outside tea.Cmd.
> 3. **Sizing contract** — does View honor assigned width/height
>    in all branches? Self-guard present? Renderer width arg
>    honored?
> 4. **Key bindings** — `key.Binding`/`bubbles/key` usage,
>    `bubbles/help` integration, key string-matching in Update.
> 5. **Layout primitives** — `JoinHorizontal`/`JoinVertical`
>    usage, `Place`/`Width`/`MaxWidth` correctness, defensive
>    parent-side clipping (anti-pattern).
> 6. **Text rendering** — wordwrap-only without hardwrap,
>    `len()` for layout math, ANSI-unaware width measurement,
>    SPUA-A icon width assumptions (BACKLOG #16 is one such).
> 7. **Async I/O patterns** — Cmd closures capturing pointers,
>    blocking calls in Update, missed `tea.Batch`.
> 8. **Program setup** — `cmd/poplar/`'s `tea.NewProgram`
>    options vs the doc's recommended set.
>
> **Output structure:**
>
> ```markdown
> # Bubbletea Conventions Audit — 2026-04-26
>
> ## Summary
>
> N findings: X high, Y medium, Z low.
>
> ## Findings
>
> ### A1 — <one-line title>
>
> **Severity:** high | medium | low
> **File:** path/to/file.go:LINE
> **Rule:** <doc section that's violated>
> **Evidence:** <actual code snippet, 5–15 lines>
> **Why it matters:** <concrete failure mode this enables>
> **Suggested fix:** <one paragraph; don't write the fix, just
> describe it>
>
> ...
> ```
>
> Severity rubric:
> - **high** — produces visible user-facing bugs (rendering
>   garbage, crashes at narrow widths, dropped events) OR
>   silently corrupts state (Cmd closures over pointers, mutex
>   in UI code).
> - **medium** — works today but is fragile or off-norm in a
>   way that costs readability or makes future changes harder.
> - **low** — stylistic divergence with no functional cost;
>   noted for completeness.
>
> Be exhaustive. The downstream pass will triage what to fix
> now vs later. Better to over-report than miss something.

- [ ] **Step 2: Triage with the user**

Read the audit. Group findings into:
- **Fix this pass** (high; medium that's a quick fix; medium
  that touches an area we're actively working in).
- **Log to BACKLOG** (medium that's larger; low that's worth
  remembering).
- **Accept as deviation** (rare; explicit override of
  community norm — must be ADR'd in the pass-end ritual).

Write the triage decision into the audit doc as a new
"Triage" section — one line per finding with `[fix-now]`,
`[backlog]`, or `[accept-as-deviation]` and a one-sentence
rationale for `[accept-as-deviation]` choices.

- [ ] **Step 3: Commit**

`git add docs/poplar/audits/2026-04-26-bubbletea-conventions.md`,
commit `Pass <n>: bubbletea conventions audit + triage`.

---

## Phase 3 — Fix divergences

### Task 5: Fix high/medium [fix-now] findings

**Files:** TBD by audit.

This is a sub-loop driven by the audit. For each [fix-now]
finding, in audit-order:

- [ ] **Step 1: Dispatch implementer subagent**

Per finding, dispatch an implementer subagent with: the audit
finding text, the exact files to modify, and the conventions
doc section number it's restoring conformance to. The implementer
follows TDD where applicable. End-state: tests added/updated,
`make check` green, finding's "Suggested fix" applied.

- [ ] **Step 2: Verify**

Live tmux capture at 120×40 if the finding touched layout. Diff
review. Confirm BACKLOG #16 is fixed if the audit's icon-width
finding is in [fix-now] (it should be — it's the original
motivator).

- [ ] **Step 3: Commit per finding**

One commit per finding: `Pass <n>: fix audit-A<id> — <title>`.
Multiple commits in this task is correct — they're independent
changes and reviewable separately.

### Task 6: Log [backlog] findings to BACKLOG.md

**Files:**
- Modify: `BACKLOG.md`

- [ ] **Step 1: Append a BACKLOG entry per finding**

Use the existing BACKLOG format. Each entry references the
audit doc by path + finding ID, includes severity + suggested
fix verbatim, and is tagged `#bubbletea-norms`.

- [ ] **Step 2: Close BACKLOG #16**

If the icon-width fix landed in Task 5, mark #16 done in
BACKLOG.md per project convention.

- [ ] **Step 3: Commit**

`git add BACKLOG.md`, commit
`Pass <n>: log non-blocking conventions findings to backlog`.

---

## Phase 4 — Forward-looking infrastructure

The point of this phase: make it structurally hard for future
passes to silently re-introduce the divergences we just fixed.

### Task 7: Add a `bubbletea-conventions-lint.sh` hook

**Files:**
- Create: `.claude/hooks/bubbletea-conventions-lint.sh`
- Modify: `.claude/settings.json` (register the hook)

Pattern follows `.claude/hooks/elm-architecture-lint.sh`:
PostToolUse on Edit/Write, scoped to `internal/ui/**/*.go`,
emits non-blocking warnings to stderr.

Checks the hook performs (start with the cheap, mechanical
ones — leave semantic checks to human review):

- [ ] **Step 1: Width-math via `len()`**

Flag any `len(` appearance inside an `internal/ui/` Go file
where the surrounding context is clearly width math (heuristic:
within 3 lines of the literal `width` or `Width`). Prefer false
positives over false negatives — the hook is a prompt to
review, not a gate.

- [ ] **Step 2: Wordwrap without Hardwrap**

Flag any `ansi.Wordwrap(` call in a renderer file (heuristic:
file is in `internal/content/` or function takes a `width int`
param) without an `ansi.Hardwrap(` call within 5 lines.

- [ ] **Step 3: Defensive parent-side clipping**

Flag any `lipgloss.NewStyle().MaxWidth(` call applied to a
sub-component's `View()` output (heuristic: the next 2 lines
contain `.View()`).

- [ ] **Step 4: Naked `JoinHorizontal`/`JoinVertical`**

Flag any `lipgloss.JoinHorizontal(` whose immediate caller
function does not visibly enforce width on its inputs in the
preceding 30 lines. This one is a stretch heuristic — leave it
out if it's too noisy after a quick test pass.

- [ ] **Step 5: Test the hook**

Pipe synthetic stdin matching the hook input format
(`{"tool_input":{"file_path":"..."}}`) at known-bad and
known-good fixtures and confirm the warnings fire/don't fire.

- [ ] **Step 6: Register in settings.json**

Add to `PostToolUse` alongside `elm-architecture-lint.sh`.

- [ ] **Step 7: Commit**

`git add .claude/hooks/bubbletea-conventions-lint.sh
.claude/settings.json`, commit
`Pass <n>: add bubbletea conventions lint hook`.

### Task 8: Refresh `elm-conventions` skill

**Files:**
- Modify: `~/.claude/skills/elm-conventions/SKILL.md`

The skill currently has Rule 6 (Components Own Their Size
Contract) added in the same session as the original draft of
the conventions doc. Replace with rules sourced from the
researched doc. If research surfaced additional Elm-architecture-
adjacent norms (e.g. specific message-flow idioms), add them
as Rules 7+.

- [ ] **Step 1: Read the new conventions doc**
- [ ] **Step 2: Rewrite Rule 6 (and add 7+ if warranted)**

Each rule cites the research doc for its authority.

- [ ] **Step 3: Commit (note: this is in `~/.claude/`, outside the project repo)**

The user's `~/.claude/` is a stowed dotfile per CLAUDE.md
("Dotfiles Management"). Edit in place; the dotfile sync flow
is the user's responsibility, not this pass's.

### Task 9: Update `poplar-pass` skill ritual

**Files:**
- Modify: `.claude/skills/poplar-pass/SKILL.md`

The current ritual already has a step 1b "Idiomatic-bubbletea
check." Strengthen it now that the conventions are
research-backed:

- [ ] **Step 1: Replace step 1b with a researched checklist**

Pull the review checklist verbatim from the rewritten
conventions doc. The skill says "run this checklist and
document deviations in an ADR" — same shape, sharper checklist.

- [ ] **Step 2: Add a "linked sources" line in step 4 (Update STATUS)**

When a pass touches `internal/ui/`, STATUS's pass entry must
link to the conventions audit (if any was produced) so future
passes can find it. Trivial format addition.

- [ ] **Step 3: Commit**

`git add .claude/skills/poplar-pass/SKILL.md`, commit
`Pass <n>: tighten poplar-pass UI review checklist`.

### Task 10: Plan & spec template additions

**Files:**
- Create or modify: `.claude/skills/poplar-pass/plan-template.md`
  (if absent — see Step 1)
- Create: `.claude/skills/poplar-pass/spec-template.md` (if absent)

If `poplar-pass` doesn't already ship a plan/spec template
(check first), the existing convention is "look at recent
plan/spec docs and copy the shape." That's fragile. This pass
adds explicit templates so future passes can't accidentally
omit the bubbletea-conventions citation.

- [ ] **Step 1: Determine whether templates already exist**

Check `.claude/skills/poplar-pass/`. If yes, modify; if no,
create.

- [ ] **Step 2: Plan template additions**

The plan template's "Required reading before starting"
section gains a conditional line: "If the pass touches
`internal/ui/`, read `docs/poplar/bubbletea-conventions.md`
and confirm the plan cites a bubbles analogue for each new
component."

- [ ] **Step 3: Spec template additions**

When a UI pass exists, the spec template's "Open questions"
section lists a fixed sub-section: "Bubbletea conventions —
what bubbles components are we using as analogues? what (if
any) are we deviating from and why?"

- [ ] **Step 4: Commit**

`git add .claude/skills/poplar-pass/`, commit
`Pass <n>: add plan + spec templates with conventions hooks`.

### Task 11: Update CLAUDE.md and invariants

**Files:**
- Modify: `CLAUDE.md`
- Modify: `docs/poplar/invariants.md`

- [ ] **Step 1: CLAUDE.md**

The current entry under "On-demand reading" already points to
the conventions doc with a "Load before any UI planning or
review" flag. After this pass, the entry should also point
to the two research docs as the authority of last resort —
"if the conventions doc and the source code disagree, the
research doc cites the primary source."

- [ ] **Step 2: invariants.md**

Add (or update, if a related fact already exists) a binding
fact under the Architecture section:

> Idiomatic bubbletea is the default. UI code uses bubbles
> components as primary analogues; deviations are documented
> in ADR per pass-end ritual. Renderers honor their width arg
> via wordwrap + hardwrap. Components self-enforce size
> contracts in `View()`. The full contract is in
> `docs/poplar/bubbletea-conventions.md`, grounded in
> `docs/poplar/research/`.

Update the decision index table — this pass writes one or
more ADRs (Phase 5) that this fact corresponds to.

- [ ] **Step 3: Commit**

`git add CLAUDE.md docs/poplar/invariants.md`, commit
`Pass <n>: anchor bubbletea conventions in CLAUDE.md and invariants`.

---

## Phase 5 — Pass-end consolidation

### Task 12: Pass-end ritual

The standard `poplar-pass` ending ritual. Note: the ritual itself
was tightened in Task 9 — execute the **new** version, not the
version that was current when this plan was written.

- [ ] **Step 1: `/simplify`** on the diff
- [ ] **Step 2: Run the new step 1b** (idiomatic-bubbletea check
  on this pass's own UI changes — there shouldn't be many, but
  the review must run)
- [ ] **Step 3: Write ADRs** for each design decision made this
  pass. At minimum:
  - ADR: "Bubbletea conventions are research-grounded" — the
    decision that primary sources govern.
  - ADR: "bubbletea-conventions-lint hook" — the decision to
    lint structurally rather than rely on review alone.
  - ADR per [accept-as-deviation] finding from the audit.
  - ADRs for any architecturally-meaningful fix from Task 5.
- [ ] **Step 4: Update invariants.md** in place per the new ADRs
- [ ] **Step 5: Update STATUS.md** — mark pass done, write next
  starter prompt
- [ ] **Step 6: Archive plan + spec** to
  `docs/superpowers/archive/`
- [ ] **Step 7: `make check`** green
- [ ] **Step 8: Commit, push, install**

---

## Out of scope

- Visual design changes (color, icons beyond the specific
  width-fix needed for BACKLOG #16, density). Those are
  `bubbletea-design`'s territory; this pass is about
  structural conventions.
- Compose / catkin design (Pass 9 — separate plan).
- Keybinding redesign. Poplar's modifier-free, single-key,
  no-command-mode philosophy is settled (ADR-0015, 0024,
  0051, 0068, 0076). The audit may surface that we should be
  using `bubbles/key.Binding` for declaration; that's a
  refactor of *how* bindings are registered, not *what* keys
  do what.
- Rewriting `bubbletea-design` skill. It governs the visual
  language; this pass governs the structural language. They
  cohabit; they don't merge.

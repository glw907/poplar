# Mailrender Training System — Design

**Date:** 2026-04-12
**Status:** Approved — ready for implementation planning
**Supersedes:** The ad-hoc pre-pivot `corpus/` + `mailrender save`
workflow and the original `.claude/skills/fix-corpus` skill that
drove it.

## Problem

Post-pivot (ADR 0058, 2026-04-12), poplar's rendering pipeline lives
in `internal/filter/` (`CleanHTML`, `CleanPlain`) and
`internal/content/` (`ParseBlocks`, `RenderBody`). The pre-pivot
feedback loop for rendering bugs was:

1. Spot a problem in aerc.
2. Press `b` in the aerc viewer to run `mailrender save` and drop
   the raw email into `corpus/`.
3. Open the `fix-corpus` Claude skill, which batch-triaged files in
   `corpus/` and drove fixes into the filter pipeline.

That loop is now broken in three ways:

- **Aerc is gone.** The `b` keybinding that saved captures lived in
  aerc, not in poplar. There is no in-app capture path.
- **`mailrender` as a binary is dead.** The stale `mailrender` file
  at the repo root is a pre-pivot artifact; the real rendering code
  is now internal library packages with no standalone CLI wrapper.
- **Captures carry no context.** `corpus/salmon-selvedge.html` and
  the 45 files in `audit-output/G*.txt` are just file dumps. By the
  time a fix is attempted, the original developer observation —
  "the list wrapping is broken here" or "this blockquote loses its
  attribution" — is lost. The skill has to rediscover each issue
  from scratch.

Meanwhile the user's long-term goal is clear: the renderer should
produce **well-crafted markdown**, where "well-crafted" is largely
objective. Today that definition lives in the developer's head and
gets re-explained on every renderer fix. That's wasted effort, and
worse, it means the skill rediscovers the standard on each pass
and judges captures inconsistently.

"Well-crafted" is multi-dimensional — not just a syntactic
checklist. Examples the developer has flagged as subtle in
practice:

- **Low-density vertical scroll.** Rendered output that produces
  a tall scrollable region with very little content per line is
  usually a rendering failure — excess blank lines, fragmented
  paragraphs, orphaned punctuation on its own line, or wrap
  columns too narrow for the content. Sometimes it really is
  what the sender wrote, but it's a strong enough signal that
  the skill must compute density metrics and flag outliers.
- **Unmarked heading content.** Content that is visually a
  heading in the source — a bold standalone line, an all-caps
  line surrounded by whitespace, a `<p>` styled like an `<h2>`,
  a `<td>` used as a table header — should become `## heading`
  or `### heading` in the output even when no `<h2>`/`<h3>` tag
  exists. The pipeline must **infer structure from presentation**,
  not just echo semantic tags.

Any definition of "good markdown" that captures only syntactic
rules misses most of what actually goes wrong in real emails. The
reference needs principles, inference heuristics, and density
signals alongside the syntactic checks.

The existing `audit-output/` contains 45 pre-captured Gmail
renders, rendered-output only. They're usable as a seed but not
full-fidelity.

## Goal

Replace the broken loop with a first-class **training capture
system** that:

1. Lets the developer capture problematic email renderings from
   inside poplar itself, paired with a free-form comment describing
   the issue.
2. Stores captures persistently, cross-platform, outside the repo,
   in a canonical location.
3. Exposes a headless API that a Claude skill can use to list,
   inspect, re-render, diff, and mark captures — without
   hardcoding file paths or reimplementing the renderer.
4. Codifies the "well-crafted markdown" target as a **rigorous,
   first-class reference document** inside the skill's directory,
   with principles, structural inference rules, syntactic rules,
   density signals, and an evaluation procedure — so the skill
   scores captures against a single shared standard and the
   developer never re-explains what "looks good" means.
5. Produces a sanitizable path from a real PII-laden capture to a
   committable `testdata/` fixture + golden test + upstream patch,
   with a mandatory human-review gate on the sanitized artifact.
6. Leaves the display pipeline untouched. Markdown is an audit
   artifact, not a runtime intermediate — the lipgloss block
   styling path from ADR 0046 is unchanged.
7. Updates poplar's Claude infrastructure so future sessions
   naturally discover and use the system.

## Non-goals

- **Not a new viewer.** `poplar train` is a dev tool that reads raw
  email from a file path and runs it through the library packages.
  It does not talk to a live backend, walk folders, or depend on
  `internal/ui/App`. The Pass 2.5b-4 viewer, when it lands, will
  reuse the same capture-save function from its own `b` key.
- **Not a markdown-first pipeline.** `RenderMarkdown` is a pure
  audit function sibling to `RenderBody`. No viewer code consumes
  it. ADR 0046 stays intact.
- **Not a community corpus.** Captures stay local forever. Only
  sanitized `testdata/` fixtures ever enter the repo.
- **Not fully automated PII scrubbing.** The skill proposes a
  minimized reproducer; the developer reviews and finalizes
  sanitization by hand. Automation proposes, humans approve.
- **Not a replacement for unit tests.** Captures do not become the
  test suite. They become the *source* of test fixtures, which
  then live under `testdata/` and run under `make check` like any
  other table-driven test.

## Architecture

Four pieces, each with one clear job.

### 1. `poplar train` — cobra subcommand

A new subcommand mounted on the existing `poplar` root command. One
binary, one code path for rendering — it calls `internal/filter`
and `internal/content` directly, the same packages the viewer will
use later. It has both an interactive TUI (default when run with
no args) and a suite of headless subcommands for the skill.

The subcommands:

```
poplar train                          # interactive TUI
poplar train list [--status <s>]      # TSV table for the skill
poplar train show <id>                # full capture dump
poplar train render <id>              # styled render, stdout
poplar train markdown <id>            # canonical markdown, stdout
poplar train diff <id>                # current render vs snapshot
poplar train capture <path> \
    [--comment <text>]                # non-interactive capture
poplar train status <id> <state>      # new|triaged|fixed|wontfix
poplar train migrate [--confirm]      # one-shot legacy migration
poplar train extract-fixture <id> \
    [--out <path>] [--no-minimize] \
    [--force]                         # minimized reproducer
poplar train submit <id>              # gh pr create wrapper
```

All subcommands write to stdout, exit nonzero on error, and
structure output for both humans and the skill. The skill parses
`list` as tab-separated `id<TAB>status<TAB>platform<TAB>comment-preview`.

### 2. Capture store — plain directory tree

One directory per capture under `xdg.StatePath("poplar","captures")`,
which resolves to:

- Linux / macOS: `~/.local/state/poplar/captures/`
- Any OS with `$XDG_STATE_HOME` set: `$XDG_STATE_HOME/poplar/captures/`

The helper already exists in `internal/mailworker/xdg/xdg.go`
(forked from aerc). No new xdg code.

Fixed file names inside each capture directory:

```
~/.local/state/poplar/captures/
└── 20260412-a3b1c4/
    ├── raw.html        # or raw.txt for plaintext sources
    ├── rendered.ansi   # pipeline output at capture time (snapshot)
    ├── comment.md      # developer note, written via bubbles/textarea
    └── meta.toml       # platform, hash, timestamp, status, notes
```

ID format: `YYYYMMDD-<short-hash>` where short-hash is the first 8
hex chars of `sha256(raw)`. Collision-safe for any realistic
capture volume; human-readable in `ls`; date-sortable.

Permissions: `0o700` on the captures root and each capture dir;
`0o600` on files inside. Captures are PII by definition — locking
them to the owning user is the minimum viable defense.

Store mutability: **append-mostly**. The only post-creation
mutations are comment edits (via TUI or `poplar train capture`
re-invocation) and status changes via `poplar train status`.
Nothing in the system deletes a capture; removal is a manual
`rm -rf` of the directory. This keeps regression signal intact
even after fixes land.

### 3. `.claude/skills/fix-corpus/` — rewritten directory skill

Same name, new structure. The skill becomes a directory (matching
the `poplar-pass` pattern already in the repo):

```
.claude/skills/fix-corpus/
├── SKILL.md                       # frontmatter + workflow body
└── well-crafted-markdown.md       # normative reference doc
```

Triggered on phrases like "fix rendering bug", "review captures",
"mailrender issue", "email rendering broken", "triage corpus".

Consumes the capture store exclusively through `poplar train`
subcommands. Never reads `~/.local/state/poplar/captures/`
directly. Two reasons:

1. **Decoupling.** The skill stays independent of the on-disk
   layout. Future changes to the store format don't ripple into
   the skill.
2. **Single source of truth.** `poplar train` is the only place
   that runs the real pipeline. The skill can't accidentally
   evaluate against a stale or parallel renderer.

`SKILL.md` holds the trigger frontmatter and the workflow body
(the triage loop described in "Review flow" below). The workflow
body explicitly instructs the skill to load
`well-crafted-markdown.md` at the start of every triage pass and
evaluate every capture against it. The reference is normative
and versioned with the skill — editing it is the way the standard
evolves.

See the dedicated "Well-crafted markdown reference" component
below for the structure and content of `well-crafted-markdown.md`.

The skill workflow:

1. `poplar train list --status new` → the triage queue.
2. For each id: `poplar train show <id>` → read comment, classify
   the issue against the spec.
3. Group issues across the queue by root cause (holistic fixes,
   not one-at-a-time).
4. Present grouped issues to the developer for confirmation.
5. Edit `internal/filter` or `internal/content` to address each
   pattern. Add a unit test in the existing table-driven style.
6. For each affected id: `poplar train diff <id>` → confirm the
   fix is regression-neutral or improved.
7. `make check` — mandatory gate.
8. For each fixed id: `poplar train extract-fixture <id>`
   produces a minimized HTML reproducer. The skill scrubs
   remaining PII (real-looking emails, account numbers, tokens)
   by proposing a scrubbed version; the developer reviews and
   approves (unless `--trust` is set). The result lands as
   `testdata/filter/<slug>.html` or `testdata/content/<slug>.html`
   with a matching golden file.
9. `poplar train status <id> fixed` for each.
10. `/ship` runs `/simplify` → commit → push → install. External
    contributors use `poplar train submit <id>` instead, which
    wraps `gh pr create`.

### 4. Submit flow — sanitize, package, submit

The real friction in upstream contribution is PII. The submit flow
has two halves:

**Half A — sanitize.** `poplar train extract-fixture <id>` runs
the capture's raw through `internal/train/minimize/`, which strips
scripts, styles, hidden divs, and wrapper elements that don't
affect the rendered output, returning the smallest HTML that still
triggers the bug. The output is *not* scrubbed — it may still
contain PII. The skill then proposes a scrubbed version (names
replaced with Alice/Bob, emails with `@example.com`, URLs with
`example.com`, tracking tokens blanked) and the developer reviews.
For the primary maintainer's own captures, `--trust` skips the
review prompt; for outside contributors, review is mandatory.

**Half B — package.**

1. Skill writes the code fix to `internal/filter` or
   `internal/content`.
2. Skill writes a table-driven test alongside the fixture.
3. `make check` runs. If it fails, the fix is incomplete; loop.
4. Primary maintainer path: `/ship` handles commit + push +
   install.
5. Contributor path: `poplar train submit <id>` detects `gh`,
   opens a PR against the poplar repo with a template-filled body
   that references the capture id, the sanitized fixture path,
   and the fix summary. If `gh` is missing, the command exits
   nonzero with an install hint — no fallback to `git format-patch`
   or `git send-email`.
6. On success, the capture's `meta.toml` gets `status = "fixed"`
   and optionally a `fixture =` pointer to the `testdata/` file.

## Components

### `internal/train/` (new package)

Owns capture storage, the TUI model, migration, and submit
helpers. Pure Go, no side effects except on-disk I/O scoped to
`xdg.StatePath("poplar","captures")` and (for `extract-fixture`)
the project's `testdata/` tree.

**Types:**

```go
type Capture struct {
    ID        string    // "20260412-a3b1c4"
    Dir       string    // absolute path to the capture dir
    RawPath   string    // raw.html or raw.txt
    Meta      Meta
    Comment   string    // loaded lazily
}

type Meta struct {
    CreatedAt    time.Time
    SourceKind   string // "html" | "plain" | "rendered-only"
    SourceHash   string // sha256(raw)[:8]
    Platform     string // "gmail" | "outlook" | "yahoo" | "fastmail" | "unknown"
    Status       string // "new" | "triaged" | "fixed" | "wontfix" | "broken"
    FixtureRef   string // optional path into testdata/ after sanitize
    Notes        string // free-form skill notes
    RenderedOnly bool   // true for migrated audit-output entries
}
```

**Functions (pure, table-driven-testable):**

```go
func Root() string                                    // xdg.StatePath wrapper
func List(root string) ([]Capture, error)             // WalkDir + parse meta.toml
func Load(root, id string) (Capture, error)
func Save(root string, raw []byte, srcKind, comment string) (Capture, error)
func Render(c Capture) (string, error)                // filter + content
func Markdown(c Capture) (string, error)              // content.RenderMarkdown
func Diff(c Capture) (string, error)                  // current vs rendered.ansi
func Migrate(root, audit, corpusLegacy string, confirm bool) (MigrateReport, error)
func UpdateStatus(root, id, status string) error
func ExtractFixture(c Capture, out string, minimize bool, force bool) (string, error)
```

**TUI model** lives in `internal/train/tui/` — a standard
bubbletea `tea.Model`. It does not live in `internal/ui/` because
`internal/ui/` is governed by Elm conventions tied to the app
root; `poplar train` must be launchable independently. The model
still follows Elm discipline by convention for consistency.

### `internal/train/minimize/` (new sub-package)

```go
func Minimize(html []byte) ([]byte, error)
```

Strips scripts, styles, meta tags, hidden nodes, and wrapper
containers that don't affect the rendered output. Returns the
shortest HTML the production pipeline would render identically (or
as close as possible). Uses the golang.org/x/net/html tokenizer —
no new dependencies.

**Not a sanitizer.** Output may still contain PII. Sanitization
happens after minimization, proposed by the skill and reviewed by
the developer.

### `internal/content/markdown.go` (new file)

One function, sibling to `RenderBody`:

```go
func RenderMarkdown(blocks []Block) string
```

Deterministic canonical markdown output from the existing block
tree. No new block types, no parser changes, no flags. A walk that
emits markdown syntax per block:

- `Paragraph` → plain text + `\n\n`
- `Heading{Level: N}` → N hashes + space + text
- `List{Ordered}` → `1. item\n`, nesting via indent
- `List{Unordered}` → `- item\n`, nesting via indent
- `Blockquote{Level: N}` → `> ` repeated, children recursed
- `Code{Inline}` → backticks
- `Code{Block, Lang}` → fenced block with language tag
- `Link` → `[text](href)`, footnote-style when the renderer
  would footnote it
- `HR` → `---\n`

Symmetric with `RenderBody` — same input, different output
transform. Unit-tested with table-driven tests mirroring
`render_test.go`.

### `cmd/poplar/train.go` (new file)

Thin cobra wiring. Each subcommand is 5–15 lines: flag parsing,
call into `internal/train`, format output, exit. No business
logic in the command layer — all the work happens in
`internal/train/`.

Uses the existing root command from `cmd/poplar/root.go`. The
train subcommand is mounted alongside `themes` and `config init`.

### `.claude/skills/fix-corpus/SKILL.md` (rewritten)

Fully replaces the previous single-file `fix-corpus` skill. The
new `SKILL.md` holds the frontmatter (trigger phrasing, skill
name, description) and the workflow body: the triage loop, the
fix loop, and explicit instructions to load
`well-crafted-markdown.md` at the start of every triage pass.
Triggers on rendering-bug phrasing. Uses `poplar train`
exclusively — never reads the capture store directly.

The reference doc is a sibling file in the same directory; see
the following component entry.

### `.claude/skills/fix-corpus/well-crafted-markdown.md` (new — normative reference)

This is the single document that defines what "well-crafted
markdown" means for the poplar pipeline. It is the standard the
skill evaluates every capture against, and the thing the
developer never has to re-explain. It is versioned, committed,
and updated in place as the understanding of good output evolves.

**Structure.** Five sections, each load-bearing.

#### §1 Principles

Three to five paragraphs of design philosophy — the "why" behind
the rules. Principles matter because rules alone cannot cover
every case, and when the skill encounters ambiguity it defers to
principles. Examples of the principles the reference encodes:

- **Density is signal.** A well-rendered email packs meaningful
  content per line. Long vertical scrolls with fragmentary lines
  are a rendering failure more often than not — even when the
  source is objectively short.
- **Structure is inferred, not copied.** HTML source tags are
  hints, not directives. A `<p>` with bold styling and a line
  break on each side is a heading regardless of the tag name.
  The pipeline crafts markdown structure from the visual
  behavior of the source, not from a one-to-one tag translation.
- **Output is for humans in a narrow column.** The target is a
  terminal reader with an 80-column window, not a browser or a
  web email client. Markdown that looks good at 80 columns wins
  over markdown that "matches" the source's 800-pixel layout.
- **Consistency beats cleverness.** A boring, predictable
  translation is better than a smart one that varies. If two
  similar emails render differently, the renderer is wrong even
  if both renders individually "look fine."
- **Failures are diagnosable.** When the output is bad, the skill
  must be able to point at a specific rule or metric that failed,
  not just "it looks off." Every rule in §3 has an observable
  failure mode.

#### §2 Structural inference rules

Heuristics for deriving markdown structure from presentation when
the source lacks semantic tags. These describe *what the pipeline
should do*, not *what the output should look like* — violations
mean the filter or content parser needs to get smarter, not that
a post-processing rule is missing.

Each rule follows the shape:

> **Rule name.** Statement of the inference. Concrete source
> pattern that triggers it. The markdown element it should
> produce. Why it matters.

Categories and representative rules:

- **Heading inference.**
  - *Bold standalone line → heading.* A `<p>`, `<div>`, or `<td>`
    whose only text content is wrapped in `<b>`/`<strong>`, has
    blank content before and after, and is less than ~60
    characters long, becomes `## heading`. Depth is inferred
    from nesting or font-size hints if present, defaulting to 2.
  - *All-caps standalone line → heading.* A standalone line that
    is entirely uppercase (with punctuation allowed) and shorter
    than ~60 characters becomes `## heading`. This catches the
    Microsoft-style "SECTION TITLE" pattern.
  - *Line followed by same-character underline → heading.* Plain
    text sources where a line is followed by a line of `=====`
    or `-----` of similar length become `#` or `##` headings
    respectively (Setext-style, rendered as ATX-style in the
    output).
  - *Table-header `<td>` → heading inside the row's content.*
    A `<td>` bearing `<th>` semantics or bold-styled first row
    content becomes a heading preceding the table's text
    content.
- **Paragraph reconstruction.**
  - *Hard-wrapped source → join into logical paragraphs.* Plain
    text emails with 72-column hard wraps should be joined on
    whitespace so the output reflows to the terminal column,
    not the sender's column.
  - *`<br>` runs → paragraph breaks or single newlines.* Two or
    more consecutive `<br>` tags become a paragraph break. A
    single `<br>` inside flowing text is collapsed to a space
    unless the surrounding context is a known line-oriented
    structure (address block, signature).
- **List detection without tags.**
  - *Lines beginning with `-`, `*`, `•`, or a number + dot, at
    consistent indent, in three or more consecutive lines →
    unordered or ordered list.* Applied to plain text sources
    and to `<p>`-wrapped line-break lists.
- **Signature separation.**
  - *`-- ` on its own line → horizontal rule + signature block.*
    The sig-sep convention. Everything after is treated as a
    block-level signature and may be rendered at reduced
    emphasis.

Each rule names the inference the pipeline must make, not the
markdown artifact per se — the output artifact is covered by §3.

#### §3 Syntactic rules

Mechanical checks on the rendered markdown string. Each rule is
RFC 2119 style (MUST / SHOULD / MAY), with a detection procedure
and a representative bad/good pair. Failures are cheap to
diagnose and usually cheap to fix.

**Organized by category:**

- **Entities (MUST).** No literal HTML entities in the output
  (`&amp;`, `&nbsp;`, `&#8217;`, `&ldquo;`, etc.). Detection:
  regex `&[a-zA-Z]+;|&#\d+;|&#x[0-9a-fA-F]+;`. Fix location:
  usually `internal/filter/` (entity decoding pass missed).
- **Whitespace hygiene (MUST).**
  - No trailing whitespace on any line. Detection: regex ` +$`.
  - No leading whitespace on paragraph text (list items and code
    blocks excepted). Detection: walk the block list and check.
  - No runs of three or more consecutive blank lines. Detection:
    regex `\n\n\n`. Two blank lines are the maximum and only
    allowed as a block separator.
- **Orphaned punctuation (MUST).** No line whose content is
  solely punctuation or a single short fragment like `.` or
  `?` or `— ` — a strong sign of paragraph fragmentation.
  Detection: walk block list; paragraphs containing only
  punctuation or under 4 characters fail.
- **Heading hygiene (MUST).**
  - No skipped levels: `#` followed directly by `###` is a
    violation. Detection: linear scan, track max depth seen.
  - Blank line above and below every heading. Detection: block
    walk.
  - Heading text has no trailing colon (`:`) unless the source
    clearly had one — colons signal "list header" which is a
    different construct.
- **List formatting (SHOULD).**
  - Consistent marker per list: all `-` or all `*`, never mixed.
  - Ordered lists use `1.` numbering (markdown convention;
    renderers fill in the sequence).
  - Nested items indent by exactly two spaces per level.
  - Wrapped continuation lines have a hanging indent matching
    the text start column.
- **Blockquote wrapping (MUST).**
  - Every quoted line prefixed with `> ` (space after the
    marker).
  - Attribution line (`On X, Y wrote:`) precedes quoted content
    in the same block; if the quoted content is unquoted in the
    source (implicit quote), the filter wraps it in a
    `Blockquote{Level: 1}` at top level only — the fix from
    BACKLOG #7.
  - Nested blockquotes use `> > ` with a space between markers,
    not `>>`.
- **Link handling (SHOULD).**
  - Footnote-style links use sequential numbering starting at 1
    with no gaps. Detection: extract `[^N]` references and
    `[^N]:` definitions; verify the set is contiguous.
  - Tracking parameters stripped from hrefs: `utm_*`, `fbclid`,
    `gclid`, `mc_cid`, `mc_eid`, `_hsenc`, `_hsmi`. Detection:
    URL parse; fix location: `internal/filter/`.
  - Link text is not a bare URL when the source provided anchor
    text distinct from the href. Bare URLs are still allowed
    when the source gave nothing else.
  - Footnote block appears at the end of the body, separated
    from content by exactly one horizontal rule `---`.
- **Code (SHOULD).**
  - Inline code uses single backticks with no leading or trailing
    space inside the backticks.
  - Multi-line code uses triple-backtick fenced blocks.
  - Language tag preserved on fenced blocks when the source
    declared one (`<pre><code class="language-python">`).
  - Indented code blocks are never produced.
- **Horizontal rules (MUST).** Exactly `---` on its own line,
  never `***` or `___`.
- **Emphasis (MAY).** `_italic_` for emphasis; `**bold**` for
  strong. Style may vary if the source mandates; consistency
  within a single output is the requirement.

Each rule includes a one-line "probable fix location" hint
(`internal/filter/` vs `internal/content/`) so the skill can
route fixes to the right layer without guessing.

#### §4 Density signals

Observable metrics computed from the rendered output. These are
not rules — they're smoke detectors. A failing signal doesn't
automatically mean the output is wrong, but it means the skill
should inspect the capture carefully and ask whether the source
really justifies the metric or whether the pipeline is
fragmenting content.

Metrics the skill computes via `poplar train markdown <id>`
plus a walk of the output:

- **Lines per paragraph (mean, median, p90).** Low values (most
  paragraphs under 2 lines) flag fragmentation, especially when
  paired with a short source. Expected: prose paragraphs run
  3–10 lines at 80 columns.
- **Characters per non-blank line (mean, p50, p90).** Low values
  under ~40 characters across most lines flag short-line output
  — probably a wrap-column mismatch or premature line breaks.
- **Blank-line ratio.** Ratio of blank lines to non-blank lines
  in the rendered output. Ratios above ~0.5 flag excessive
  spacing.
- **Block count per 100 characters of source.** A high block
  count relative to source size flags over-fragmentation — the
  pipeline is splitting content that should stay joined.
- **Vertical extent.** Total line count of the rendered output
  vs character count of the source. A plain 500-character source
  producing 80 lines of rendered output is suspicious.
- **Orphan rate.** Count of paragraphs under 10 characters as a
  fraction of total paragraphs. High rates flag fragmentation.
- **Heading density.** Count of headings as a fraction of total
  blocks. Zero headings in a 40-block document is suspicious —
  probably missed heading inference. Many headings in a short
  document is also suspicious — probably over-eager inference.

The skill does not fail a capture on any single metric. It uses
metrics to prioritize manual inspection, and to detect
regressions between `rendered.ansi` snapshots and current
renders in `poplar train diff`.

#### §5 Evaluation procedure

The deterministic procedure the skill follows when scoring a
capture. Codified so two different sessions score the same
capture the same way.

1. `poplar train show <id>` — load raw, comment, current
   markdown, current rendering, meta.
2. Load `well-crafted-markdown.md` (§1–§4).
3. Compute §4 metrics from the current markdown. Record any
   outliers against the documented thresholds.
4. Walk the current markdown and apply §3 syntactic rules. For
   each violation, record: rule name, location in output, likely
   fix layer.
5. Walk the raw source and apply §2 structural inference rules.
   For each expected inference that did not happen in the
   current markdown, record: rule name, source pattern, expected
   markdown output.
6. Read the developer comment. Weight the comment's complaints
   against the rule violations — if the comment says "the list
   wraps badly" and §3 list-formatting rules flag wrapping
   issues, that's high confidence. If the comment and the
   automated checks disagree, flag for manual inspection.
7. Apply §1 principles to any ambiguous case. When principles
   conflict with a specific rule, principles win and the rule
   is flagged as a candidate for revision.
8. Produce a scoring report per capture: `{id, rule violations,
   metric outliers, inference misses, principle conflicts,
   comment alignment}`.
9. Aggregate across the triage queue. Group by root cause (same
   rule violated across multiple captures → single fix target).
10. Proceed to the fix loop.

The procedure is the skill's internal algorithm, but documenting
it here means a second Claude session — or a future skill
revision — produces the same judgments for the same inputs. That
consistency is the whole point of having a normative reference.

---

### `docs/poplar/training.md` (new, on-demand)

Matches the loading pattern of `styling.md` / `wireframes.md` /
`keybindings.md`. Roughly 200 lines. Sections:

- Overview — what the training loop is and who it's for.
- Capture location — `xdg.StatePath`, cross-platform paths, env
  override.
- Capturing — the TUI flow, keys, comment entry.
- Inspecting — `list`, `show`, `render`, `markdown`, `diff`.
- PII policy — captures are local, sanitized fixtures only.
- Submitting a fix — `extract-fixture`, manual sanitize
  checklist, golden tests, `submit` or PR.
- Migration — one-shot `poplar train migrate`.
- Well-crafted markdown — a pointer to
  `.claude/skills/fix-corpus/well-crafted-markdown.md`, which is
  the single normative source. `docs/poplar/training.md` never
  duplicates the spec content; it only explains where to find
  it, how to propose changes, and how the skill loads it.

### `docs/poplar/decisions/0059-training-capture-system.md` (new)

ADR codifying the decisions in this spec. See Section "Infrastructure
updates" below.

## Data flow

### Capture flow

```
developer opens poplar train
       │
       ▼
list of captures rendered from xdg.StatePath("poplar","captures")
       │  press b
       ▼
TUI prompts for a source file path (textinput)
       │
       ▼
internal/train.Save(raw, kind, comment="")
       ├─► sha256(raw)[:8] → source hash
       ├─► detect kind: ".html" → html, else plain
       ├─► detect platform: regex over first 4KB
       ├─► id = YYYYMMDD-<hash>
       ├─► mkdir ~/.local/state/poplar/captures/<id>
       ├─► write raw.(html|txt), meta.toml (status="new")
       └─► run pipeline once, write rendered.ansi snapshot
       │
       ▼
TUI opens bubbles/textarea for the comment
       │  Enter saves, Esc cancels
       ▼
write comment.md; refresh list; selection jumps to new entry
```

When the Pass 2.5b-4 viewer lands, its `b` key calls the same
`internal/train.Save` — the raw comes from the message viewer's
already-loaded body instead of a file picker. One code path, two
entry points.

### Review flow (skill)

```
skill: poplar train list --status new
       │
       ▼
for each id:
    poplar train show <id>
       │
       ▼
skill classifies against the well-crafted-markdown spec
       │
       ▼
skill groups issues across the queue by root cause
       │
       ▼
skill edits internal/filter or internal/content; adds unit test
       │
       ▼
for each affected id:
    poplar train diff <id>
       │
       ▼
make check
       │
       ▼
for each fixed id:
    poplar train extract-fixture <id>
      → skill proposes scrubbed HTML
      → developer reviews (unless --trust)
      → skill writes testdata/<area>/<slug>.html + golden
    poplar train status <id> fixed
       │
       ▼
/ship (primary maintainer) OR poplar train submit <id> (contributor)
```

## Error handling

Training is a dev tool. Errors surface immediately; no silent
fallbacks, no retry loops, no best-effort partial results.

- **Missing captures dir.** Auto-create on read via `os.MkdirAll`
  with `0o700`. First-run UX is "no captures yet" in the TUI.
- **Corrupt `meta.toml`.** `List` skips the bad capture and logs
  the path + decode error to stderr. The TUI header shows the
  skipped count. One bad file does not hide good ones.
- **Missing `raw.html`/`raw.txt`.** The entry is listed with
  status `broken`. `render`/`markdown`/`diff` on broken entries
  exit nonzero with a clear message.
- **Renderer errors** (panic or error from `filter.CleanHTML` /
  `content.ParseBlocks`). Caught at the `poplar train
  render`/`markdown` boundary, logged to stderr with the capture
  id, non-zero exit. The skill treats a render error as a
  first-class bug to fix — the capture becomes the reproducer.
- **Save collisions** (duplicate id — theoretically impossible).
  `Save` returns an error; `capture` exits nonzero; TUI shows a
  toast; the existing capture stays untouched. No auto-overwrite.
- **`extract-fixture` write conflict.** Refuses to overwrite an
  existing `testdata/` file unless `--force`.
- **`submit` with missing `gh`.** Detect upfront, print one-line
  install hint, exit nonzero. No fallback.

## Migration of legacy files

One-shot command: `poplar train migrate`. Idempotent; safe to
re-run.

**Targets:**

- `corpus/salmon-selvedge.html` — 1 file, raw HTML available.
- `audit-output/G*.txt` — 45 files, rendered output only, no raw
  source.

**Outcomes:**

- **`salmon-selvedge.html`** is a full-fidelity capture. `raw.html`
  copies the source. The pipeline runs to produce a fresh
  `rendered.ansi`. `comment.md` is seeded with "Migrated from
  legacy corpus/. No original comment." `meta.toml` gets
  `status = "triaged"` (the bug is already fixed — BACKLOG #7)
  and a note tagging the migration.
- **`audit-output/G*.txt`** files are **degraded** captures. No
  raw source available, so `raw.txt` holds the rendered ANSI;
  `rendered.ansi` holds the same content. `meta.toml` gets
  `source_kind = "rendered-only"` and `rendered_only = true`.
  `status = "unscored"`. The flag tells the skill that `render` /
  `markdown` / `diff` are meaningless for these captures — the
  skill can only inspect the stored ANSI and the user comment.
  The skill's first pass over them is "discover issues visually
  and annotate."

**Preflight.** `migrate` without `--confirm` is a dry run — it
lists what it would do and exits 0. With `--confirm`, it executes
the moves.

**Post-migration deletion.** After a successful migration, the
same commit that drops the training code also:

- Deletes `corpus/` and `audit-output/` from the repo.
- Adds both directories to `.gitignore`.
- Whitelists `testdata/` explicitly so no future gitignore rule
  hides sanitized fixtures.

**Rollback.** None needed. The legacy files are content-identical
to what migration writes. If something goes wrong, `rm -rf
~/.local/state/poplar/captures` and re-run.

## Testing

Everything in `internal/train/` is unit-testable with plain
`t.TempDir()` scoped capture roots.

- `internal/train/store_test.go` — `Save` → `List` → `Load`
  roundtrip; collision handling; corrupt-meta skipping; status
  updates.
- `internal/train/migrate_test.go` — dry-run output; full
  migration with a fixture legacy tree in `testdata/`;
  idempotency (running twice is a no-op).
- `internal/train/minimize/minimize_test.go` — table-driven:
  input HTML → expected minimized output, plus an assertion that
  the pipeline produces identical rendered output before and
  after minimization.
- `internal/content/markdown_test.go` — table-driven: input block
  trees → expected markdown, covering each block type and
  nesting.
- `cmd/poplar/train_test.go` — cobra command wiring tests. Each
  subcommand parses flags, handles missing ids, exits with the
  right code. Uses a `t.TempDir()` capture root via a hidden flag
  or env var.

No mocks. The filter/content packages are already pure;
`internal/train/` touches the filesystem only under `t.TempDir()`.
Submit helpers hide behind a small interface and are tested with
a fake; the real implementation shells out to `gh` directly.

No integration tests against real `gh`; no network; no live
`poplar train submit` in CI. Those are manual.

## Infrastructure updates

These edits land in the same pass as the code drop so future
sessions discover the system naturally.

### `CLAUDE.md`

One new line under "On-demand reading":

> `docs/poplar/training.md` — capture store + fix-corpus loop.
> **Load when a user reports a rendering bug or asks to review
> captures.**

No other changes; `CLAUDE.md` stays under the 200-line hook limit.

### `docs/poplar/invariants.md`

Four new bullets plus one decision-index row:

- **Architecture.** The training capture store lives at
  `xdg.StatePath("poplar","captures")`. Never in-repo, never in
  git. Read exclusively by `internal/train/` via `poplar train`
  subcommands. Captures contain PII; only sanitized fixtures
  under `testdata/filter/` and `testdata/content/` are
  committable.
- **Architecture.** `internal/train/` owns the capture store, the
  `poplar train` TUI, and migration. `internal/content/RenderMarkdown`
  is an audit sibling to `RenderBody` — markdown is **not** a
  pipeline intermediate. ADR 0046's lipgloss display path is
  unchanged.
- **UX.** `poplar train` is a dev tool, not user-facing. It does
  not obey the vim-first / no-modifiers rules as a user-visible
  invariant, but does so by convention for consistency.
- **Build & verification.** The `fix-corpus` Claude skill is the
  authoritative loop for renderer bug fixes. Ad-hoc renderer
  edits without a capture and a sanitized `testdata/` fixture are
  not acceptable — the skill enforces the loop.
- **Decision index, new row.**
  `| Training capture system, markdown as audit artifact | 0059 |`

### `docs/poplar/system-map.md`

- **Package layout, new row.**
  `| internal/train/ | Capture store (xdg state), poplar train TUI,
  migration, HTML minimizer. Consumers: cmd/poplar/train.go,
  .claude/skills/fix-corpus. |`
- **Package layout, amend `internal/content/` row.** Append:
  "`RenderMarkdown` is an audit sibling to `RenderBody`, consumed
  by `poplar train markdown` and `fix-corpus`."
- **Binary section, amend.** Note that `cmd/poplar/` now includes
  the `train` subcommand.
- **Docs section, new line.**
  `docs/poplar/training.md` — capture store, poplar train
  workflow, PII policy, fix-corpus loop.

### `docs/poplar/STATUS.md`

Add a new "Tooling" section below the pass table:

```markdown
## Tooling

- **Training capture system** — `poplar train` + `internal/train/` +
  `fix-corpus` skill. Built out-of-band from the pass sequence.
  Provides the authoritative loop for renderer bug-fixing. See
  `docs/poplar/training.md`.
```

### `.claude/skills/`

- **`fix-corpus`** is restructured from a single file to a
  directory skill matching the `poplar-pass` pattern:
  `SKILL.md` holds the frontmatter and workflow body;
  `well-crafted-markdown.md` holds the normative reference
  (principles, inference rules, syntactic rules, density
  signals, evaluation procedure — see the component description
  above). `SKILL.md` instructs the skill to load the reference
  at the start of every triage pass.
- **`poplar-pass`** gains one pass-end checklist item: "If any
  training captures were touched this pass, update their status
  via `poplar train status <id> <state>`."
- No new skills.

### `docs/poplar/decisions/0059-training-capture-system.md`

New ADR. Cross-refs ADR 0001 (monorepo), 0046 (lipgloss pivot),
0058 (post-pivot structure), BACKLOG #7.

**Context.** The pre-pivot `corpus/` + `mailrender save` + aerc
loop is broken. `audit-output/` accumulated 45 rendered-only
files with no comments. The ad-hoc approach was lossy: by the
time a fix was attempted, the original observation was gone. The
"well-crafted markdown" goal lived in the developer's head and
was re-explained on every fix.

**Decision.** Build a first-class training capture system.
Captures live in `xdg.StatePath("poplar","captures")`, one
directory per capture, holding raw source + capture-time
rendering + developer comment + meta. `poplar train` is both the
interactive capture tool and the headless harness the `fix-corpus`
skill uses. The skill consumes captures through `poplar train`
exclusively. Markdown is introduced as an audit artifact via
`RenderMarkdown` sibling to `RenderBody` — deliberately not as a
pipeline intermediate.

**Consequences.**

- PII is contained. Captures never enter the repo.
- The fix loop is reproducible. The skill always runs the real
  pipeline. "Well-crafted markdown" becomes an objective target
  the skill evaluates against without re-asking the developer
  what "looks good" means.
- The 46 legacy files migrate in one shot. `corpus/` and
  `audit-output/` disappear from the repo.
- New tool surface adds maintenance cost: one cobra subcommand,
  one new internal package, one new renderer function, one
  rewritten skill, one new doc.

**Alternatives considered.**

- *Keep legacy `corpus/` + `mailrender` binary.* Stale post-pivot;
  no comments; poor skill loop.
- *Markdown as a real pipeline intermediate (B2).* Reverses ADR
  0046; churns the display pipeline for no display win.
- *In-repo capture dir at `.poplar/captures/`.* PII in git is a
  one-`git add`-away disaster.
- *Fully automated PII scrubbing.* Risky; can't be trusted.

### `.gitignore`

Add `corpus/` and `audit-output/` after migration. Prevents
accidental rebirth. Explicitly whitelist `testdata/` so no future
rule hides sanitized fixtures.

### Hooks

None. Considered a pre-commit hook scanning `testdata/*.html` for
PII patterns (real-looking emails outside `@example.com`, long
hex tokens). Rejected for v1 — noise-prone, and the human-review
gate on `extract-fixture` is the real defense. Can be added later
as a BACKLOG item if a leak ever occurs.

### Auto-memory

One new memory to write at implementation start (not before),
type `project`:

> **Training capture system.** `poplar train` + `internal/train/`
> + `fix-corpus` skill are the authoritative loop for renderer
> bug fixes. Captures live in `xdg.StatePath("poplar","captures")`,
> never in repo. Markdown is an audit artifact only.
> **Why:** replaces the stale pre-pivot corpus/mailrender flow
> and gives the fix-corpus skill an objective "well-crafted
> markdown" target.
> **How to apply:** when a rendering bug is reported, start with
> `poplar train capture <path>`, write a comment, and drive fixes
> through the skill — don't hand-edit the renderer from a single
> example.

## Deliverables

Tracked for the implementation plan:

1. `internal/train/` — new package: store, TUI model, migration,
   submit helpers.
2. `internal/train/minimize/` — new sub-package: HTML subtree
   minimization.
3. `internal/content/markdown.go` — new file: `RenderMarkdown` +
   table-driven tests.
4. `cmd/poplar/train.go` — new cobra subcommand mounting the
   `train` TUI entry point plus 10 named child subcommands
   (`list`, `show`, `render`, `markdown`, `diff`, `capture`,
   `status`, `migrate`, `extract-fixture`, `submit`).
5. `.claude/skills/fix-corpus/SKILL.md` — rewritten skill as a
   directory; frontmatter + workflow body that loads the
   reference doc and drives the triage/fix loop.
6. `.claude/skills/fix-corpus/well-crafted-markdown.md` — new
   normative reference with the five-section structure
   (principles, inference, syntax, density, procedure).
7. `.claude/skills/poplar-pass` — add pass-end capture status
   update.
8. `docs/poplar/training.md` — new on-demand reference.
9. `docs/poplar/decisions/0059-training-capture-system.md` — new
   ADR.
10. `docs/poplar/invariants.md` — 4 new bullets + 1
    decision-index row.
11. `docs/poplar/system-map.md` — package row, content-package
    note, doc-list line.
12. `docs/poplar/STATUS.md` — new Tooling section.
13. `CLAUDE.md` — one on-demand-reading line.
14. `.gitignore` — add `corpus/` and `audit-output/`.
15. **Deletions.** `corpus/` and `audit-output/` wiped in the
    same commit as the migration code.

## Commit plan

The spec lands on its own first (docs-only; no `/simplify`
needed). The implementation plan that writing-plans produces
will split the code drop into commits roughly as:

1. `internal/content/markdown.go` + tests.
2. `internal/train/` store + tests.
3. `internal/train/minimize/` + tests.
4. `internal/train/` TUI + tests.
5. `cmd/poplar/train.go` subcommand wiring.
6. `poplar train migrate` + legacy deletion in one commit.
7. `.claude/skills/fix-corpus/` rewrite as a directory:
   `SKILL.md` + `well-crafted-markdown.md`. The reference doc is
   written in full in this commit — it is the load-bearing
   artifact, not a stub.
8. Docs + invariants + ADR + system-map + STATUS + CLAUDE.md
   line.

`make check` runs on every commit; `/ship` runs at the end.

## Cross-references

- ADR 0001 — single-repo, single-binary architecture.
- ADR 0046 — lipgloss block model replacing glamour; the display
  pipeline this design leaves untouched.
- ADR 0058 — the 2026-04-12 pivot that broke the legacy
  `corpus/`+`mailrender` loop.
- BACKLOG #7 — first blockquote wrapping fix; the canonical
  example of what the skill loop produces when working well.
- `docs/poplar/invariants.md` — updated with 4 new bullets plus
  decision-index row pointing here.
- `internal/mailworker/xdg/xdg.go` — existing cross-platform
  path helper; `StatePath` is the function `poplar train` uses.

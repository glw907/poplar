# Audit-2 Findings: Library Package Shape

> Run on 2026-04-25 against
> [audits/2026-04-25-library-packages.md](2026-04-25-library-packages.md).
> Each section answers the audit's four questions: current shape,
> intended consumer, fit assessment, verdict + recommendation.

## Summary

| Package | Verdict | Action |
|---|---|---|
| `internal/filter/` | keep (with delete sub-items) | Drop the unused reflow family + the empty `headers.go` placeholder. Fix the stale aerc package doc. CleanHTML/CleanPlain/ToHTML are the real consumer surface and they're well-shaped. |
| `internal/content/` | keep | Already partially consumed in Pass 2.5b-4 and the API matches. Minor ergonomic note on dead `blockKind`/`spanKind` enums. |
| `internal/tidy/` | collapse | Core algorithm (`SplitQuoted`/`Reassemble`/`BuildPrompt`/`CallAPI`/`Tidy`) fits Pass 9.5 cleanly. The TOML loader + CLI override parsers + `ConfigString` pretty-printer are leftovers from a standalone-CLI shape — drop them when Pass 9.5 lands. |

## `internal/filter/`

### Current shape

The package has three loosely-related subsystems, all unexported except
for the entry-point functions:

- **HTML→markdown cleanup** (`html.go`, `convert.go`):
  `CleanHTML(string) string` runs a goldmark/JohannesKaufmann
  html-to-markdown pipeline with two custom plugins (`imageStripPlugin`,
  `layoutTablePlugin`) and a chain of post-conversion passes
  — Mozilla-class scrubbing, hidden-`<div>` stripping, zero-size image
  removal, NBSP / zero-width filler collapse, paren-style list marker
  normalization, deduplication keyed on link-text-only, empty-link
  removal, short-block "·"-joining, attribution-line quote
  unflattening for Outlook-mobile, and signature compaction. ~510 LOC.
- **Plain-text cleanup** (`plain.go`): `CleanPlain(string) string` —
  detects HTML inside `text/plain` and routes it through `CleanHTML`,
  otherwise unescapes entities and de-tabs list markers. ~25 LOC.
- **Markdown→HTML** (`tohtml.go`): `ToHTML(io.Reader, io.Writer) error`
  — wraps a goldmark+table render in a minimal HTML document
  envelope. Replaces pandoc in aerc's multipart converters.
- **Paragraph reflow** (`html.go::reflowMarkdown` and friends):
  minimum-raggedness DP wrapper, blockquote-aware. **Exported nothing,
  called by nothing in production — only invoked from
  `html_test.go`.**
- `headers.go` and `headers_test.go` are empty placeholders
  (`package filter` and nothing else).

Package doc comment reads: *"Package filter implements aerc email
content filters."* That's stale — these are not aerc filters; they are
poplar's own content-cleanup pipeline.

### Intended consumer

Two distinct consumers, both deferred:

- **Pass 3 (wire prototype to live backend)** will need `CleanHTML` and
  `CleanPlain` in the body-fetch path. Today `cmds.go::loadBodyCmd`
  reads `mail.Backend.FetchBody` bytes and feeds them straight into
  `content.ParseBlocks`, which assumes pre-cleaned markdown. The mock
  backend in `internal/mail/mock.go` returns pre-cleaned markdown so
  the gap is invisible. Real IMAP/JMAP bodies are `text/html` or
  raw `text/plain`, so the gap will surface immediately.
- **Pass 9 (compose + send)** will need `ToHTML` for the
  `multipart/alternative` HTML half of outbound mail.

### Fit assessment

- **CleanHTML/CleanPlain are well-shaped for Pass 3.** Both are pure
  `string → string`. The consumer is whatever code stands between
  `FetchBody` and `ParseBlocks` — a one-line dispatch on MIME type.
  The pipeline encodes serious email-quirk knowledge (Mozilla
  attributes, Outlook-flat quotes, layout tables, tracking pixels,
  signature compaction) that took real iteration to develop. Throwing
  this away to rebuild later would be expensive.
- **The wiring gap is at the *boundary*, not in the package.**
  `mail.Backend.FetchBody(uid) (io.Reader, error)` doesn't expose a
  MIME type, so the call site has nothing to dispatch on. Pass 3 will
  need either (a) a `FetchBody` signature change to return MIME, or
  (b) a sniff helper (`detectHTML` already exists privately inside
  `plain.go` — could be promoted). This is a Pass-3 wiring concern,
  not a filter shape problem.
- **`ToHTML` is well-shaped for Pass 9.** `io.Reader → io.Writer` with
  a sealed HTML envelope. Small, no dependencies on the cleanup
  subsystem.
- **`reflowMarkdown` and friends are dead code.** They were written
  when filter was responsible for fixed-width paragraph wrapping.
  That responsibility moved into `content/render.go` when the
  lipgloss block model landed in Pass 2.5-render —
  `content.RenderBody` calls `ansi.Wordwrap` on each block at render
  time, with width determined by the panel. Filter's reflow pass is
  no longer in any production path; only the tests still exercise it.
  ~150 LOC including tests.
- **`headers.go` and `headers_test.go` are stub files** with no
  content beyond the package declaration. Holdover from a planned
  exported header-handling surface that never materialized. Now
  redundant with `content/headers.go::ParseHeaders`.
- **The package doc comment is wrong.** "aerc email content filters"
  describes neither what this code does nor where it lives.

### Verdict — keep (with delete sub-items)

The CleanHTML/CleanPlain/ToHTML surface is the right shape for its
intended consumers. No structural refactor needed. But the package
carries dead weight that should be cleaned up before Pass 3 wires the
real consumers, while it's still cheap to do so.

**Recommendation (small docs+removal commit, no consumer dependency):**

1. Delete `reflowMarkdown`, `reflowParagraph`, `reflowBlockquote`,
   `markdownTokens`, `isParagraph`, `isBlockquote`, `reOrderedList`
   (in its reflow-only role — keep if `isShortPlain`/`isCompactLine`
   still need it; verify by ripgrep), and the corresponding
   `html_test.go` cases (~150 LOC across source + tests).
2. Delete `headers.go` and `headers_test.go`.
3. Rewrite the package doc comment in `convert.go` (or wherever it
   ends up after step 2) to describe poplar's own pipeline:
   *"Package filter cleans inbound email bodies (HTML or plain text)
   into normalized markdown for the content renderer, and converts
   outbound markdown to HTML for `multipart/alternative` send."*
4. **Note for Pass 3 plan:** `loadBodyCmd` will need a MIME-aware
   shim. Easiest path is to extend `mail.Backend.FetchBody` to
   return MIME alongside the body. Track separately — not part of
   this audit.

## `internal/content/`

### Current shape

A self-contained block-model rendering pipeline:

- **`blocks.go`** — `Block` and `Span` interfaces with a sealed-
  sum-type discipline. Concrete blocks: `Paragraph`, `Heading`,
  `Blockquote`, `QuoteAttribution`, `Signature`, `Rule`, `CodeBlock`,
  `Table`, `ListItem`. Concrete spans: `Text`, `Bold`, `Italic`,
  `Code`, `Link`. `Address` and `ParsedHeaders` value types live
  here too.
- **`parse.go`** — `ParseBlocks(string) []Block` parses
  already-normalized markdown into the block model, including
  email-aware behavior (signature `-- ` cutoff, `On … wrote:`
  attribution, recursive blockquote nesting, implied-quote wrapping
  for HTML emails missing `<blockquote>` tags). `parseSpans` is the
  inline-formatting parser.
- **`headers.go`** — `ParseHeaders(string) ParsedHeaders` parses
  raw RFC 2822 headers (folded continuation lines, CRLF normalization,
  bare-email fallback).
- **`render.go`** — `RenderBody(blocks, theme, width) string` renders
  the block model via lipgloss + `ansi.Wordwrap`, capped at
  `maxBodyWidth = 72`. `RenderHeaders` renders parsed headers with
  full-width address wrapping.
- **`render_footnote.go`** — `RenderBodyWithFootnotes(blocks, theme,
  width) (string, []string)` walks the block tree, rewrites each
  non-auto-link `Link` span to glue ` [^N]` to its last word with a
  no-break space, and appends an `[N]: <url>` footnote list below
  a horizontal rule. Auto-linked bare URLs (`Text == URL`) skip
  the marker. Duplicate URLs share a footnote number.

### Intended consumer

The viewer is the consumer, **and it's already wired up** in Pass
2.5b-4:

- `cmds.go::loadBodyCmd` calls `content.ParseBlocks` after fetching.
- `viewer.go::layout` calls `content.RenderHeaders` and
  `content.RenderBodyWithFootnotes`.
- `viewer.go::Viewer` holds `[]content.Block` plus `[]string` URLs
  for `1`–`9` link launching.

The footnote and width-cap behavior is codified in ADRs 0066 (body
width 72) and 0067 (footnote bracketed marker).

### Fit assessment

- **The exported API matches the consumer 1:1.** `ParseBlocks`,
  `RenderHeaders`, `RenderBodyWithFootnotes` are exactly what the
  viewer calls. No gaps, no friction.
- **`RenderBody` (no footnotes) is consumer-less in production.**
  Only tests call it directly; production flows through
  `RenderBodyWithFootnotes`. Could be unexported (`renderBody`)
  with the tests reaching through `RenderBodyWithFootnotes` instead,
  but the cost is low and the function is a natural building block
  worth keeping accessible. **Leave exported.**
- **`ParseHeaders` has no production consumer today** but is
  forward-looking. The viewer constructs `ParsedHeaders` literals
  from `MessageInfo`, bypassing parsing because the worker layer
  already has structured fields. Pass 3 may or may not need the
  raw-bytes parser depending on whether the JMAP/IMAP fork's headers
  arrive pre-parsed. **Keep exported** — speculative use is fine
  here because the cost is one ~80 LOC file.
- **The `blockKind` / `spanKind` enums are dead machinery.** Both
  interfaces (`Block`, `Span`) require an unexported method
  (`blockType() blockKind`, `spanType() spanKind`) returning a
  private kind constant. The sealed-sum-type pattern. But the
  consumer never calls those methods — discrimination always
  happens via Go type switches (`switch b := block.(type)`). The
  enum return values are never inspected. The interface marker
  methods could be the empty no-arg form (`isBlock()`,`isSpan()`)
  and the `kindParagraph...`/`kindText...` constants could be
  deleted entirely. ~30 LOC of compile-but-unused code.
- **`maxBodyWidth = 72` is unexported as a package constant.**
  Correct — it's an invariant codified in ADR 0066, not a knob.
- **Headers are uncapped, body is capped.** Header width comes
  from the panel width as the consumer passes it; body caps at 72
  inside `RenderBody`. Matches the invariant
  ("Body content rendering caps at maxBodyWidth = 72 cells.
  Headers wrap at the panel content width (uncapped).").

### Verdict — keep

The package is the right shape for its consumer. The Pass 2.5b-4
integration validated it under live use. No structural change.

**Recommendation:**

- **Optional ergonomic cleanup** (not blocking, low priority): drop
  the `blockKind`/`spanKind` enum constants and the `blockType()`/
  `spanType()` method bodies' return values; reduce the marker
  methods to no-args. Ship with the next pass that touches the
  package — no need for a dedicated commit.

## `internal/tidy/`

### Current shape

A complete, self-contained tidytext implementation. ~835 LOC including
tests. Layers:

- **`api.go`** — `CallAPI(url, key, model, system, text) (string,
  error)`: bare HTTP client for Anthropic's messages API. Uses a
  package-level `httpClient` with a 30-second timeout and a constant
  `defaultAPIURL`.
- **`prompt.go`** — `BuildPrompt(Config) string`: assembles a
  Claude system prompt from enabled rules + style preferences. Hard-
  coded "Do NOT" guardrails (no rephrasing, no tone changes, no
  contraction expansion, no code-fence touching) plus user-supplied
  custom instructions.
- **`quote.go`** — `SplitQuoted(string) (author string, []QuotedBlock)`
  separates author text from `>`-prefixed reply quotes. `Reassemble(
  corrected, original)` walks the original line-by-line, preserving
  quoted lines verbatim and substituting corrected non-quoted lines
  in order. The whole point of the package: don't proofread the
  person you're replying to.
- **`tidy.go`** — `Tidy(input, Config, key, url) (Result, error)`:
  the orchestrator. Status codes
  (`StatusCorrected`/`StatusNoChanges`/`StatusNoAuthorText`/
  `StatusError`) and a never-non-nil error return ("reserved for
  future extensibility").
- **`config.go`** — `Config`/`APIConfig`/`RulesConfig`/`StyleConfig`
  TOML structs. `DefaultConfig()` ships sensible defaults.
  `LoadConfig(path)` reads a TOML file and merges with defaults.
  `ApplyRuleOverrides([]string)` and `ApplyStyleOverrides([]string)`
  parse `"key=value"` strings into the config. `ResolveAPIKey(cfg)`
  prefers config over `$ANTHROPIC_API_KEY`. `ConfigString(cfg)`
  pretty-prints for human display. Validators for the enum-typed
  fields (`oxford_comma`, `ellipsis`, `time_format`).

### Intended consumer

Pass 9.5 (tidytext in compose). Compose presents the user's body in
the editor; on a key (likely `T` in compose mode, TBD), the body is
fed through tidy and the result is shown for accept/reject. Quoted
reply text is preserved unchanged regardless of what the LLM does
with the surrounding prose.

No live consumer today; not even imported anywhere.

### Fit assessment

The package was originally a standalone `tidytext` CLI binary —
visible in the `~/.local/bin` listing in the workstation context.
The current code shape was inherited from that lineage and never
re-shaped for a library consumer. That history shows:

**Core algorithm — fits compose well:**

- `Tidy(input, cfg, key, url)` is exactly what compose wants: a single
  entry point, a Result with status codes that drive UI feedback (toast
  / dim badge / error indicator), no errors to handle separately.
- `SplitQuoted`/`Reassemble` are the load-bearing innovation. The
  consumer doesn't call them directly — `Tidy` orchestrates — but
  they're the reason the package earns its keep.
- `BuildPrompt(cfg)` is internal to `Tidy`'s contract. Exporting it
  costs nothing and makes the system prompt inspectable.
- `CallAPI` is the only outbound network call. Exported so tests can
  stub it cleanly via a captured `httpClient`. Reasonable.

**CLI machinery — wrong shape for compose:**

- `LoadConfig(path)` assumes tidy has its own TOML file. Poplar's
  invariant says config lives at `~/.config/poplar/accounts.toml`
  with `[[account]]` blocks and a `[ui]` table both decoded by
  `internal/config`. A separate tidy TOML doesn't fit. Pass 9.5 will
  either add `[tidy]` to the unified config or ship `DefaultConfig()`
  unmodified. Either way `LoadConfig` is dead.
- `ApplyRuleOverrides` and `ApplyStyleOverrides` parse
  `"oxford_comma=insert"`-style strings. That's CLI flag plumbing.
  Compose users will toggle in a settings UI, not type
  `key=value` lines.
- `ConfigString(cfg)` pretty-prints for `tidy --print-config`.
  Compose has no equivalent surface.
- `ResolveAPIKey(cfg)` falls back to `$ANTHROPIC_API_KEY`. In poplar
  the API key will live in `~/.local/secrets` or the unified config;
  the fallback rule encodes a CLI assumption.

**Estimated dead weight when collapsed:** roughly half of `config.go`
plus the `LoadConfig`/`Apply*Overrides`/`ConfigString` test cases —
on the order of 100-150 LOC across source + tests. The validators
(`validateOxfordComma` etc.) stay because compose's settings layer
will still need to enforce the enum constraints.

### Verdict — collapse

The core algorithm is the right shape for Pass 9.5. The CLI
ergonomics around it were built for a different consumer (a
standalone binary) and need to come out before they leak into
compose's wiring.

**Recommendation (defer until Pass 9.5 plan):**

Don't touch the package now — it's not in anyone's path and
collapsing it ahead of the consumer risks rework if the compose
design surfaces a need we don't currently see. Instead, mark the
Pass 9.5 plan to:

1. Wire `Tidy()` from compose with a `Config` value built from the
   unified poplar config (or `DefaultConfig()` for v1).
2. Delete `LoadConfig`, `ApplyRuleOverrides`, `ApplyStyleOverrides`,
   `ConfigString`, `ResolveAPIKey`, and their tests. Move any
   surviving validators next to the unified config decode site.
3. Optionally unexport `CallAPI` and `BuildPrompt` if no test
   reaches them directly after the trim.

Goal of the collapse: make tidy a small library that exposes
`Config`, `DefaultConfig()`, `Tidy()`, `Result`, and the status
constants — and nothing else.

## Cross-cutting note

None of these three packages have a current production consumer
problem; the Pass 2.5b-4 viewer relies on `content/` and that
integration is healthy. The audit's premise — "this is the last
cheap moment to course-correct shape before consumer code starts
depending on the current API" — holds for `filter/` (Pass 3 is the
next pass that will demand it) and for `tidy/` (Pass 9.5, several
passes out). For `content/`, the moment has already passed without
incident — the API was right.

## Follow-up

| Verdict | Pass to attach to | Action |
|---|---|---|
| `filter/` keep + delete sub-items | Pass 2.5b-4.5 (already queued) or its own small commit | Drop reflow family + headers stubs + fix package doc. |
| `content/` keep | none | Optional: drop dead kind enums next time the package is touched. |
| `tidy/` collapse | Pass 9.5 plan note | Delete CLI machinery as part of consumer wiring. Don't touch the package before Pass 9.5 lands. |

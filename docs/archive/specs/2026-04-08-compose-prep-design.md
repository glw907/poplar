# compose-prep Design Spec

## Overview

`compose-prep` is a new Go binary in the beautiful-aerc project that
normalizes aerc compose buffers before the user starts typing. It
replaces ~250 lines of Lua in nvim-mail's `init.lua` with a tested,
correct Go implementation.

**Problem:** The current Lua pipeline does RFC 2822 header processing
with regex patterns that have known bugs: comma-in-display-names
breaks address splitting, byte-length is used instead of display-width
for international names, and the code is untestable. For a project
intended for public use, this is unacceptable.

**Solution:** A stdin/stdout Go binary that runs five normalization
steps on the compose buffer. The Lua shrinks to a three-line
`vim.fn.systemlist` call. The Go code uses `net/mail` for correct
RFC 2822 parsing and `go-runewidth` for display-width-aware line
wrapping.

## Interface

```
compose-prep [--no-cc-bcc] [--debug]
```

- Reads raw aerc compose buffer from **stdin**
- Writes normalized buffer to **stdout**
- No subcommands — single purpose

### Flags

| Flag | Default | Purpose |
|------|---------|---------|
| `--no-cc-bcc` | off (inject by default) | Suppress empty Cc/Bcc header injection |
| `--debug` | off | Write diagnostic messages to stderr |

### Error Behavior

On any processing error, the original input is written to stdout
unchanged. The compose window must always open. Exit code is always 0.

If `--debug` is set, errors are also logged to stderr via Go's
standard `log` package.

### Lua Caller

```lua
local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
local result = vim.fn.systemlist("compose-prep", lines)
if vim.v.shell_error == 0 and #result > 0 then
  vim.api.nvim_buf_set_lines(0, 0, -1, false, result)
end
```

If `compose-prep` is not installed, `systemlist` returns an empty
table and `shell_error` is non-zero. The guard leaves the buffer
unchanged — usable but not pretty. No crash, no error popup.

## Pipeline

Five steps, always in this order, operating on a line array. The
first blank line separates headers from body. Steps 1–4 operate on
headers. Step 5 operates on body.

### Step 1: Unfold RFC 2822 Continuation Lines

Lines starting with whitespace (space or tab) are joined to the
previous line with a single space replacing the leading whitespace.
Standard RFC 2822 unfolding. Only applies to the header section.

**Input:**
```
To: Alice <alice@example.com>,
    Bob <bob@example.com>
```

**Output:**
```
To: Alice <alice@example.com>, Bob <bob@example.com>
```

### Step 2: Strip Bare Angle Brackets

On address headers (From, To, Cc, Bcc), bare `<email@dom>` addresses
— those preceded by `:` or `,` rather than a display name — have
their angle brackets removed. Named addresses like
`Jane <jane@dom>` are untouched.

Uses `net/mail.ParseAddressList` for proper parsing. This correctly
handles quoted display names with commas (e.g.,
`"Smith, John" <j@dom.com>`), which the current Lua regex cannot.
If `ParseAddressList` fails on a header value (malformed addresses),
that header line passes through unchanged.

**Input:**
```
From: <alice@example.com>
To: Bob <bob@example.com>, <charlie@example.com>
```

**Output:**
```
From: alice@example.com
To: Bob <bob@example.com>, charlie@example.com
```

### Step 3: Re-fold Address Headers

To, Cc, and Bcc lines with multiple recipients are wrapped at 72
columns, breaking at recipient boundaries (commas). Continuation
lines are indented to align under the first address: key length + 2
spaces (e.g., `To: ` = 4 spaces, `Bcc: ` = 5 spaces).

Single-recipient headers pass through unchanged. Column counting
uses `go-runewidth` for correct display-width measurement of
international names and CJK characters.

**Input:**
```
To: Alice <alice@example.com>, Bob <bob@example.com>, Charlie <charlie@example.com>
```

**Output (if > 72 cols):**
```
To: Alice <alice@example.com>, Bob <bob@example.com>,
    Charlie <charlie@example.com>
```

### Step 4: Inject Empty Cc/Bcc Headers

If no `Cc:` header exists, insert `Cc:` after the last To: line
(including any continuation lines). Same for `Bcc:`. Resulting order:
To, Cc, Bcc.

Skipped entirely when `--no-cc-bcc` is set.

If both headers are already present, this step is a no-op.

### Step 5: Reflow Quoted Text

In the body (after the header-separating blank line), consecutive
quoted lines at the same `>` depth are joined into paragraphs, then
re-wrapped at 72 columns with normalized `> ` prefixes (space after
each `>`).

- Blank quoted lines are preserved as paragraph breaks
- Decorative lines (no alphanumeric content) are preserved as-is
- Different quote levels are reflowed independently
- Unquoted body lines pass through unchanged
- Column counting uses `go-runewidth`

**Input:**
```
> This is a long quoted line that was wrapped at an odd point by the
> original sender's email client and looks
> ragged.
```

**Output:**
```
> This is a long quoted line that was wrapped at an odd point by the
> original sender's email client and looks ragged.
```

## Package Structure

```
cmd/compose-prep/
  main.go            boilerplate main()
  root.go            cobra root command, flags struct, RunE

internal/compose/
  prepare.go         Prepare(input []byte, opts Options) []byte
  unfold.go          unfoldHeaders(headers []string) []string
  bracket.go         stripBrackets(headers []string) []string
  fold.go            foldAddresses(headers []string) []string
  inject.go          injectCcBcc(headers []string) []string
  reflow.go          reflowQuoted(body []string) []string
  prepare_test.go    integration tests (full buffer in/out)
  unfold_test.go     unit tests
  bracket_test.go    unit tests
  fold_test.go       unit tests
  inject_test.go     unit tests
  reflow_test.go     unit tests
```

### Options Type

```go
type Options struct {
    InjectCcBcc bool
    Debug       bool
}
```

### Public API

`Prepare` is the only exported function. The five step functions are
unexported — tested directly in same-package unit tests, but external
consumers only call `Prepare`.

The `cmd/` layer reads stdin, calls `compose.Prepare`, writes stdout.
No business logic in `cmd/`.

### Dependency

One new dependency: `github.com/mattn/go-runewidth`. Used in
`fold.go` and `reflow.go` for display-width-correct column counting.

### Makefile

`compose-prep` added to `build`, `install`, `clean`, and `.PHONY`
targets alongside the four existing binaries.

## Testing Strategy

### Unit Tests (per step)

Table-driven, same package, testing each pipeline step in isolation:

**unfold_test.go:**
- Single-line headers pass through
- Continuation lines (space-prefixed) join to previous
- Tab-prefixed continuation lines join
- Multiple consecutive continuations
- Body lines untouched

**bracket_test.go:**
- Bare `<email>` after colon stripped
- Bare `<email>` after comma stripped
- `Name <email>` preserved
- Non-address headers (Subject, Date) untouched
- Quoted display names with commas (`"Smith, John" <j@dom>`)

**fold_test.go:**
- Short recipient list stays on one line
- Long list wraps at 72 columns
- Continuation indent = key length + 2
- Single recipient is a no-op
- CJK display names counted by display width

**inject_test.go:**
- Missing Cc and Bcc both added
- Existing Cc/Bcc preserved
- Insertion order: To, Cc, Bcc
- To with continuation lines — injection after last continuation
- `InjectCcBcc=false` skips entirely

**reflow_test.go:**
- Single-level quotes reflowed
- Nested quotes reflowed independently
- Blank quoted lines preserved as paragraph breaks
- Decorative lines (no alphanumeric) preserved
- Mixed quote levels not merged
- Unquoted body lines untouched

### Integration Tests (prepare_test.go)

Full compose buffers in, normalized buffers out:

- New compose (empty To, no body)
- Reply (populated To, quoted text)
- Forward (empty To, quoted text)
- Multi-recipient with long addresses
- International display names (accented, CJK)
- Malformed input (no blank line separator) — pass-through

### No E2E Tests

The e2e framework in `e2e/` is built around piping HTML fixtures
through mailrender. compose-prep's integration tests with full buffer
strings serve the same purpose without the binary build overhead.

## Lua Integration

### What Gets Deleted from init.lua

- `get_quote_prefix()` function (~3 lines)
- `normalize_prefix()` function (~4 lines)
- `wrap_text()` function (~22 lines)
- `reflow_quoted()` function (~57 lines)
- Header unfold loop (~7 lines)
- Bracket stripping gsub (~8 lines)
- Address re-folding loop (~32 lines)
- Cc/Bcc injection block (~18 lines)
- Quote reflow call (~2 lines)

Total: ~150 lines removed.

### What Stays in Lua

- `vim.fn.systemlist("compose-prep", lines)` call (3 lines)
- Blank line insertion and extmark separator overlays
- Cursor positioning logic (To: line vs body)
- BufWritePre blank-line stripping
- Filetype setting, spell check config, all keybindings
- Tidytext, khard picker, signature — all independent features

### Dual Config Update

Both copies get the same Lua change:
- `.config/nvim-mail/init.lua` (project repo)
- `~/.dotfiles/beautiful-aerc/.config/nvim-mail/init.lua` (personal)

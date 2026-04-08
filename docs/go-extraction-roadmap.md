# Go Extraction Roadmap

Shell scripts and Lua code that should be ported to Go binaries for
reliability, testability, and a simpler setup experience.

## Priority 1: compose-prep (new binary)

**Status:** In design — see `docs/superpowers/specs/` for current spec.

**What it replaces:** ~250 lines of Lua in `init.lua` VimEnter autocmd
that process the compose buffer before the user types anything:

- RFC 2822 header unfolding (continuation lines)
- Bare angle bracket stripping (`<email>` → `email`)
- Address header re-folding (To/Cc/Bcc wrapping at 72 cols)
- Empty Cc:/Bcc: injection when absent
- Quoted text reflow (join paragraphs, re-wrap at 72 cols)

**Why:** Regex-based RFC 2822 parsing in Lua has known bugs
(comma-in-display-names breaks address splitting, byte-length vs
display-width miscounts international names). Go's `net/mail` handles
these correctly. Moving to Go also makes the pipeline testable with
table-driven tests and e2e fixtures.

**Result:** The VimEnter autocmd shrinks to a single
`vim.fn.systemlist("compose-prep", lines)` call. The Lua retains only
editor-specific logic (extmark separators, cursor positioning, insert
mode entry).

## Priority 2: mailrender themes generate (new subcommand)

**What it replaces:** `themes/generate` — a 285-line POSIX shell script.

**Why it's fragile:**
- Uses `eval` for variable expansion (silent wrong output on typos)
- Hand-rolled hex-to-ANSI conversion with `cut` and shell arithmetic
- Override preservation via `sed` with embedded regex
- Must be discovered and run manually — no `make` integration

**What already exists in Go:**
- `internal/palette.HexToANSI()` — exact same conversion
- `internal/palette.parseAssignment()` — reads the format it produces

**Result:** `mailrender themes generate themes/nord.sh` — one binary,
one command, testable. Could also add `--kitty` flag to auto-sync
`kitty-mail.conf` color values (currently manual and prone to drift).

## Priority 3: mailrender save (new subcommand)

**What it replaces:** `.local/bin/aerc-save-email` — 62-line bash script.

**Why:** Path-traversal fragility (`cd "$AERC_CONFIG/../.."` can resolve
wrong), HTML detection via `grep`, collision-avoidance via shell loop,
corpus counting via `find | wc -l`. All untestable. `mailrender` already
understands `AERC_ROWS`/`AERC_COLUMNS` env vars and stdin piping.

**Result:** `:pipe -m mailrender save` — same user workflow, but the
implementation is testable and the path logic is robust.

## Priority 4: mailrender audit (new subcommand)

**What it replaces:** `scripts/audit.sh` — 200-line bash script.

**Why:** Currently broken (references wrong binary name). Uses 10+
subprocess chains for platform detection. Developer-facing maintenance
tool, so lower urgency, but would benefit from structured output and
proper Go testing.

**Result:** `mailrender audit ~/.cache/aerc/907.life/blobs/` — walks
blobs, clusters by sending platform, renders samples, produces a report.

## Not Moving

- **`filters/unwrap-tables.lua`** — pandoc Lua filter for AST
  manipulation. Correct tool for the job; no Go equivalent without
  replacing pandoc entirely.
- **`.local/bin/mail`** — one-line kitty launcher. Nothing to simplify.
- **`.local/bin/nvim-mail`** — one-line NVIM_APPNAME wrapper.

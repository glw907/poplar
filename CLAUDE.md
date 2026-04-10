# beautiful-aerc

Themeable aerc email filters and configuration, distributed as a
single GNU Stow package. Builds four binaries from one Go module:
mailrender, fastmail-cli, tidytext, poplar.

## Poplar

@docs/poplar/STATUS.md
@docs/poplar/architecture.md

## MANDATORY: Go Conventions

**Read and follow `~/.claude/docs/go-conventions.md` before writing
ANY Go code.** Key rules:

- No unnecessary interfaces, goroutines, builder patterns
- `cmd/` for CLI wiring only, `internal/` for business logic
- cobra with `SilenceUsage: true`, flags in a struct
- `fmt.Errorf("context: %w", err)` at every error boundary
- Table-driven tests, no assertion libraries
- `make check` (vet + test) must pass before any commit

## MANDATORY: Go Skill

**Use superpowers:go skill for all Go development tasks.**

## Project Structure

```
cmd/mailrender/        CLI: filters, themes, compose (cobra)
cmd/fastmail-cli/      CLI: rules, masked, folders (cobra)
cmd/tidytext/          CLI: fix, config (cobra)
internal/filter/       Filter implementations (headers, html, plain)
internal/compose/      Compose buffer normalization (mailrender compose)
internal/theme/        TOML theme files -> ANSI tokens
internal/tidy/         Prose tidying: config, prompt, API
internal/jmap/         JMAP session, mail ops, masked email
internal/header/       RFC 2822 header parsing
internal/rules/        Local JSON rule file operations
e2e/                   E2E tests (build binary, pipe fixtures, golden files)
.config/aerc/          aerc configuration files + themes
.config/nvim-mail/     Neovim compose editor profile
```

## aerc Filter Protocol

aerc calls filters as shell commands. Each filter:
- Receives email content on **stdin**
- Writes ANSI-styled text to **stdout**
- Has access to `AERC_COLUMNS` env var (terminal width)

## Theme System

Theme files (`.config/aerc/themes/*.toml`) define 16 semantic hex
color slots + token definitions. Go binaries read `.toml` files
directly at runtime. Active theme from `styleset-name` in `aerc.conf`.

**Never hardcode ANSI color codes in Go source.** All styling must
use tokens from the theme file via the theme package.

Generate aerc styleset: `mailrender themes generate [name]`.
See `docs/styling.md` for visual hierarchy and `docs/themes.md`
for the token reference.

## Build

```
make build     # build all three binaries
make test      # run tests
make check     # vet + test (gate before commits)
make install   # install all three to ~/.local/bin/
```

## Testing

- **Unit tests:** table-driven, same package, alongside source
- **E2E tests:** build binary in TestMain, pipe fixtures, golden files
- **Live verification:** see `.claude/docs/tmux-testing.md`

**MANDATORY: When the user reports a rendering problem, always
verify the fix against the live email.** Fetch the raw HTML via
the Fastmail JMAP API (see memory for access details), pipe it
through the rebuilt binary, and confirm the issue is resolved.
Do not rely solely on unit tests or synthetic fixtures.

## Corpus

`corpus/` holds raw email parts flagged for rendering issues.
Save from aerc using `aerc-save-email`. The `/fix-corpus` skill
batch-processes accumulated corpus emails.

The Go binaries are installed via `make install` (not stowed).

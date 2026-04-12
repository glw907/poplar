# beautiful-aerc

Themeable aerc email filters and configuration, distributed as a
single GNU Stow package. Builds four binaries from one Go module:
mailrender, fastmail-cli, tidytext, poplar.

## Poplar

@docs/poplar/STATUS.md
@docs/poplar/architecture.md
@docs/poplar/styling.md

## MANDATORY: Elm Architecture (Poplar UI)

**Read and follow `~/.claude/docs/elm-conventions.md` before writing
ANY code in `internal/ui/`.** Key rules:

- All state in tea.Model structs, no package-level mutable vars
- State changes only in Update, never in View/Init/Cmd closures
- All I/O in tea.Cmd, never blocking in Update or View
- Children signal parents via Msg types, never method calls
- Shared state hoisted to root, passed down read-only

## MANDATORY: Go Conventions

**Read and follow `~/.claude/docs/go-conventions.md` before writing
ANY Go code.** Key rules:

- No unnecessary interfaces, goroutines, builder patterns
- `cmd/` for CLI wiring only, `internal/` for business logic
- cobra with `SilenceUsage: true`, flags in a struct
- `fmt.Errorf("context: %w", err)` at every error boundary
- Table-driven tests, no assertion libraries
- `make check` (vet + test) must pass before any commit

## Project Structure

```
cmd/mailrender/        CLI: filters, themes, compose (cobra)
cmd/fastmail-cli/      CLI: rules, masked, folders (cobra)
cmd/tidytext/          CLI: fix, config (cobra)
cmd/poplar/            Poplar bubbletea email client entry point
internal/filter/       Content pipeline: CleanHTML, CleanPlain (raw email -> markdown)
internal/content/      Block model + lipgloss renderer (ParseBlocks, RenderBody, RenderHeaders)
internal/compose/      Compose buffer normalization (mailrender compose)
internal/theme/        Compiled lipgloss themes (Palette -> CompiledTheme)
internal/tidy/         Prose tidying: config, prompt, API
internal/jmap/         JMAP session, mail ops, masked email
internal/header/       RFC 2822 header parsing
internal/rules/        Local JSON rule file operations
internal/ui/           Poplar bubbletea components (app, sidebar, msglist, styles, footer)
internal/mail/         Poplar Backend interface, poplar-native types, mock backend
internal/poplar/       Poplar AccountConfig and poplar-specific types
internal/aercfork/     Forked aerc workers (IMAP + JMAP) and supporting packages
e2e/                   E2E tests (build binary, pipe fixtures, golden files)
.config/aerc/          aerc configuration files + themes
.config/nvim-mail/     Neovim compose editor profile
```

## aerc Filter Protocol

aerc calls filters as shell commands. Each filter:
- Receives email content on **stdin**
- Writes ANSI-styled text to **stdout**
- Has access to `AERC_COLUMNS` env var (terminal width)

## Charmbracelet Libraries (Bubbletea, Lipgloss)

**Read the library docs before writing custom code.** Lipgloss
handles all styling. Do not build custom ANSI manipulation when
a library feature already exists.

## Theme System

Themes are compiled Go values in `internal/theme/`. Each theme
is a `Palette` (16 hex colors) → `NewCompiledTheme()` →
`*CompiledTheme` with lipgloss.Style fields. Ships 15 themes
(10 dark, 5 light) with One Dark as the default.

**Never hardcode ANSI color codes in Go source.** All styling
must use lipgloss styles from `CompiledTheme`.

Generate aerc styleset: `mailrender themes generate [name]`.
See `docs/themes.md` for the theme reference and
`docs/styling.md` for mailrender visual hierarchy.
`docs/poplar/styling.md` (imported at the top of this file) is
the separate, authoritative map for poplar UI surfaces — don't
confuse the two.

## Build

```
make build     # build all four binaries
make test      # run tests
make check     # vet + test (gate before commits)
make install   # install all four to ~/.local/bin/
```

## Testing

- **Unit tests:** table-driven, same package, alongside source
- **E2E tests:** build binary in TestMain, pipe fixtures, golden files
- **Live verification:** see `.claude/docs/tmux-testing.md`

**MANDATORY: Install and verify live before finishing any
rendering work.** The loop is: `make install` → render the real
message in aerc via tmux-testing (`.claude/docs/tmux-testing.md`)
→ confirm the fix. When the user reports a rendering problem,
fetch the raw HTML via the Fastmail JMAP API (see memory) and
pipe it through the rebuilt binary — do not rely on unit tests
or synthetic fixtures alone.

**Config has two copies.** The project repo (`.config/aerc/`)
holds the distributable starter config; `~/.dotfiles/beautiful-aerc/`
holds the user's local config deployed via `stow -R beautiful-aerc`.
The local copy diverges in personal settings (signature, account,
mailbox names/order) and optional tool keybindings (tidytext,
fastmail-cli). Update whichever copy the change targets — update
both when it applies to both.

## Corpus

`corpus/` holds raw email parts flagged for rendering issues.
Save from aerc using `aerc-save-email`. The `/fix-corpus` skill
batch-processes accumulated corpus emails.

The Go binaries are installed via `make install` (not stowed).

## Backlog

`BACKLOG.md` is the project issue tracker. Log issues there using
`/log-issue`. Check it before starting work — it may contain
known limitations or upstream blockers relevant to the task.

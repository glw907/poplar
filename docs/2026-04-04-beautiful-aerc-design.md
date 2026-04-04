# beautiful-aerc Design Spec

A themeable, distributable aerc email setup with Nord-based defaults,
a Go filter binary, and optional nvim-mail and kitty integration.

## Goals

- Single `stow beautiful-aerc` installs the full mail stack
- Easy theme switching: change one file, regenerate, done
- Ships with 3 themes (Nord, Solarized Dark, Gruvbox Dark)
- Users can `git pull` for upstream updates while keeping local overrides
- Minimal runtime dependencies: aerc, pandoc, Go binary

## Repository Structure

```
beautiful-aerc/
  README.md
  go.mod                           # go 1.23 (matches aerc)
  go.sum
  Makefile                         # build + install Go binary
  cmd/
    beautiful-aerc/
      main.go
  internal/
    palette/
      palette.go
    filter/
      headers.go
      html.go
      plain.go
  .config/aerc/
    aerc.conf
    binds.conf
    accounts.conf.example
    themes/
      nord.sh
      solarized-dark.sh
      gruvbox-dark.sh
      generate
    generated/
      palette.sh
    stylesets/
      nord-custom
    filters/
      unwrap-tables.lua
  .config/nvim-mail/
    init.lua
    syntax/aercmail.vim
  .config/kitty/
    kitty-mail.conf
  .local/bin/
    mail
    nvim-mail
  .local/share/applications/
    aerc-mail.desktop
```

## Go Binary

Single binary with three cobra subcommands. Each reads stdin,
writes styled ANSI text to stdout - aerc's filter protocol.

### aerc.conf integration

```ini
.headers=beautiful-aerc headers
text/html=beautiful-aerc html
text/plain=beautiful-aerc plain
```

### Project layout

```
cmd/beautiful-aerc/main.go       # cobra root, SilenceUsage: true
internal/palette/palette.go       # parse palette.sh, expose tokens
internal/filter/headers.go        # reorder, colorize, wrap, separator
internal/filter/html.go           # cleanup, exec pandoc, highlight, links
internal/filter/plain.go          # detect HTML-in-plain, route or wrap
```

### Subcommands

| Subcommand | Replaces | Calls pandoc? |
|---|---|---|
| `headers` | format-headers (sh) + format-headers.awk | No |
| `html` | html-to-text (sh + perl + sed) | Yes, as subprocess |
| `plain` | wrap-plain (sh) | Delegates to `html` when HTML detected, otherwise execs `wrap \| colorize` |

### What the Go binary absorbs

- All sed/perl regex cleanup (pandoc artifacts, zero-width chars, whitespace)
- Markdown syntax highlighting (headings, bold, italic, rules)
- Link styling (colored text, dimmed URL, strip colorize ANSI from URLs)
- Header formatting (reorder, colorize, address wrapping, separator)
- Palette loading and hex-to-ANSI conversion

### What stays external

- `pandoc` - HTML-to-markdown conversion, called as subprocess
- `unwrap-tables.lua` - pandoc Lua filter, passed via `-L` flag
- `colorize`, `wrap`, `linkify` - aerc built-ins, used by `plain` subcommand

### Palette loading

The binary finds `palette.sh` by checking in order:
1. `$AERC_CONFIG/generated/palette.sh`
2. Relative to binary: `../../.config/aerc/generated/palette.sh`
3. `~/.config/aerc/generated/palette.sh`
4. Built-in Nord defaults (always works, zero config)

Parsing is simple: `KEY=value` and `KEY="value"` lines. Ignores
comments and blank lines.

## Theme System

### Theme file format

Each theme file defines 16 semantic color slots (hex) plus markdown
tokens that reference those slots:

```sh
# Semantic color slots
BG_BASE="#2e3440"
BG_ELEVATED="#3b4252"
BG_SELECTION="#394353"
BG_BORDER="#49576b"
FG_BASE="#d8dee9"
FG_BRIGHT="#e5e9f0"
FG_BRIGHTEST="#eceff4"
FG_DIM="#616e88"
ACCENT_PRIMARY="#81a1c1"
ACCENT_SECONDARY="#88c0d0"
ACCENT_TERTIARY="#8fbcbb"
COLOR_ERROR="#bf616a"
COLOR_WARNING="#d08770"
COLOR_SUCCESS="#a3be8c"
COLOR_INFO="#ebcb8b"
COLOR_SPECIAL="#b48ead"

# Markdown tokens (reference slots + style modifiers)
C_HEADING="$COLOR_SUCCESS bold"
C_BOLD="bold"
C_ITALIC="italic"
C_LINK_TEXT="$ACCENT_SECONDARY"
C_LINK_URL="$FG_DIM"
C_RULE="$FG_DIM"
```

### Generator

`themes/generate` is a POSIX shell script. No nvim dependency.

- Reads one theme file as input
- Resolves slot references in markdown tokens
- Converts hex to ANSI escape parameters
- Produces two artifacts:
  - `generated/palette.sh` - hex colors + resolved ANSI tokens
  - `stylesets/<theme-name>` - aerc styleset with hex values

Usage:

```sh
themes/generate                        # uses nord.sh (default)
themes/generate themes/solarized-dark.sh   # switch theme
```

### Override mechanism

Both generated files include a marker line:
```
# --- overrides below this line are preserved across regeneration ---
```
User customizations below this line survive re-running the generator.

### Shipped themes

- `nord.sh` - daily driver, fully tested
- `solarized-dark.sh` - classic dark theme
- `gruvbox-dark.sh` - warm dark theme

## Config Files

### Tracked (users can fork or use as reference)

- `aerc.conf` - working config with comments. Personal hooks removed.
  References `beautiful-aerc` binary for filters, `nvim-mail` for editor,
  `nord-custom` styleset.
- `binds.conf` - fully generic keybindings, ships as-is.

### Template

- `accounts.conf.example` - placeholder values, user copies and edits.

### Gitignored

- `accounts.conf` - user's real account config
- Personal credential files

## nvim-mail

Ships as-is with two changes:
- Signature block in init.lua replaced with placeholder
- Hardcoded Nord colors in `syntax/aercmail.vim` documented as
  "edit these if you change themes" (vim syntax files don't support
  variable colors)

## kitty-mail

Ships as-is. Nord color block (color0-color15) is hardcoded.
Documented as "edit to match your theme." kitty and nvim have
established theme ecosystems -- beautiful-aerc covers the aerc layer.

## Launcher scripts

- `mail` - `exec kitty --class aerc-mail --config ... aerc`
- `nvim-mail` - `NVIM_APPNAME=nvim-mail exec nvim "$@"`

Both generic, ship as-is.

## What gets deleted from current setup

- `filters/format-headers` (shell wrapper) - absorbed into Go
- `filters/format-headers.awk` - absorbed into Go
- `filters/html-to-text` (shell/perl pipeline) - absorbed into Go
- `filters/wrap-plain` (shell) - absorbed into Go
- `filters/palette.sh` - moves to `generated/palette.sh`
- `filters/generate-palette` - becomes `themes/generate`
- `mailrules.json` - personal, not shipped
- `fastmail-*.age` - personal, not shipped
- `fastmail-password`, `fastmail-dav-password` - personal, not shipped

## Runtime dependencies

- `aerc` (email client)
- `pandoc` (HTML-to-markdown, called by Go binary)
- `beautiful-aerc` (Go binary, built from this repo)

## Optional dependencies

- `kitty` (terminal, for `mail` launcher)
- `nvim` (editor, for compose via nvim-mail)
- `khard` (address book completion)

## Build dependencies

- Go 1.23+ (matches aerc's go.mod)
- `make`

## Future ideas (not in scope)

- Install script that asks questions and personalizes accounts.conf,
  editor choice, etc.
- Blog post: "beautiful aerc" detailing the setup
- Additional themes contributed by users

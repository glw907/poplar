# Glamour Pipeline Migration

Replace the pandoc-based HTML rendering pipeline with Go-native
libraries (html-to-markdown + Glamour), eliminate the pick-link
binary, and add table rendering support. Consolidate mailrender as
the single binary for the full email rendering lifecycle.

## Goals

- Replace pandoc subprocess with Go-native HTML-to-markdown conversion
- Replace custom markdown-to-ANSI highlighting with Glamour rendering
- Add data table rendering (pipe tables with box-drawing borders)
- Use OSC 8 hyperlinks for all links (no visible footnote section)
- Eliminate pick-link binary (OSC 8 + Ctrl+click replaces it)
- Add `mailrender to-html` to replace pandoc in multipart-converters
- Add `mailrender markdown` for clean markdown output in reply templates
- ~750 lines deleted, ~150-200 lines added (~550-600 net reduction)

## New Dependencies

| Library | Purpose |
|---|---|
| `github.com/JohannesKaufmann/html-to-markdown/v2` | HTML → markdown conversion (replaces pandoc) |
| `github.com/charmbracelet/glamour` | Markdown → styled ANSI terminal output |
| `github.com/charmbracelet/glamour/ansi` | Custom style config for theme integration |

goldmark is pulled in transitively by Glamour and used directly
for the `to-html` subcommand.

## External Dependency Eliminated

pandoc is no longer required at runtime. It was used in two places:
- `mailrender html` (HTML → markdown) — replaced by html-to-markdown
- aerc `multipart-converters` (markdown → HTML) — replaced by `mailrender to-html`

## Pipeline Architecture

### Reading: `mailrender html`

```
HTML → prepareHTML → html-to-markdown (with plugins) → Glamour → stdout
```

1. **Pre-clean** (`prepareHTML`): Strip Mozilla attributes,
   `display:none` divs, zero-size tracking images. Borrowed from
   current code.

2. **HTML → Markdown** (`html-to-markdown` v2): Go-native conversion.
   Layout tables flattened via custom plugin rule (no `<th>` = layout
   table → unwrap to paragraphs). Data tables with `<th>` survive as
   pipe tables.

3. **Markdown → ANSI** (Glamour): Renders markdown to styled terminal
   output. Custom `ansi.StyleConfig` built from theme TOML tokens.
   OSC 8 hyperlinks on all links. Tables rendered with box-drawing
   borders. Terminal width from `AERC_COLUMNS`.

### Replying: `mailrender markdown`

```
HTML → prepareHTML → html-to-markdown (with plugins) → stdout
```

Same as reading but skips Glamour. Outputs clean markdown for use
in aerc reply templates so quoted text matches what the user read.

### Sending: `mailrender to-html`

```
stdin (markdown) → goldmark → stdout (HTML document)
```

Replaces `pandoc -f markdown -t html --standalone` in aerc's
`multipart-converters`. Clean semantic HTML output (`<h1>`,
`<strong>`, `<em>`, `<table>`, `<a href>`). No inline CSS.

## mailrender Subcommands (After Migration)

| Subcommand | Direction | Purpose |
|---|---|---|
| `html` | reading | HTML → Glamour-styled ANSI (viewer) |
| `headers` | reading | RFC 2822 → styled header block |
| `plain` | reading | plain text passthrough |
| `markdown` | replying (new) | HTML → clean markdown, no ANSI |
| `to-html` | sending (new) | markdown → HTML document |
| `themes generate` | config | theme TOML → aerc styleset |

## Theme Integration

Theme TOML tokens map to Glamour's `ansi.StyleConfig`:

| Theme token | Glamour style element |
|---|---|
| `heading` | `H1`–`H6` (color + bold) |
| `bold` | `Strong` (bold attribute) |
| `italic` | `Emph` (italic attribute) |
| `link_text` | `Link.Color` |
| `rule` | `HorizontalRule` |

New method on theme package: `t.GlamourStyle(cols int)` builds the
style config from loaded TOML tokens.

Document-level margins/indent set to 0 (aerc's viewer handles
padding).

## Layout Table Detection

Implemented as an html-to-markdown plugin rule:

- **Data table**: has `<th>` elements → convert to pipe table
- **Layout table**: no `<th>` → unwrap cells to sequential paragraphs
  (same behavior as current Lua filter)

Conservative default: flatten when uncertain. Marketing emails
abuse `<table>` heavily and must not render as data tables.

## OSC 8 Links

All links rendered as OSC 8 hyperlinks by Glamour (native support).
Link text is styled per theme `link_text` token. URLs are embedded
in escape sequences — clickable in supporting terminals, invisible
otherwise.

**Terminal support**: kitty, WezTerm, foot, Alacritty, Windows
Terminal, iTerm2, Ghostty, Konsole, recent GNOME Terminal.
Unsupported terminals silently ignore the sequences — link text
shows as plain text.

**tmux**: requires 3.4+ with `set -ga terminal-features "*:hyperlinks"`.

## Deletions

### Binaries
- `cmd/pick-link/` — entire directory
- `internal/picker/` — entire directory

### Filter code (`internal/filter/`)
- `footnotes.go` + tests (~364 lines)
- `highlightMarkdown`, `wrapLines`, `replaceMarkerPairs` (~100 lines)
- `markdownColors` struct, `boldPlaceholder`, sentinel logic
- `normalizeBoldMarkers` (~20 lines)
- `cleanPandocArtifacts` (~10 lines)
- `normalizeLists` (~60 lines)
- `runPandoc`, `writeLuaFilter`, `unwrapTablesLua` (~80 lines)
- `htmlToFootnotes`, `HTMLLinks`, `ExtractFootnoteLinks` exports
- Most of the 18 compiled regexes

### Theme tokens
- `picker_*` tokens (6 tokens)
- `msg_dim`, `link_url` (footnote-only)

### Kept from current code
- `prepareHTML`, `stripHiddenElements` (~55 lines)
- `normalizeWhitespace` (~10 lines)
- `stripANSI` utility
- Header filter (unchanged)
- Plain text filter (unchanged)

## Config Changes

### aerc.conf
```ini
[multipart-converters]
text/html = mailrender to-html
```

### binds.conf
- Remove `Tab` binding for pick-link in `[view]`

### Reply template
- Pipe body through `mailrender markdown` for clean quoted text

### Makefile
- Remove `pick-link` from `build`, `install`, `clean` targets
- Add `to-html` and `markdown` if they become separate binaries
  (they won't — they're subcommands of mailrender)

## Risk: Per-Line ANSI State

aerc's viewer may render each line independently, stripping
carry-over ANSI state. If Glamour emits bold-open on line 5 and
bold-close on line 7, lines 6-7 may lose styling.

**Mitigation**: test early with aerc. If needed, add a thin
post-processing pass (~20 lines) that tracks ANSI state and
re-emits open codes at the start of each line.

## Testing

- E2E golden files regenerated with `-update-golden` (output format
  changes significantly)
- E2E test structure unchanged (build binary, pipe fixtures, compare)
- Unit tests for new code: theme-to-Glamour mapping, layout table
  detection, `to-html` conversion
- Manual verification with email corpus and live aerc

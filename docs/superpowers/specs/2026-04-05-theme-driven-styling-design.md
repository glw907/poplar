# Theme-Driven Styling

Eliminate all hardcoded ANSI style modifiers (bold, italic, underline)
from Go source. Every text styling attribute flows through composite
palette tokens defined in the theme file.

## Problem

Four locations in Go code hardcode bold (`\033[1m` or `;1m`):

- `headers.go:61` — header key bold
- `save.go:69-70` — heading bold
- `save.go:86` — `#` marker bold
- `picker.go:204` — `#` marker bold

Additionally, `ColorsFromPalette` in `picker.go` and
`colorsFromPalette` in `headers.go` call `palette.HexToANSI()` on
base color tokens and manually assemble ANSI sequences. This means
users cannot control weight, italic, or underline for those elements
through the theme.

## Design Principle

Theme tokens follow a CSS-like model:

- **Base tokens** are variables — pure hex colors (`ACCENT_PRIMARY`,
  `FG_DIM`, etc.). They never carry modifiers.
- **Composite tokens** are classes — they reference base tokens and
  add style modifiers. Every styled UI element in Go uses a composite
  token, never a base token directly.

The generator already resolves composite tokens
(`"$COLOR_SUCCESS bold"` becomes `"1;38;2;163;190;140"`). The Go
code receives pre-resolved ANSI parameter strings and wraps them as
`\033[<value>m`. No modifier logic in Go.

## New Composite Tokens

Added to theme files alongside the existing markdown tokens:

```sh
# -- Header tokens --
C_HDR_KEY="$ACCENT_PRIMARY bold"
C_HDR_VALUE="$FG_BASE"
C_HDR_DIM="$FG_DIM"

# -- Picker tokens --
C_PICKER_NUM="$ACCENT_PRIMARY"
C_PICKER_LABEL="$FG_BASE"
C_PICKER_URL="$FG_DIM"
C_PICKER_SEL_BG="$BG_SELECTION"
C_PICKER_SEL_FG="$FG_BRIGHT"

# -- Message UI tokens --
C_MSG_MARKER="$FG_DIM bold"
C_MSG_TITLE_SUCCESS="$COLOR_SUCCESS bold"
C_MSG_TITLE_ACCENT="$ACCENT_PRIMARY bold"
C_MSG_DETAIL="$FG_BASE"
C_MSG_DIM="$FG_DIM"
```

Users can customize any of these with any combination of color +
`bold`, `italic`, `underline`:

```sh
C_HDR_KEY="$ACCENT_PRIMARY bold italic"
C_PICKER_LABEL="$FG_BASE underline"
```

## Token Reuse

Semantic tokens are reused across contexts rather than creating
one-off tokens for each element. New tokens are only added when an
element has genuinely distinct styling needs (e.g., header keys vs.
picker labels).

## Changes by File

### Theme files (`themes/nord.sh`, `gruvbox-dark.sh`, `solarized-dark.sh`)

Add the header, picker, and message UI composite token sections
shown above. Each theme uses its own base color palette.

### Generator (`themes/generate`)

Add the new tokens to the palette.sh emit block:

```sh
# -- Header tokens (ANSI) --
C_HDR_KEY="$(resolve_token "$C_HDR_KEY")"
C_HDR_VALUE="$(resolve_token "$C_HDR_VALUE")"
C_HDR_DIM="$(resolve_token "$C_HDR_DIM")"

# -- Picker tokens (ANSI) --
C_PICKER_NUM="$(resolve_token "$C_PICKER_NUM")"
C_PICKER_LABEL="$(resolve_token "$C_PICKER_LABEL")"
C_PICKER_URL="$(resolve_token "$C_PICKER_URL")"
C_PICKER_SEL_BG="$(resolve_token "$C_PICKER_SEL_BG")"
C_PICKER_SEL_FG="$(resolve_token "$C_PICKER_SEL_FG")"

# -- Message UI tokens (ANSI) --
C_MSG_MARKER="$(resolve_token "$C_MSG_MARKER")"
C_MSG_TITLE_SUCCESS="$(resolve_token "$C_MSG_TITLE_SUCCESS")"
C_MSG_TITLE_ACCENT="$(resolve_token "$C_MSG_TITLE_ACCENT")"
C_MSG_DETAIL="$(resolve_token "$C_MSG_DETAIL")"
C_MSG_DIM="$(resolve_token "$C_MSG_DIM")"
```

### Go code — `headers.go`

Replace `colorsFromPalette`:

- `cs.HdrKey` reads `p.Get("C_HDR_KEY")` and wraps as
  `\033[` + val + `m`. No hardcoded `;1`.
- `cs.HdrFG` reads `p.Get("C_HDR_VALUE")`.
- `cs.HdrDim` reads `p.Get("C_HDR_DIM")`.

Remove all `palette.HexToANSI()` calls from this function.

### Go code — `save.go`

Replace `printSaveNotification`:

- Heading color reads `p.Get("C_MSG_TITLE_SUCCESS")`.
- `#` marker reads `p.Get("C_MSG_MARKER")`.
- Detail text reads `p.Get("C_MSG_DETAIL")`.
- Dim text reads `p.Get("C_MSG_DIM")`.

Remove all `palette.HexToANSI()` calls and hardcoded bold.

### Go code — `picker.go`

Replace `ColorsFromPalette`:

- `c.Number` reads `p.Get("C_PICKER_NUM")`.
- `c.Label` reads `p.Get("C_PICKER_LABEL")`.
- `c.URL` reads `p.Get("C_PICKER_URL")`.
- `c.Selected` assembles from `p.Get("C_PICKER_SEL_BG")` (with
  `38→48` replacement for background) + `p.Get("C_PICKER_SEL_FG")`.

Replace `render` heading line:

- `#` marker uses `C_MSG_MARKER` for styling.
- Heading text ("OPEN LINK") uses `C_MSG_TITLE_ACCENT`.

Remove all `palette.HexToANSI()` calls and hardcoded bold.

### Go code — palette package

`palette.HexToANSI()` may still be used by the footnote rendering
code in `internal/filter/`. If so, it stays but is no longer used
by the UI assembly code in `cmd/`. No changes needed to the palette
package itself.

## Documentation Changes

### `docs/themes.md`

- Add full composite token inventory table (header, picker, message
  UI tokens) alongside existing markdown tokens.
- Add the rule: all text styling uses composite tokens; never
  hardcode ANSI modifiers in Go source.
- Document available modifiers: `bold`, `italic`, `underline`.

### `docs/styling.md` (new — absorbs `docs/message-ui.md`)

Broader styling guide covering all UI elements:

- Visual hierarchy (title/detail/secondary)
- Layout patterns (1/3 vertical positioning, left-aligned, full
  terminal width)
- Overlay mechanics (alt screen, /dev/tty, cursor hiding)
- Confirmation screens (stdout for `:pipe` widget)
- Picker UI patterns
- Complete token-to-ANSI pipeline reference
- References `themes.md` for token definitions

### `docs/message-ui.md`

Removed. Content moves to `docs/styling.md`.

### `docs/filters.md`

Update color references to use composite token names instead of base
tokens + hardcoded modifiers. Specifically:

- Header colorization section: reference `C_HDR_KEY`, `C_HDR_VALUE`,
  `C_HDR_DIM` instead of "ACCENT_PRIMARY bold"
- Picker colors section: reference `C_PICKER_NUM`, `C_PICKER_URL`,
  `C_PICKER_SEL_BG` + `C_PICKER_SEL_FG`
- Footnote link rendering: already uses `C_LINK_TEXT`, `C_LINK_URL`
  (no change needed)

### `CLAUDE.md`

- Update pointer from `docs/message-ui.md` to `docs/styling.md`.
- Add directive to Theme System section: **Never hardcode ANSI color
  codes or style modifiers (bold, italic, underline) in Go source.
  All text styling must use composite palette tokens defined in the
  theme file. If a UI element needs styling, add a token to the theme
  and reference it through the palette.**

## Testing

- `make check` passes after all changes.
- Regenerate palette.sh from each theme and verify new tokens appear.
- E2E tests: golden files may need updating if header output changes
  (the ANSI sequences for header keys will differ since bold now
  comes from the token rather than being hardcoded).
- Visual verification via tmux: headers, save notification, and
  picker all render with correct styling.

## Out of Scope

- Footnote rendering in `internal/filter/` — already uses
  `palette.HexToANSI()` for `FG_DIM` on footnote markers. This
  should migrate to composite tokens in a follow-up, but the current
  work focuses on the four hardcoded modifier locations.
- kitty and nvim-mail color syncing (separate systems).

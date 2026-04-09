---
paths:
  - "cmd/pick-link/**"
  - "internal/picker/**"
---

# Link Picker

`pick-link` is a standalone binary that provides an interactive URL
picker for the message viewer. Invoked via `:pipe` so aerc feeds
the raw message on stdin.

## Pipeline

raw HTML -> `filter.HTML` (same filter the viewer uses)
-> extract URLs from footnotes and plain text -> interactive picker
-> `xdg-open` selected URL.

## Keybinding (in `[view]` section of `binds.conf`)

- `Tab` -- open the link picker (`:pipe pick-link<Enter>`)
- `Ctrl-l` -- manually type a URL to open (`open-link`)

## Picker Controls

- `1`-`9`, `0` -- instant-select link by number (0 = 10th)
- `j`/`k` or arrows -- move selection
- `Enter` -- open selected link
- `q` or `Escape` -- cancel

## Key Design Decisions

- Reads keyboard from `/dev/tty` (not stdin) since stdin is the
  piped message content.
- Runs the HTML filter internally to extract clean footnoted URLs
  rather than parsing raw HTML (avoids DTD, image, and tracking URLs).
- Opens URLs directly via `xdg-open` since `:pipe` cannot feed
  output back to aerc's `:open-link`.

## Footnote URLs

Long URLs in the footnote reference section are visually truncated
with `...` to fit within the terminal width. The full URL is embedded
in an OSC 8 hyperlink escape sequence so terminals that support it
can still make the truncated text clickable. The link picker extracts
full URLs from OSC 8 hrefs, so truncation does not affect link opening.

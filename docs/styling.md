# Styling

Guidelines for visual styling across all beautiful-aerc UI
elements. Every color, weight, italic, and underline attribute
comes from the theme — nothing is hardcoded in Go source.

## Principle

Use composite palette tokens for all text styling. Base color
tokens (`ACCENT_PRIMARY`, `FG_DIM`, etc.) are pure hex values —
think CSS variables. Composite tokens (`C_HDR_KEY`, `C_MSG_MARKER`,
etc.) reference base tokens and add modifiers — think CSS classes.

Go code reads pre-resolved ANSI parameter strings via
`palette.ANSI(key)`. It never assembles ANSI codes manually or
hardcodes modifiers like bold, italic, or underline.

See [themes.md](themes.md) for the full token reference and
modifier syntax.

## Visual Hierarchy

Every message screen uses a three-tier hierarchy:

1. **Title** — markdown header style: `# ALL CAPS`. The `#` in
   `C_MSG_MARKER`, the title in bold + semantic color
   (`C_MSG_TITLE_SUCCESS` or `C_MSG_TITLE_ACCENT`). Short and
   scannable.
2. **Detail** — normal case, `C_MSG_DETAIL`. The key information
   (filename, list items). Left-aligned.
3. **Secondary** — normal case, `C_MSG_DIM`. Counts, hints,
   metadata. Left-aligned.

A blank line separates the title from the content below it.

### Examples

```
 # SAVED TO CORPUS          (# C_MSG_MARKER, title: C_MSG_TITLE_SUCCESS)

 20260404-220235.html       (C_MSG_DETAIL)
 10 pending                 (C_MSG_DIM)
```

```
 # OPEN LINK                (# C_MSG_MARKER, title: C_MSG_TITLE_ACCENT)

 1  Download invoice  …     (number: C_PICKER_NUM,
 2  Download receipt  …      label: C_PICKER_LABEL, max 72 chars
 3  support site      …      url: C_PICKER_URL, fills to edge)
```

## Layout

- **Vertically at the 1/3 mark.** Compute the block height (title +
  blank line + content lines) and pad from the top by
  `(rows - blockHeight) / 3`. Query terminal size from the tty fd
  with `TIOCGWINSZ`, or fall back to `AERC_ROWS` / 24.
- **Left-aligned** with a single space indent.
- **Full terminal width.** The picker reads actual terminal width
  from the tty (not `AERC_COLUMNS`, which reflects the viewer pane).
  URLs fill to the terminal edge.

## Color Token Reference

| Role | Token | Usage |
|------|-------|-------|
| Success title | `C_MSG_TITLE_SUCCESS` | confirmations |
| Interactive title | `C_MSG_TITLE_ACCENT` | picker, prompts |
| Title marker | `C_MSG_MARKER` | `#` prefix |
| Detail text | `C_MSG_DETAIL` | filenames, labels |
| Secondary text | `C_MSG_DIM` | counts, hints, URLs |
| Header keys | `C_HDR_KEY` | From, Subject, etc. |
| Header values | `C_HDR_VALUE` | field values |
| Header dim | `C_HDR_DIM` | angle brackets |
| Selection bg | `C_PICKER_SEL_BG` | picker row |
| Selection fg | `C_PICKER_SEL_FG` | picker row |
| Shortcut numbers | `C_PICKER_NUM` | picker digits |
| Link labels | `C_PICKER_LABEL` | picker link text |
| Link URLs | `C_PICKER_URL` | picker URL text |

Always pair color sequences with a `\033[0m` reset.

## Interactive Overlays (picker)

- Use the **alternate screen buffer** (`\033[?1049h` / `\033[?1049l`)
  so aerc's view restores cleanly on exit.
- **Hide the cursor** (`\033[?25l`) during interaction; restore on
  exit.
- Read keyboard from `/dev/tty` opened `O_RDWR` — write UI output
  to the same fd for full terminal independence from aerc.
- Flicker-free updates: cursor-home (`\033[H`) + per-line clear
  (`\033[2K`) instead of full screen clear.

## Confirmation Screens (save)

- Output goes to **stdout** so aerc's `:pipe` terminal widget
  displays it. aerc appends "Process complete, press any key to
  close" automatically.
- Follow the three-tier hierarchy: title, detail, secondary.

## Launching External Processes

- Detach with `SysProcAttr{Setsid: true}` so the child survives
  aerc's process group cleanup.
- Route `mailto:` URLs to `aerc` (IPC compose); everything else
  to `xdg-open`.

## aerc Constraints

aerc has no overlay modal API. The two feedback mechanisms are:

- **`:pipe`** — runs a command, shows stdout in a terminal widget,
  then "Process complete, press any key to close." Best for
  confirmations and interactive UIs.
- **`:pipe -b`** — runs in background, shows "completed with
  status 0" briefly in the status bar. Too brief for user-facing
  messages; avoid.

Interactive overlays work around `:pipe` limitations by opening
`/dev/tty` directly for both input and output, bypassing aerc's
I/O capture entirely.

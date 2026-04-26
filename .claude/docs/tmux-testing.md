# Live UI Verification via tmux

Poplar is a bubbletea TUI. Type-checking and unit tests do not catch
visual regressions — pane width math, lipgloss styling, box-drawing
glyphs, cursor positioning, modal overlays, etc. Verify real renders
before claiming any UI task is done.

The standard workflow: launch poplar in a detached tmux session at a
fixed size, drive it with `send-keys`, capture the pane, inspect.

## Prerequisites

```bash
make install    # poplar must be on PATH at the version you're testing
```

Always reinstall after editing UI code. Stale binaries are the most
common source of "the test passes but it still looks wrong."

## Standard pane size

Use 120×40 unless the task is specifically about narrow-terminal
behavior. It matches the reference wireframes in
`docs/poplar/wireframes.md`.

```bash
tmux kill-session -t poplar 2>/dev/null
tmux new-session -d -s poplar -x 120 -y 40 'poplar'
sleep 0.5    # let bubbletea draw the initial frame
```

## Drive and capture

```bash
# Send keys (literal — no leading colon, no Enter unless you want one)
tmux send-keys -t poplar 'j' 'j' 'Enter'
sleep 0.2

# Capture rendered pane (no ANSI codes — readable diff target)
tmux capture-pane -t poplar -p

# Capture WITH ANSI codes — use when verifying colors/styles
tmux capture-pane -t poplar -p -e
```

`capture-pane -p` prints to stdout. Pipe to a file or `diff` against a
golden capture saved in `testdata/`.

## Common key sequences

| Action | send-keys |
|---|---|
| Open viewer | `'Enter'` |
| Close viewer / quit | `'q'` |
| Open help popover | `'?'` |
| Folder jump | `'I'` (Inbox), `'D'` (Drafts), `'S'` (Sent), etc. |
| Search shelf | `'/'` then characters then `'Enter'` |
| Visual-mode space | `'Space'` (literal token, not `' '`) |
| Escape | `'Escape'` |

## Cleanup

```bash
tmux kill-session -t poplar
```

Always kill the session at the end of your verification. Leftover
sessions accumulate and the next run gets confused about which one
is current.

## When the capture looks wrong

1. Confirm `which poplar` points to `~/.local/bin/poplar` and the
   mtime is recent (post your last edit).
2. Re-run `make install` — the most common failure mode.
3. Capture with `-e` and grep for the ANSI sequence you expect — a
   missing escape often means the wrong style branch fired.
4. Resize the pane (`tmux resize-pane -t poplar -x N -y M`) and
   re-capture if the bug looks layout-dependent.

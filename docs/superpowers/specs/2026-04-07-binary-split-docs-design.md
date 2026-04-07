# Binary Split and Documentation Overhaul

Split the `beautiful-aerc` binary into focused components and overhaul
project documentation for public release.

## Motivation

The `beautiful-aerc` binary currently bundles five subcommands with
different concerns: three stdin/stdout filters (headers, html, plain),
an interactive TUI (pick-link), and a dev utility (save). Splitting
them clarifies the project's architecture and makes each component
independently useful. The docs need to frame the project as a cohesive
email environment rather than just a filter binary.

## Binary Restructuring

### mailrender (renamed from beautiful-aerc)

Keeps the three filter subcommands: `headers`, `html`, `plain`.

- Rename `cmd/beautiful-aerc/` to `cmd/mailrender/`.
- Update `Use` field in root cobra command to `mailrender`.
- Remove `newPickLinkCmd()` and `newSaveCmd()` from `root.go`.
- Delete `cmd/beautiful-aerc/save.go` (move is not needed; save
  becomes a shell script).
- Delete `cmd/beautiful-aerc/picklink.go` (logic moves to
  `cmd/pick-link/`).

### pick-link (new standalone binary)

Interactive URL picker, invoked via `:pipe` in aerc.

- Create `cmd/pick-link/` with `main.go`, `root.go`.
- The root command runs the picker directly (no subcommands).
- Duplicate `loadPalette()` and `termCols()` helpers from the
  mailrender main package (~10 lines each). No shared CLI package.
- Imports `internal/filter` (for `HTMLLinks`) and `internal/picker`
  (for `Run` and `ColorsFromPalette`).

### aerc-save-email (shell script, replaces save subcommand)

Dev utility for saving email parts to the corpus directory.

- Create `.local/bin/aerc-save-email` as a shell script.
- Reads stdin, writes to `corpus/` with a timestamped filename.
- Counts pending files and prints a summary.
- Does not need palette colors or Go infrastructure.

### Deletions

- Delete `internal/corpus/` (package and tests). Only consumer was
  `save.go`.
- Delete `cmd/beautiful-aerc/save.go`.

## Config Updates

### aerc.conf

Update filter commands:

```ini
text/plain=mailrender plain
text/html=mailrender html
.headers=mailrender headers
```

Update inline comments for a newcomer audience. Every section and
non-obvious setting should explain what it does and why it's
configured that way. Remove the comment referencing the
`beautiful-aerc` binary name.

### binds.conf

Update commands:

```ini
# [view] section
b = :pipe -m aerc-save-email<Enter>
<Tab> = :pipe pick-link<Enter>
```

## Makefile

Build targets:

```makefile
build:
    go build -o mailrender ./cmd/mailrender
    go build -o pick-link ./cmd/pick-link
    go build -o fastmail-cli ./cmd/fastmail-cli
    go build -o tidytext ./cmd/tidytext

install:
    GOBIN=$(HOME)/.local/bin go install ./cmd/mailrender
    GOBIN=$(HOME)/.local/bin go install ./cmd/pick-link
    GOBIN=$(HOME)/.local/bin go install ./cmd/fastmail-cli
    GOBIN=$(HOME)/.local/bin go install ./cmd/tidytext

clean:
    rm -f mailrender pick-link fastmail-cli tidytext
```

## E2E Tests

Update `e2e/e2e_test.go`:

- Build `mailrender` instead of `beautiful-aerc`.
- Update temp dir prefix and binary path references.
- Test logic unchanged (same subcommands, same golden files).

## Documentation Overhaul

### README.md

Full rewrite for public release. Position beautiful-aerc as a project
that turns aerc into a polished, productive email environment. The
README should:

- Open with a clear one-line description and motivation.
- List all components with their purpose:
  - **mailrender** — message rendering filters (headers, html, plain)
  - **pick-link** — interactive URL picker for the message viewer
  - **fastmail-cli** — Fastmail JMAP CLI for rules, masked emails,
    folders
  - **tidytext** — Claude-powered prose tidier for compose
  - **nvim-mail** — Neovim compose editor profile with custom syntax
  - **aerc config** — theme system, stylesets, keybindings, icons
  - **aerc-save-email** — dev utility for saving test corpus emails
- Cover prerequisites, installation, theme generation, and basic usage.
- Link to detailed docs for each component.
- Be written at Opus quality — clear, concise, well-structured prose
  that invites contribution.

### CLAUDE.md

Update project structure section:

- Replace `cmd/beautiful-aerc/` with `cmd/mailrender/`.
- Add `cmd/pick-link/`.
- Remove `internal/corpus/` line.
- Update filter protocol section to reference `mailrender`.
- Update component descriptions.

### contributing.md

- Update project layout tree.
- Update build commands and binary names.
- Update architecture section for the split.
- Remove `save.go` and `internal/corpus/` references.

### nvim-mail comments

Improve inline comments in `.config/nvim-mail/init.lua` and
`.config/nvim-mail/syntax/aercmail.vim` so a newcomer can understand
and customize:

- Each major section (plugins, editor settings, quote reflow, buffer
  prep, save cleanup, tidytext, keybindings, khard) should have a
  brief explanation of what it does and why.
- Non-obvious Lua patterns should have inline comments.
- The aercmail syntax file should explain the highlight groups and
  how to customize colors.

### Quick reference

Add `<space>t` (tidytext) to the nvim-mail compose section in
`~/.dotfiles/docs/aerc-quickref.html`, in the "khard & Tools" card
(or a new "Tools" card if that's cleaner).

## Out of Scope

- No changes to `internal/filter/`, `internal/picker/`,
  `internal/palette/`, or any filter logic.
- No changes to the theme system or generator.
- No changes to fastmail-cli or tidytext binaries.
- No new Go packages (helpers are duplicated, not extracted).

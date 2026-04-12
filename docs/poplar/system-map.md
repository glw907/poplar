# Poplar System Map

On-demand reference. Load when you need the package layout or when
you're looking for where a piece of code lives. Kept lean — no
rationale, no decision history (see `decisions/` for that).

## Binary

Single binary: `cmd/poplar/`. Build with `make build`. Install with
`make install` (drops into `~/.local/bin/poplar`).

## Package layout

| Package | Role |
|---------|------|
| `cmd/poplar/` | Cobra CLI wiring, `main`, `root`, subcommands (`themes`, `config init`) |
| `internal/ui/` | Bubbletea components — root `App`, `AccountTab`, `Sidebar`, `MessageList`, `footer`, `styles` |
| `internal/mail/` | `Backend` interface, poplar-native types, `Classify([]Folder) []ClassifiedFolder`, mock backend |
| `internal/mailjmap/` | Async→sync adapter wrapping the forked JMAP worker |
| `internal/mailworker/` | Forked aerc IMAP + JMAP workers and supporting packages (`worker`, `models`, `log`, `parse`, `xdg`, `auth`, `keepalive`) |
| `internal/config/` | `accounts.toml` parsing: `AccountConfig`, `ParseAccounts`, `UIConfig`, `LoadUI`, config init writer |
| `internal/theme/` | Compiled lipgloss themes (15 themes, `Palette` → `NewCompiledTheme` → `*CompiledTheme`) |
| `internal/filter/` | Email cleanup pipeline (`CleanHTML`, `CleanPlain`). Library — awaiting Pass 2.5b-4 viewer consumer |
| `internal/content/` | Block model + lipgloss renderer (`ParseBlocks`, `RenderBody`, `ParseHeaders`, `RenderHeaders`). Library — same |
| `internal/tidy/` | Claude API prose tidier (config, prompt, API call). Library — awaiting Pass 9.5 compose consumer |

## Data flow

```
accounts.toml ─► config.ParseAccounts ─┐
                                        ├─► cmd/poplar wires ─► tea.NewProgram
accounts.toml ─► config.LoadUI ────────┘

mail.Backend (interface)
    ├── internal/mailjmap (Fastmail JMAP, async→sync wrapper)
    ├── internal/mailworker/worker/imap (Gmail IMAP, forked from aerc)
    └── internal/mail/mock.go (development / prototype)

internal/ui/App
    ├── AccountTab
    │   ├── Sidebar  (consumes mail.Folder + config.UIConfig + mail.Classify)
    │   └── MessageList
    └── (viewer, compose, popover — later passes)
```

## Testing

- Unit tests alongside source files: `*_test.go` in the same package.
- Table-driven pattern with `[]struct{ name, input, expected }`.
- No third-party assertion libraries.
- Live UI verification uses the tmux testing workflow in
  `.claude/docs/tmux-testing.md`.

## Build gates

| Command | Purpose |
|---------|---------|
| `make vet` | `go vet ./...` |
| `make test` | `go test ./...` |
| `make check` | vet + test — the commit gate |
| `make build` | `go build -o poplar ./cmd/poplar` |
| `make install` | `go install` into `~/.local/bin/` |
| `make clean` | remove built binary |

## Hooks

- `.claude/hooks/claude-md-size.sh` — enforces 200-line CLAUDE.md limit
- `.claude/hooks/make-check-before-commit.sh` — runs `make check` on pre-commit
- `.claude/hooks/elm-architecture-lint.sh` — guards `internal/ui/` against common Elm violations

## Docs

- `CLAUDE.md` — project identity + convention pointers. One `@`-import.
- `docs/poplar/invariants.md` — binding facts, always auto-loaded.
- `docs/poplar/styling.md` — palette-to-surface map. Load before touching any color.
- `docs/poplar/wireframes.md` — UI reference for every screen.
- `docs/poplar/keybindings.md` — authoritative key map.
- `docs/poplar/STATUS.md` — current pass + next starter prompt.
- `docs/poplar/decisions/` — ADR archive, ~55 files.
- `docs/poplar/system-map.md` — this file.
- `docs/superpowers/plans/` — active plan files.
- `docs/superpowers/specs/` — active spec files.
- `docs/superpowers/archive/` — completed plans and specs.

## Global Claude infrastructure

- `~/.claude/skills/go-conventions/` — mandatory Go rules
- `~/.claude/skills/elm-conventions/` — mandatory `internal/ui/` rules
- `.claude/skills/poplar-pass/` — pass-end ritual, starter-prompt format
- `.claude/skills/fix-corpus/` — corpus-driven rendering fix workflow
- `.claude/rules/poplar-development.md` — trigger-phrase rule pointing at `poplar-pass`

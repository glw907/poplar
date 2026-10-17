---
name: poplar-pass
description: >
  The poplar adapter for the lean pass process: gate, fast lane, full-suite home, worktree setup,
  CI watch, risk-class path map, checklist globs, and close checklist. Use when planning,
  executing, resuming, or closing a poplar pass, or when a fresh poplar session follows a STATUS
  launch prompt.
---

# Poplar pass

Load `pass-core` first; it holds the lifecycle and the rules this adapter fills in. Paths are
relative to the repo root. The cursor and resume prompt are in
`docs/superpowers/specs/poplar-refounding-STATUS.md`; plans live under `docs/superpowers/plans/`.

## Adapter

**Gate**: `bash scripts/gate.sh`, which runs `make check` (`Makefile:68`) through
`cairn-run-gate`. A worktree session's Bash guard refuses the quoted string, so the wrapper is the
call (`scripts/gate.sh`). It opens no browser, so it takes the `light` lane. Measured peak
(`systemd-run --user --wait -P`, cold build, 2026-10-11): 2G, under the 3G cap. Never commit red.

**Fast lane**: the same command. CI's `make check` job takes about 2.5 minutes cold
(`.github/workflows/ci.yml:14-27`), so no narrower lane is worth a classifier.

**Full-suite home**: CI (`.github/workflows/ci.yml:3-5`), which fires on every push and pull
request. It runs `make check` (`:14-27`), `go test -race` (`:29-38`), and the darwin build and test
(`:40-53`). Race and darwin never run locally.

**CI watch**: `ci-green <sha> --pr <n> --wait` as a background task, re-issued on exit 75, per
`cairn-pass`.

**Worktree setup**: none. `make check` builds `tools/bin` on first use (`Makefile:84-96,198-202`).
Vale must be on `PATH`, because `scripts/vale-comments.sh` hard-fails without it.

**Risk-class path map**: `auth-data` covers `internal/keyring/**`,
`internal/backend/jmapsource/{credentials,session}.go`, `internal/store/migrations/**`,
`internal/store/{writer,apply,migrate,recovery}.go`, and `internal/outbox/**`. `runner` covers
`.claude/**`, `scripts/**`, `tools/**`, `Makefile`, `.golangci.yml`, `.vale.ini`, `.vale/**`, and
`.github/workflows/**`. Everything else is `ordinary`. Any `go test -tags live` run carries
`live-account`.

**Checklist globs**: `*.go` loads `go-conventions` and `poplar-go-reviewer`. `internal/ui/**` and
`internal/theme/**` load `elm-conventions` and `bubbletea-design`. `internal/**` and `cmd/**` load
`poplar-reviewer` for its design-language and ADR lens (`.claude/agents/poplar-reviewer.md` Part 3).

**Close checklist**: Beyond `pass-core` "Close":

- CI green on the head.
- When a painted screen changed, run `tui-visual-verify` through `scripts/gate-captures` and do
  one `make tmux-check` (`CLAUDE.md` "Conventions", `Makefile:189-196`).
- Read gallery churn (`Makefile:50-58`).
- Record the `deadcode ./...` count in the PR body (STATUS phase-end ritual, step 2).
- Archive the plan to `docs/superpowers/archive/plans/`. STATUS stays at
  `docs/superpowers/specs/poplar-refounding-STATUS.md`.
- Product touchpoints: the wireframe review before a screen task dispatches
  (`.claude/rules/poplar-development.md`), plus `pass-core`'s design approval and packet read.
- Commit specific files only.

# Design: poplar adopts `pass-core`

For Geoff's approval, 2026-10-11, with his rulings of that day folded in (below). Source: `~/.dotfiles/ROADMAP.md:59-63` (ruling R4) and the
planning prompt `docs/plans/2026-10-11-poplar-adapter-planning-prompt.txt`.

## Risk checked first: pass 2c

Pass 2c (real-terminal verification and the help screen, plan
`docs/plans/2026-08-21-pass-2c-real-terminal-verification.md`) is paused, not running.
Tasks 1-3 landed on `master` on 2026-08-21 (63d02b1 through 32afa8f), and tasks 4-6 wait on Geoff's
23 rulings and the help popover wireframe (STATUS:706-714). No session is live in poplar, and
nothing has been committed since bcb2fd6. The resume prompt (STATUS:718-736) cannot run as written.
It runs on the deleted global "Conducting a pass" rule (STATUS:715, `CLAUDE.md:57-63`), and its
cursor is the ledger under `.superpowers/sdd/`, which is gitignored (`.gitignore:10`) and gone: this
checkout is a fresh clone, and no copy of the ledger exists under `~`. The 22 rulings survive with
their defaults in the plan, and the popover ruling survives as one sentence (STATUS:710-713). The
wireframe draft is lost.

The two passes share no code. They share STATUS's "Next" section and the process the agents run.
**Geoff's ruling: adapter first.** 2c needs a re-plan anyway, and the re-plan should be written
against the process it will run under.

Two more findings bear on the order. `master` is red in CI: the darwin job has failed four
`internal/uerr` log tests on every push since at least e55fc4a, because the tests expect the XDG
state path where macOS resolves `Library/Application Support` (run 33288988216). The local gate has
no darwin leg, so it went unseen. The nightly `govulncheck` has also been red since 2026-10-09, on
three `net/http` advisories fixed in go1.26.9.

## Outcome

A poplar session that is asked to plan, run, or close a pass loads `poplar-pass`, then `pass-core`,
and finds every adapter field filled from poplar's own tooling. The Bubble Tea refactor is logged
in poplar's ROADMAP as its own pass.

## Adapter fields (`.claude/skills/poplar-pass/SKILL.md`, dubplate's shape)

- **Gate**: `bash scripts/gate.sh`, a new no-argument wrapper that runs `cairn-run-gate 'make check'`.
  A worktree session's Bash guard refuses the quoted string (`~/.dotfiles/scripts/gate.sh:3-5`).
  `make check` is `Makefile:68`. The lane comes from task 1's measured peak: `light` if
  `systemd-run --user --wait -P` reports a peak at or under 3G, `heavy` otherwise. The gate opens
  no browser.
- **Fast lane**: the same command. CI's `make check` job takes about 2.5 minutes cold (run
  33288988216), so no narrower lane is worth a classifier.
- **Full-suite home**: CI (`.github/workflows/ci.yml:3-5`), which fires on every push and pull
  request. It runs `make check` (`:14-27`), `go test -race` (`:29-38`), and the darwin build and
  test (`:40-53`). Race and darwin never run locally.
- **CI watch**: `ci-green <sha> --pr <n> --wait` as a background task, re-issued on 75, per
  `cairn-pass`.
- **Worktree setup**: none. `make check` builds `tools/bin` on first use (`Makefile:84-96,198-202`).
  Vale must be on `PATH`, because `scripts/vale-comments.sh` hard-fails without it.
- **Risk-class path map**: `auth-data` covers `internal/keyring/**`,
  `internal/backend/jmapsource/{credentials,session}.go`, `internal/store/migrations/**`,
  `internal/store/{writer,apply,migrate,recovery}.go`, and `internal/outbox/**`. `runner` covers
  `.claude/**`, `scripts/**`, `tools/**`, `Makefile`, `.golangci.yml`, `.vale.ini`, `.vale/**`,
  and `.github/workflows/**`. Everything else is `ordinary`. Any `go test -tags live` run carries
  `live-account`.
- **Checklist globs**: `*.go` loads `go-conventions` and `poplar-go-reviewer`.
  `internal/ui/**` and `internal/theme/**` load `elm-conventions` and `bubbletea-design`.
  `internal/**` and `cmd/**` load `poplar-reviewer` for its design-language and ADR lens
  (`poplar-reviewer.md` Part 3).
- **Close checklist**: CI green on the head. When a painted screen changed, run `tui-visual-verify`
  through `scripts/gate-captures` and do one `make tmux-check` (`CLAUDE.md:46-50`,
  `Makefile:189-196`). Read gallery churn (`Makefile:50-58`). Record the `deadcode ./...` count in
  the PR body (STATUS:19-23). Archive the plan to `docs/archive/plans/`. STATUS stays at
  `docs/specs/poplar-refounding-STATUS.md`. The product touchpoints are the wireframe
  review before a screen task dispatches (`.claude/rules/poplar-development.md:20-23`), plus
  `pass-core`'s design approval and PR read.

## Files touched

- New: `.claude/skills/poplar-pass/SKILL.md`, `scripts/gate.sh`, and `ROADMAP.md`.
- `CLAUDE.md:57-63`: replace "Conducting a pass" with two lines that point at `poplar-pass`.
  `CLAUDE.md:88-89`: drop the `glw907-vendor.sh` line, since that script does not exist.
- `.claude/rules/poplar-development.md:18-24`: repoint "subagent-driven with per-task review" and
  the consolidation list to `poplar-pass`. The wireframe ritual and the split rule stay.
- `.claude/agents/poplar-reviewer.md:5-6`: `claude-opus-5` at `high` becomes `claude-opus-5-5` at
  `medium` (`pass-core` "Models"). Its description becomes a close-review checklist.
- `.claude/agents/poplar-implementer.md`: add `effort: medium`. The gate at `:95` and `:155` runs
  `bash scripts/gate.sh`, and the description drops subagent-driven-development.
- STATUS "Next": one line at the head saying 2c and 2b re-plan under `poplar-pass`.
- `ROADMAP.md`: an `Active` entry for the re-founding (its cursor stays in STATUS), and a `Planned`
  entry for the Bubble Tea refactor pass, as specified below.

**The refactor entry.** The target is `elm-conventions` as infra sweep 2 left it: one top-level
model, with components as plain structs that have imperative methods (`SKILL.md:36-72`). Line 34
of that skill already routes poplar's migration to this pass. The pass converts the six nested
`tea.Model` types beside `App`: `Confirm` (`confirm.go:44-199`), `HelpScreen` (`help.go:118-362`),
and the four placeholders (`placeholder.go:160-375`). It also amends ADR-0011:17-19, whose `Screen`
embeds `tea.Model` (`registry.go:20-23`) and registers into a package-level registry at init,
against Rule 1's "no package-level mutable variables." The pass re-homes two rules from the deleted
workstation doc (`claude/.claude/docs/bubbletea-conventions.md`, removed in dotfiles 7d5bdd0;
poplar's own copy went in 041eb329):

- **JoinHorizontal padding** (§2, lines 161-169). `lipgloss.JoinHorizontal` right-pads every line
  to its block's widest line, so a ragged child is the child's bug. The generic half now lives in
  `elm-conventions` Rule 6. The pass records the lipgloss fact beside poplar's size contract.
- **Icon width** (§8, lines 464-475, ADR-0084 on branch `legacy`). Nerd Font SPUA glyphs are
  undercounted, so `JoinHorizontal` is barred on SPUA rows. The re-founding design language already
  forbids private-use glyphs and keeps each fallback at its glyph's width
  (`2026-07-27-poplar-design-language.md:409-413`). The pass confirms that sentence as the rule's
  home rather than porting `displayCells`.

## Out of scope

This pass prepares infrastructure only; it does not develop poplar (Geoff, 2026-10-11). Project
work for a poplar session: resuming or re-planning 2c, the Bubble Tea refactor itself and its order
against 2c's tasks 4 to 6, the darwin `uerr` fix (named in this pass's PR body, since the head stays
red), the go1.26.9 bump that clears the nightly `govulncheck`, `poplar-go-reviewer`'s missing
`effort`, slimming the 918-line STATUS to `pass-core`'s 60 lines, and the dead `--check-vendor`
path at `scripts/vale-comments.sh:46-52`.

## End-to-end check

After the merge, ask a fresh `claude -p` in poplar's main checkout "what is this repo's pass
process and gate?". It should name `poplar-pass`, `pass-core`, and `bash scripts/gate.sh`. Then, in
a fresh worktree, run that command as printed. It should end on `gate exit: 0`.

## Risk class and estimate

`runner` (`.claude/skills`, `.claude/agents`, `scripts/`), so Geoff reads and merges the PR.
Estimate: **50 minutes**. Measure and wrap the gate: 10. The skill: 15. CLAUDE.md, the rules file,
and the agents: 10. ROADMAP and STATUS: 5. Close: 10.

## Rulings (Geoff, 2026-10-11)

1. Adapter first, then 2c re-plans under it.
2. The refactor's order against 2c is decided during poplar development, not now.
3. The darwin red is noted in the PR body and fixed in a poplar session.
4. The go1.26.9 bump and `poplar-go-reviewer`'s effort land in a poplar pass.

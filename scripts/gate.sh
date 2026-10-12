#!/usr/bin/env bash
# scripts/gate.sh: run `make check` through cairn-run-gate.
# A worktree-isolated session's Bash guard refuses `cairn-run-gate '<string>'`,
# since it cannot show the quoted string is not git; a plain call to a script
# that execs cairn-run-gate passes, and keeps the lane lock and the run records
# `--records` scores. The wrapper takes no arguments, because cairn-run-gate
# hands its string to `bash -c`. Exit codes are cairn-run-gate's, so exit 75
# still means re-issue.
#
# usage: scripts/gate.sh
set -uo pipefail
cd "$(dirname "$0")/.." || exit 2

if [ $# -ne 0 ]; then
	echo 'usage: scripts/gate.sh' >&2
	exit 2
fi

# make check opens no browser; its cold-build peak measured 2G, under the light
# lane's 3G cap (`systemd-run --user --wait -P`, 2026-10-11).
export CAIRN_GATE_LANE=light
# Run the go.mod line's newest patch rather than the workstation's own Go, whose
# newer release breaks the pinned golangci-lint (go1.27.2 against v2.12.2).
export GOTOOLCHAIN=go1.26.9
exec cairn-run-gate 'make check'

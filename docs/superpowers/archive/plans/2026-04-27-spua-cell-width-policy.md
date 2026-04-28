# SPUA-A Cell-Width Policy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace ADR-0079's static "+1 per SPUA-A glyph" assumption with a three-mode iconography system (auto / simple / fancy) that autodetects Nerd Font presence and probes the actual rendered cell width via DSR/CPR at startup.

**Architecture:** New `internal/term/` package providing `HasNerdFont()` (sysfont-backed) and `MeasureSPUACells()` (CPR probe). `internal/ui/icons.go` defines `IconSet` plus `SimpleIcons` (Unicode Narrow) and `FancyIcons` (SPUA-A) tables. `internal/ui/iconwidth.go` is refactored to use a package-level `spuaCellWidth` set once at startup. `cmd/poplar/root.go` resolves the icon mode + cell width before constructing the tea program. New `poplar diagnose` subcommand prints the resolved detection state for empirical verification.

**Tech Stack:** Go 1.26, bubbletea, lipgloss. New dep: `github.com/adrg/sysfont` (MIT). Vendored snippet: ~50 LOC adapted from `github.com/hymkor/go-cursorposition` (MIT) into `internal/term/probe.go`. Test-only: `github.com/creack/pty` (MIT).

**Spec:** `docs/superpowers/specs/2026-04-27-spua-cell-width-policy-design.md`

**Workflow rules:**
- Invoke `go-conventions` before writing/modifying any Go file.
- Invoke `elm-conventions` before modifying any file in `internal/ui/`.
- After each commit, `make check` must pass before moving to the next task.
- Commits land directly on `master` (poplar pre-1.0 convention).

---

## File Map

| Path | Action | Responsibility |
|---|---|---|
| `internal/term/font.go` | create | `HasNerdFont() bool` — sysfont-based detection |
| `internal/term/font_test.go` | create | table-driven font detection tests with mocked enumerator |
| `internal/term/probe.go` | create | `MeasureSPUACells() int` — CPR probe via `/dev/tty` |
| `internal/term/probe_test.go` | create | pty round-trip tests with `creack/pty` |
| `internal/term/resolve.go` | create | `Resolve(cfg) (IconMode, spuaCellWidth)` decision logic |
| `internal/term/resolve_test.go` | create | truth-table tests across all `(mode, hasNF, probe)` combos |
| `internal/ui/icons.go` | create | `IconSet` struct + `SimpleIcons` + `FancyIcons` tables |
| `internal/ui/icons_test.go` | create | enforce Narrow class (simple) + SPUA-A class (fancy) |
| `internal/ui/iconwidth.go` | modify | replace `spuaACorrection` const logic with `spuaCellWidth` var + `SetSPUACellWidth` |
| `internal/ui/iconwidth_test.go` | modify | parameterize tests across `spuaCellWidth = 1` and `2` |
| `internal/ui/app.go` | modify | `NewApp` accepts `IconSet`; thread to children |
| `internal/ui/sidebar.go` | modify | `sidebarIcon` consumes `IconSet`; remove hardcoded SPUA-A |
| `internal/ui/sidebar_search.go` | modify | search-shelf glyph from `IconSet.Search` |
| `internal/ui/msglist.go` | modify | flag-cell glyphs from `IconSet`; remove `mlIcon*` consts |
| `internal/ui/account_tab.go` | modify | propagate `IconSet` to `MessageList`/`Sidebar`/`SidebarSearch` |
| `internal/ui/app_test.go` | modify | parameterize border-alignment tests across modes |
| `internal/ui/msglist_test.go` | modify | parameterize row-equality tests across modes |
| `internal/ui/sidebar_test.go` | modify | parameterize across modes |
| `internal/ui/sidebar_search_test.go` | modify | use `FancyIcons.Search` literal |
| `internal/config/ui.go` | modify | add `Icons string` field to `UIConfig`; default `"auto"`; validate |
| `internal/config/ui_test.go` | modify | new cases for `icons = "auto" | "simple" | "fancy"` and invalid |
| `cmd/poplar/root.go` | modify | resolve mode + cell width pre-tea; pass `IconSet` to `NewApp` |
| `cmd/poplar/diagnose.go` | create | `poplar diagnose` cobra subcommand |
| `cmd/poplar/main.go` | modify | register `diagnose` subcommand |
| `docs/poplar/decisions/0079-display-cells-icon-width.md` | modify | mark `superseded by 0084` |
| `docs/poplar/decisions/0083-displaycells-everywhere-no-lipgloss-join.md` | modify | mark `narrowed by 0084`; note discipline scoped to `spuaCellWidth != 1` |
| `docs/poplar/decisions/0084-icon-mode-policy-with-runtime-probe.md` | create | new ADR |
| `docs/poplar/invariants.md` | modify | replace ADR-0079 invariants; add `Icons` mode + probe references |
| `docs/poplar/bubbletea-conventions.md` | modify | narrow `Join*` rule to `spuaCellWidth != 1` |
| `docs/poplar/testing/icon-modes.md` | create | manual matrix gating doc with diagnose output per row |
| `BACKLOG.md` | modify | close #20; retroactive note on #16 |
| `go.mod`, `go.sum` | modify | add `github.com/adrg/sysfont`, `github.com/creack/pty` (test) |

---

## Task 1: Add `internal/term/` skeleton + sysfont dependency

**Files:**
- Create: `internal/term/doc.go`
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1.1: Add sysfont and creack/pty to go.mod**

Run:
```bash
cd /home/glw907/Projects/poplar
go get github.com/adrg/sysfont@latest
go get github.com/creack/pty@latest
go mod tidy
```

Expected: `go.mod` gains a `require github.com/adrg/sysfont vX.Y.Z` line; `go.sum` updated. Both are MIT-licensed.

- [ ] **Step 1.2: Create package doc**

Create `internal/term/doc.go`:

```go
// Package term provides terminal capability detection used at poplar
// startup: Nerd Font installation discovery (HasNerdFont) and DSR/CPR
// probe of an SPUA-A glyph's rendered cell width (MeasureSPUACells).
//
// The package is consumed only by cmd/poplar; internal/ui receives the
// resolved values via constructor injection (IconSet, spuaCellWidth).
package term
```

- [ ] **Step 1.3: Verify build**

Run: `go build ./...`
Expected: success.

- [ ] **Step 1.4: Commit**

```bash
git add go.mod go.sum internal/term/doc.go
git commit -m "$(cat <<'EOF'
Pass: scaffold internal/term with sysfont + pty deps

Add github.com/adrg/sysfont (MIT) for cross-platform installed-font
enumeration and github.com/creack/pty (MIT, test-only) for the upcoming
CPR probe round-trip tests.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 2: Nerd Font detection (`term.HasNerdFont`)

**Files:**
- Create: `internal/term/font.go`
- Create: `internal/term/font_test.go`

- [ ] **Step 2.1: Write the failing test**

Create `internal/term/font_test.go`:

```go
package term

import "testing"

func TestHasNerdFontFromList(t *testing.T) {
	tests := []struct {
		name     string
		families []string
		want     bool
	}{
		{"empty list", nil, false},
		{"none match", []string{"DejaVu Sans Mono", "Inter", "Hack"}, false},
		{"Nerd Font suffix", []string{"DejaVu Sans Mono", "JetBrainsMonoNL Nerd Font"}, true},
		{"NF abbreviation", []string{"Hack NF"}, true},
		{"case-insensitive", []string{"hack nerd font"}, true},
		{"trailing whitespace tolerated", []string{"  Hack Nerd Font  "}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasNerdFontIn(tt.families)
			if got != tt.want {
				t.Errorf("hasNerdFontIn(%v) = %v, want %v", tt.families, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2.2: Run test to verify it fails**

Run: `go test ./internal/term/ -run TestHasNerdFontFromList -v`
Expected: build failure (`hasNerdFontIn` undefined).

- [ ] **Step 2.3: Implement detection**

Create `internal/term/font.go`:

```go
package term

import (
	"strings"
	"sync"

	"github.com/adrg/sysfont"
)

var (
	cachedHasNF   bool
	cachedHasNFOK bool
	cachedHasNFMu sync.Mutex
)

// HasNerdFont reports whether a Nerd Font is installed on this system.
// First call enumerates installed fonts via sysfont; subsequent calls
// return the cached result. Returns false on enumeration failure.
func HasNerdFont() bool {
	cachedHasNFMu.Lock()
	defer cachedHasNFMu.Unlock()
	if cachedHasNFOK {
		return cachedHasNF
	}
	finder := sysfont.NewFinder(nil)
	fonts := finder.List()
	families := make([]string, 0, len(fonts))
	for _, f := range fonts {
		families = append(families, f.Family)
	}
	cachedHasNF = hasNerdFontIn(families)
	cachedHasNFOK = true
	return cachedHasNF
}

// hasNerdFontIn is the pure-string check; isolated for testability.
// A family qualifies if its lower-cased + trimmed name contains
// "nerd font" or ends with " nf".
func hasNerdFontIn(families []string) bool {
	for _, f := range families {
		s := strings.ToLower(strings.TrimSpace(f))
		if strings.Contains(s, "nerd font") || strings.HasSuffix(s, " nf") {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2.4: Run test to verify it passes**

Run: `go test ./internal/term/ -run TestHasNerdFontFromList -v`
Expected: PASS, all 6 cases.

- [ ] **Step 2.5: Commit**

```bash
git add internal/term/font.go internal/term/font_test.go
git commit -m "$(cat <<'EOF'
Pass: term.HasNerdFont via sysfont enumeration

Pure-string match isolated as hasNerdFontIn for table tests; the
exported HasNerdFont caches its first sysfont.NewFinder.List() result
for the process lifetime.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 3: CPR probe parser (pure-function unit)

**Files:**
- Create: `internal/term/cpr.go`
- Create: `internal/term/cpr_test.go`

This task isolates the byte-stream parser from the I/O so we can unit-test it without a TTY.

- [ ] **Step 3.1: Write the failing test**

Create `internal/term/cpr_test.go`:

```go
package term

import (
	"bytes"
	"testing"
)

func TestParseCPR(t *testing.T) {
	tests := []struct {
		name      string
		in        []byte
		wantRow   int
		wantCol   int
		wantErr   bool
	}{
		{"basic", []byte("\x1b[12;34R"), 12, 34, false},
		{"col 1", []byte("\x1b[1;1R"), 1, 1, false},
		{"three digit col", []byte("\x1b[1;120R"), 1, 120, false},
		{"missing R", []byte("\x1b[12;34"), 0, 0, true},
		{"missing semicolon", []byte("\x1b[1234R"), 0, 0, true},
		{"missing CSI", []byte("12;34R"), 0, 0, true},
		{"empty", nil, 0, 0, true},
		{"only CSI", []byte("\x1b["), 0, 0, true},
		{"trailing junk after R is ignored", []byte("\x1b[1;2Rxyz"), 1, 2, false},
		{"leading junk is skipped", []byte("garbage\x1b[7;8R"), 7, 8, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, col, err := parseCPR(bytes.NewReader(tt.in))
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Fatalf("parseCPR err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && (row != tt.wantRow || col != tt.wantCol) {
				t.Errorf("parseCPR = (%d,%d), want (%d,%d)", row, col, tt.wantRow, tt.wantCol)
			}
		})
	}
}
```

- [ ] **Step 3.2: Run test to verify it fails**

Run: `go test ./internal/term/ -run TestParseCPR -v`
Expected: build failure (`parseCPR` undefined).

- [ ] **Step 3.3: Implement the parser**

Create `internal/term/cpr.go`:

```go
package term

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

var errCPRParse = errors.New("term: failed to parse CPR response")

// parseCPR reads bytes from r until it consumes a complete
// Cursor-Position-Report sequence "ESC [ <row> ; <col> R" and returns
// the (row, col) pair. Bytes preceding the ESC are skipped. Bytes
// following 'R' are left in the reader if r supports it; otherwise they
// are discarded (we only parse the first complete sequence).
func parseCPR(r io.Reader) (row, col int, err error) {
	br := bufio.NewReader(r)

	// Skip until ESC.
	for {
		b, e := br.ReadByte()
		if e != nil {
			return 0, 0, errCPRParse
		}
		if b == 0x1b {
			break
		}
	}
	// Expect '['.
	b, e := br.ReadByte()
	if e != nil || b != '[' {
		return 0, 0, errCPRParse
	}
	rowStr, e := readDigits(br)
	if e != nil || rowStr == "" {
		return 0, 0, errCPRParse
	}
	b, e = br.ReadByte()
	if e != nil || b != ';' {
		return 0, 0, errCPRParse
	}
	colStr, e := readDigits(br)
	if e != nil || colStr == "" {
		return 0, 0, errCPRParse
	}
	b, e = br.ReadByte()
	if e != nil || b != 'R' {
		return 0, 0, errCPRParse
	}
	row, _ = strconv.Atoi(rowStr)
	col, _ = strconv.Atoi(colStr)
	return row, col, nil
}

func readDigits(br *bufio.Reader) (string, error) {
	var s []byte
	for {
		b, err := br.ReadByte()
		if err != nil {
			return "", err
		}
		if b < '0' || b > '9' {
			_ = br.UnreadByte()
			return string(s), nil
		}
		s = append(s, b)
	}
}
```

- [ ] **Step 3.4: Run test to verify it passes**

Run: `go test ./internal/term/ -run TestParseCPR -v`
Expected: PASS, all 10 cases.

- [ ] **Step 3.5: Commit**

```bash
git add internal/term/cpr.go internal/term/cpr_test.go
git commit -m "$(cat <<'EOF'
Pass: term.parseCPR for ESC[row;colR responses

Pure byte-stream parser, isolated from TTY I/O so the round-trip
probe in the next commit can focus on the raw-mode lifecycle. Skips
preamble bytes before ESC; tolerates trailing junk after R.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: CPR probe round-trip (`term.MeasureSPUACells`)

**Files:**
- Create: `internal/term/probe.go`
- Create: `internal/term/probe_test.go`

- [ ] **Step 4.1: Write the failing pty round-trip test**

Create `internal/term/probe_test.go`:

```go
//go:build unix

package term

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/creack/pty"
)

// fakeTerminal scripts a slave-side responder that, on seeing the
// glyph or CSI bytes from the probe, replies with CPRs scripted by
// the test. It is a simple byte-pattern reader; not a full terminal
// emulator.
type fakeTerminal struct {
	slave            *os.File
	colBefore, after int
	delay            time.Duration
}

// run consumes bytes from the master side. The probe sequence is:
//
//   1. Write ESC[6n  (initial position request)
//   2. Write the SPUA-A glyph (variable bytes)
//   3. Write ESC[6n  (post-glyph request)
//
// We respond to each ESC[6n with a CPR using the scripted columns.
func (f *fakeTerminal) run(t *testing.T) {
	buf := make([]byte, 64)
	cpr := 0
	for {
		n, err := f.slave.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}
		// Find ESC [ ... n requests; reply once per request.
		i := 0
		for i < n {
			if i+2 < n && buf[i] == 0x1b && buf[i+1] == '[' {
				// Find the terminator.
				j := i + 2
				for j < n && (buf[j] == ';' || (buf[j] >= '0' && buf[j] <= '9')) {
					j++
				}
				if j < n && buf[j] == 'n' {
					if f.delay > 0 {
						time.Sleep(f.delay)
					}
					col := f.colBefore
					if cpr == 1 {
						col = f.after
					}
					_, _ = f.slave.Write([]byte("\x1b[1;" + intToStr(col) + "R"))
					cpr++
					i = j + 1
					continue
				}
			}
			i++
		}
	}
}

func intToStr(i int) string {
	// avoid strconv import for build-tag isolation
	if i == 0 {
		return "0"
	}
	var b [8]byte
	n := 0
	for i > 0 {
		b[n] = byte('0' + i%10)
		i /= 10
		n++
	}
	out := make([]byte, n)
	for k := 0; k < n; k++ {
		out[k] = b[n-1-k]
	}
	return string(out)
}

func TestMeasureSPUACellsViaPTY(t *testing.T) {
	tests := []struct {
		name      string
		colBefore int
		colAfter  int
		want      int
	}{
		{"narrow glyph (1 cell)", 1, 2, 1},
		{"wide glyph (2 cells)", 1, 3, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			master, slave, err := pty.Open()
			if err != nil {
				t.Fatalf("pty.Open: %v", err)
			}
			defer master.Close()
			defer slave.Close()

			ft := &fakeTerminal{slave: slave, colBefore: tt.colBefore, after: tt.colAfter}
			done := make(chan struct{})
			go func() { ft.run(t); close(done) }()

			got, err := measureSPUACellsOn(master, 200*time.Millisecond)
			if err != nil {
				t.Fatalf("measureSPUACellsOn err = %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}

			_ = master.Close()
			_ = slave.Close()
			select {
			case <-done:
			case <-time.After(time.Second):
			}
		})
	}
}

func TestMeasureSPUACellsTimeout(t *testing.T) {
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer master.Close()
	defer slave.Close()

	// No goroutine reading the slave — probe should time out.
	got, err := measureSPUACellsOn(master, 50*time.Millisecond)
	if err == nil {
		t.Fatalf("want timeout error, got nil (returned %d)", got)
	}
	if got != 0 {
		t.Errorf("on timeout want 0, got %d", got)
	}
}

func TestMeasureSPUACellsMalformed(t *testing.T) {
	master, slave, err := pty.Open()
	if err != nil {
		t.Fatalf("pty.Open: %v", err)
	}
	defer master.Close()
	defer slave.Close()

	go func() {
		buf := make([]byte, 64)
		for {
			n, err := slave.Read(buf)
			if err != nil {
				return
			}
			if bytes.Contains(buf[:n], []byte("\x1b[")) {
				_, _ = slave.Write([]byte("garbage-no-CPR"))
			}
		}
	}()

	_, err = measureSPUACellsOn(master, 100*time.Millisecond)
	if err == nil {
		t.Errorf("want parse error on malformed reply, got nil")
	}
}
```

- [ ] **Step 4.2: Run test to verify it fails**

Run: `go test ./internal/term/ -run TestMeasure -v`
Expected: build failure (`measureSPUACellsOn` undefined).

- [ ] **Step 4.3: Vendor the probe with attribution**

Create `internal/term/probe.go`:

```go
//go:build unix

package term

import (
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// MeasureSPUACells returns the rendered cell width (1 or 2) of an
// SPUA-A glyph in the current terminal+font, or 0 on any failure.
//
// The probe writes ESC[6n, records the cursor column, writes a test
// SPUA-A glyph, writes ESC[6n again, and computes the delta. Falls
// back to 0 if /dev/tty cannot be opened, the terminal does not
// reply within the timeout, or the response is unparseable.
//
// Adapted from github.com/hymkor/go-cursorposition (MIT) — credit
// retained per the prior-art-with-license discipline. The
// AmbiguousWidth pattern (probe-glyph-probe) is the same shape.
func MeasureSPUACells() int {
	w, err := measureSPUACells(200 * time.Millisecond)
	if err != nil {
		return 0
	}
	return w
}

func measureSPUACells(timeout time.Duration) (int, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return 0, err
	}
	defer tty.Close()
	return measureSPUACellsOn(tty, timeout)
}

// testGlyph is a Nerd Font SPUA-A icon (nf-md-mailbox U+F01EE — chosen
// because it is in the same range as poplar's actual icons; per-glyph
// width is uniform within a Nerd Font).
const testGlyph = "\U000F01EE"

var probeMu sync.Mutex

func measureSPUACellsOn(tty *os.File, timeout time.Duration) (int, error) {
	probeMu.Lock()
	defer probeMu.Unlock()

	fd := int(tty.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return 0, err
	}
	defer term.Restore(fd, oldState)

	colBefore, err := requestCPR(tty, timeout)
	if err != nil {
		return 0, err
	}
	if _, err := io.WriteString(tty, testGlyph); err != nil {
		return 0, err
	}
	colAfter, err := requestCPR(tty, timeout)
	if err != nil {
		return 0, err
	}
	delta := colAfter - colBefore
	if delta != 1 && delta != 2 {
		return 0, errors.New("term: unexpected cell delta")
	}
	return delta, nil
}

func requestCPR(tty *os.File, timeout time.Duration) (int, error) {
	if _, err := io.WriteString(tty, "\x1b[6n"); err != nil {
		return 0, err
	}
	type result struct {
		col int
		err error
	}
	ch := make(chan result, 1)
	go func() {
		_, col, err := parseCPR(tty)
		ch <- result{col, err}
	}()
	select {
	case r := <-ch:
		return r.col, r.err
	case <-time.After(timeout):
		return 0, errors.New("term: CPR timeout")
	}
}
```

- [ ] **Step 4.4: Run tests to verify they pass**

Run: `go test ./internal/term/ -run TestMeasure -v`
Expected: PASS for narrow, wide, timeout, and malformed cases.

- [ ] **Step 4.5: Commit**

```bash
git add internal/term/probe.go internal/term/probe_test.go
git commit -m "$(cat <<'EOF'
Pass: term.MeasureSPUACells via DSR/CPR probe

Probe-glyph-probe pattern adapted from hymkor/go-cursorposition (MIT).
Opens /dev/tty directly (not stdin) so it works under stdin redirection
and isolates the raw-mode lifecycle from bubbletea. 200ms timeout;
returns 0 on any failure (parse, timeout, unexpected delta). Round-trip
tested against creack/pty fakes.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 5: Mode resolution (`term.Resolve`)

**Files:**
- Create: `internal/term/resolve.go`
- Create: `internal/term/resolve_test.go`

- [ ] **Step 5.1: Write the failing test**

Create `internal/term/resolve_test.go`:

```go
package term

import "testing"

func TestResolve(t *testing.T) {
	tests := []struct {
		name        string
		cfg         string
		hasNerdFont bool
		probe       int  // 0 = failed
		wantMode    IconMode
		wantWidth   int
	}{
		{"auto + NF + probe=1", "auto", true, 1, IconModeFancy, 1},
		{"auto + NF + probe=2", "auto", true, 2, IconModeFancy, 2},
		{"auto + NF + probe=0", "auto", true, 0, IconModeFancy, 2},     // assume Mono on probe failure
		{"auto + no NF",         "auto", false, 0, IconModeSimple, 1},
		{"auto + no NF + probe=1 ignored", "auto", false, 1, IconModeSimple, 1},
		{"simple forced + NF",   "simple", true, 2, IconModeSimple, 1},
		{"simple forced + no NF","simple", false, 0, IconModeSimple, 1},
		{"fancy forced + probe=1", "fancy", false, 1, IconModeFancy, 1},
		{"fancy forced + probe=2", "fancy", false, 2, IconModeFancy, 2},
		{"fancy forced + probe=0", "fancy", false, 0, IconModeFancy, 2},
		{"fancy forced + NF + probe=1", "fancy", true, 1, IconModeFancy, 1},
		{"unknown defaults to auto+no-NF", "garbage", false, 0, IconModeSimple, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode, width := Resolve(tt.cfg, tt.hasNerdFont, tt.probe)
			if mode != tt.wantMode || width != tt.wantWidth {
				t.Errorf("Resolve(%q,%v,%d)=(%v,%d), want (%v,%d)",
					tt.cfg, tt.hasNerdFont, tt.probe, mode, width, tt.wantMode, tt.wantWidth)
			}
		})
	}
}
```

- [ ] **Step 5.2: Run test to verify it fails**

Run: `go test ./internal/term/ -run TestResolve -v`
Expected: build failure (`Resolve`, `IconMode` undefined).

- [ ] **Step 5.3: Implement the resolver**

Create `internal/term/resolve.go`:

```go
package term

// IconMode is the resolved iconography mode the UI should render.
type IconMode int

const (
	IconModeSimple IconMode = iota
	IconModeFancy
)

func (m IconMode) String() string {
	switch m {
	case IconModeFancy:
		return "fancy"
	default:
		return "simple"
	}
}

// Resolve maps a config-icons value plus runtime detection results to
// the (mode, spuaCellWidth) pair the UI uses.
//
//   cfg          — UIConfig.Icons literal: "auto" | "simple" | "fancy".
//                  Unknown values are treated as "auto".
//   hasNerdFont  — result of HasNerdFont().
//   probe        — result of MeasureSPUACells(): 1, 2, or 0 (failed).
//
// Defaults on probe failure in fancy mode: spuaCellWidth=2 (Mono Nerd
// Font, the legacy assumption). In simple mode: spuaCellWidth=1
// (lipgloss.Width is canonical and the helper degenerates).
func Resolve(cfg string, hasNerdFont bool, probe int) (IconMode, int) {
	mode := IconModeSimple
	switch cfg {
	case "fancy":
		mode = IconModeFancy
	case "simple":
		mode = IconModeSimple
	default: // "auto" or unknown
		if hasNerdFont {
			mode = IconModeFancy
		}
	}
	if mode == IconModeSimple {
		return IconModeSimple, 1
	}
	switch probe {
	case 1, 2:
		return IconModeFancy, probe
	default:
		return IconModeFancy, 2
	}
}
```

- [ ] **Step 5.4: Run test to verify it passes**

Run: `go test ./internal/term/ -run TestResolve -v`
Expected: PASS, all 12 cases.

- [ ] **Step 5.5: Commit**

```bash
git add internal/term/resolve.go internal/term/resolve_test.go
git commit -m "$(cat <<'EOF'
Pass: term.Resolve for icon-mode + cell-width decision

Pure decision function from (config string, hasNerdFont, probe int) to
(IconMode, spuaCellWidth). Truth-table tested across all 12 documented
combinations.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 6: `IconSet` type + `SimpleIcons` / `FancyIcons` tables

**Files:**
- Create: `internal/ui/icons.go`
- Create: `internal/ui/icons_test.go`

This task introduces the icon abstraction without yet wiring it into renderers. Renderer migration is Task 8.

- [ ] **Step 6.1: Write the failing test**

Create `internal/ui/icons_test.go`:

```go
package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// surfaceFields returns one rune (or short string) per IconSet field
// — added to in lockstep with IconSet itself. Used by both
// width/range tests below.
func surfaceFields(s IconSet) map[string]string {
	return map[string]string{
		"Inbox":              s.Inbox,
		"Drafts":             s.Drafts,
		"Sent":               s.Sent,
		"Archive":            s.Archive,
		"Spam":               s.Spam,
		"Trash":              s.Trash,
		"Notification":       s.Notification,
		"Reminder":           s.Reminder,
		"CustomFolder":       s.CustomFolder,
		"Search":             s.Search,
		"FlagFlagged":        s.FlagFlagged,
		"FlagAnswered":       s.FlagAnswered,
		"FlagUnread":         s.FlagUnread,
	}
}

func TestSimpleIcons_AllNarrow(t *testing.T) {
	for name, s := range surfaceFields(SimpleIcons) {
		if s == "" {
			t.Errorf("SimpleIcons.%s is empty", name)
			continue
		}
		if w := lipgloss.Width(s); w != 1 {
			t.Errorf("SimpleIcons.%s = %q has lipgloss.Width=%d, want 1 (must be Narrow class)", name, s, w)
		}
		// Reject any rune in SPUA-A.
		for _, r := range s {
			if r >= 0xF0000 && r <= 0xFFFFD {
				t.Errorf("SimpleIcons.%s = %q contains SPUA-A rune U+%X", name, s, r)
			}
		}
	}
}

func TestFancyIcons_AllSPUA(t *testing.T) {
	for name, s := range surfaceFields(FancyIcons) {
		if s == "" {
			t.Errorf("FancyIcons.%s is empty", name)
			continue
		}
		runes := []rune(s)
		if len(runes) != 1 {
			t.Errorf("FancyIcons.%s = %q is %d runes, want 1", name, s, len(runes))
			continue
		}
		r := runes[0]
		if r < 0xF0000 || r > 0xFFFFD {
			t.Errorf("FancyIcons.%s = U+%X is outside SPUA-A [F0000..FFFFD]", name, r)
		}
	}
}
```

- [ ] **Step 6.2: Run tests to verify they fail**

Run: `go test ./internal/ui/ -run TestSimpleIcons -v`
Expected: build failure (`IconSet` undefined).

- [ ] **Step 6.3: Implement the icon tables**

Create `internal/ui/icons.go`:

```go
package ui

// IconSet is the per-mode iconography vocabulary for poplar's UI
// surfaces. SimpleIcons uses Unicode Narrow-class codepoints — every
// field has lipgloss.Width == 1. FancyIcons uses Nerd Font SPUA-A
// glyphs (U+F0000–U+FFFFD); their rendered cell width is determined
// at startup by term.MeasureSPUACells and applied via spuaCellWidth.
//
// Add a field here whenever a new render surface needs an icon; both
// tables must be updated together. Tests in icons_test.go enforce the
// class invariants.
type IconSet struct {
	Inbox        string
	Drafts       string
	Sent         string
	Archive      string
	Spam         string
	Trash        string
	Notification string
	Reminder     string
	CustomFolder string
	Search       string
	FlagFlagged  string
	FlagAnswered string
	FlagUnread   string
}

// SimpleIcons is the Unicode-Narrow iconography used when no Nerd Font
// is detected (or icons = "simple"). Every rune must be East Asian
// Width Na or N. Verified by TestSimpleIcons_AllNarrow.
var SimpleIcons = IconSet{
	Inbox:        "▣", // U+25A3
	Drafts:       "✎", // U+270E
	Sent:         "→", // U+2192
	Archive:      "▢", // U+25A2
	Spam:         "!", // ASCII; U+26A0 ⚠ is Ambiguous-class on some terminals
	Trash:        "✗", // U+2717
	Notification: "•", // U+2022
	Reminder:     "◷", // U+25F7
	CustomFolder: "▪", // U+25AA
	Search:       "/", // ASCII; canonical search affordance
	FlagFlagged:  "⚑", // U+2691
	FlagAnswered: "↩", // U+21A9
	FlagUnread:   "●", // U+25CF
}

// FancyIcons is the Nerd Font SPUA-A iconography used when a Nerd Font
// is detected (or icons = "fancy"). Verified by TestFancyIcons_AllSPUA.
var FancyIcons = IconSet{
	Inbox:        "\U000F01F0", // nf-md-inbox
	Drafts:       "\U000F03EB", // nf-md-file_document_edit
	Sent:         "\U000F045A", // nf-md-send
	Archive:      "\U000F003C", // nf-md-archive
	Spam:         "\U000F0377", // nf-md-shield-alert
	Trash:        "\U000F0A7A", // nf-md-trash-can-outline
	Notification: "\U000F009A", // nf-md-bell
	Reminder:     "\U000F0474", // nf-md-clock-alert
	CustomFolder: "\U000F0861", // nf-md-folder-outline
	Search:       "\U000F0349", // nf-md-magnify
	FlagFlagged:  "\U000F023B", // nf-md-flag
	FlagAnswered: "\U000F01EE", // nf-md-mailbox (placeholder; legacy used this for unread; reuse here for answered)
	FlagUnread:   "\U000F01EE", // legacy mlIconUnread
}
```

> Note: the FancyIcons SPUA-A codepoints above mirror the literals
> currently embedded in `sidebar.go`/`msglist.go`/`sidebar_search.go`
> (per `grep` of the existing sources). If a reviewer prefers a
> different fancy glyph for a given surface, change it here — the
> test only enforces the SPUA-A class.

- [ ] **Step 6.4: Run tests to verify they pass**

Run: `go test ./internal/ui/ -run "TestSimpleIcons|TestFancyIcons" -v`
Expected: PASS for both.

- [ ] **Step 6.5: Commit**

```bash
git add internal/ui/icons.go internal/ui/icons_test.go
git commit -m "$(cat <<'EOF'
Pass: IconSet with SimpleIcons (Narrow) and FancyIcons (SPUA-A)

Two compiled icon tables; class invariants enforced by tests
(SimpleIcons must be Narrow; FancyIcons must be in U+F0000..FFFFD).
Renderers still hold their hardcoded literals — wired up in the next
commit.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 7: Refactor `iconwidth.go` for runtime-set `spuaCellWidth`

**Files:**
- Modify: `internal/ui/iconwidth.go`
- Modify: `internal/ui/iconwidth_test.go`

- [ ] **Step 7.1: Update tests first (parameterize across cell widths)**

Replace `internal/ui/iconwidth_test.go` with:

```go
package ui

import "testing"

func TestDisplayCells(t *testing.T) {
	// SPUA-A test glyph: U+F01EE.
	const glyph = "\U000F01EE"

	tests := []struct {
		name      string
		cellWidth int
		in        string
		want      int
	}{
		{"ascii w=1", 1, "abc", 3},
		{"ascii w=2", 2, "abc", 3},
		{"empty w=1", 1, "", 0},
		{"empty w=2", 2, "", 0},
		{"glyph alone w=1", 1, glyph, 1},
		{"glyph alone w=2", 2, glyph, 2},
		{"glyph + ascii w=1", 1, "x" + glyph + "y", 3},
		{"glyph + ascii w=2", 2, "x" + glyph + "y", 4},
		{"two glyphs w=2", 2, glyph + glyph, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetSPUACellWidth(tt.cellWidth)
			defer SetSPUACellWidth(1) // restore default for other tests
			got := displayCells(tt.in)
			if got != tt.want {
				t.Errorf("displayCells(%q) @ w=%d = %d, want %d", tt.in, tt.cellWidth, got, tt.want)
			}
		})
	}
}

func TestSetSPUACellWidthRejectsBadValue(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("SetSPUACellWidth(3) should panic")
		}
		SetSPUACellWidth(1) // restore
	}()
	SetSPUACellWidth(3)
}
```

- [ ] **Step 7.2: Replace `internal/ui/iconwidth.go`**

```go
package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Nerd Font icons live in the Supplementary Private Use Area-A
// (U+F0000–U+FFFFD). Their rendered cell width depends on the
// terminal+font+symbol_map configuration. We set spuaCellWidth at
// startup from term.MeasureSPUACells(); see ADR-0084.
//
// In simple mode (no Nerd Font icons present in rendered strings) the
// value is 1 and displayCells degenerates to lipgloss.Width.
const (
	spuaAStart = 0xF0000
	spuaAEnd   = 0xFFFFD
)

var spuaCellWidth = 1

// SetSPUACellWidth sets the per-glyph rendered cell width for SPUA-A
// runes. Must be 1 or 2; any other value panics. Idempotent.
func SetSPUACellWidth(w int) {
	if w != 1 && w != 2 {
		panic("ui: SetSPUACellWidth requires 1 or 2")
	}
	spuaCellWidth = w
}

// displayCells returns the actual terminal display width of s, given
// the runtime-determined SPUA-A cell width.
func displayCells(s string) int {
	return lipgloss.Width(s) + (spuaCellWidth-1)*spuaCount(s)
}

// spuaCount counts SPUA-A runes in s. Fast-paths plain ASCII via a
// byte scan: SPUA-A codepoints are 4-byte UTF-8 sequences, so a string
// with no high-bit byte cannot contain one.
func spuaCount(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return spuaCountSlow(s)
		}
	}
	return 0
}

func spuaCountSlow(s string) int {
	n := 0
	for _, r := range s {
		if r >= spuaAStart && r <= spuaAEnd {
			n++
		}
	}
	return n
}

// displayTruncate truncates the ANSI string s to at most n terminal
// display cells. ansi.Truncate uses runewidth internally and undercounts
// SPUA-A by (spuaCellWidth-1) per glyph; this wrapper decrements the
// runewidth limit until the result is within n cells. At most
// (spuaCellWidth-1)*spuaCount(s) iterations.
func displayTruncate(s string, n int) string {
	limit := n
	for {
		t := ansi.Truncate(s, limit, "")
		if displayCells(t) <= n {
			return t
		}
		limit--
		if limit < 0 {
			return ""
		}
	}
}
```

- [ ] **Step 7.3: Run iconwidth tests**

Run: `go test ./internal/ui/ -run "TestDisplayCells|TestSetSPUA" -v`
Expected: PASS, all cases. Default `spuaCellWidth=1` after package init.

- [ ] **Step 7.4: Run the full UI suite to surface any breakage**

Run: `go test ./internal/ui/ -count=1`
Expected: PASS. Existing callsites use the same `displayCells`/`displayTruncate` signatures; the only behavioral change is `spuaCellWidth` defaulting to 1 (so simple-mode-equivalent) instead of the implicit +1.

If any test fails, the failing assertion was relying on the implicit `spuaCellWidth=2` behavior. Resolve in Task 9 (regression test parameterization).

- [ ] **Step 7.5: Commit**

```bash
git add internal/ui/iconwidth.go internal/ui/iconwidth_test.go
git commit -m "$(cat <<'EOF'
Pass: parameterize displayCells on runtime spuaCellWidth

Replace the static "+1 per SPUA-A rune" correction (ADR-0079, premise
falsified) with a package var spuaCellWidth set once at startup via
SetSPUACellWidth(int). Default 1; simple mode degenerates to
lipgloss.Width. Tests exercise both width regimes.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 8: Wire `IconSet` through App → renderers

**Files:**
- Modify: `internal/ui/app.go`
- Modify: `internal/ui/account_tab.go`
- Modify: `internal/ui/sidebar.go`
- Modify: `internal/ui/msglist.go`
- Modify: `internal/ui/sidebar_search.go`
- Modify: `internal/ui/app_test.go`, `sidebar_test.go`, `msglist_test.go`, `sidebar_search_test.go` (constructors only)

This is a refactor — no behavior change in fancy mode; renderers consume `IconSet` instead of literals. Use `FancyIcons` as the test default to preserve existing test fixtures' visual expectations.

- [ ] **Step 8.1: Modify `App` to accept and store `IconSet`**

In `internal/ui/app.go`, change `NewApp` signature and add `icons IconSet` field on `App`:

```go
// before
func NewApp(t *theme.CompiledTheme, backend mail.Backend, uiCfg config.UIConfig) App {
    // ...
}

// after
func NewApp(t *theme.CompiledTheme, backend mail.Backend, uiCfg config.UIConfig, icons IconSet) App {
    // store icons on the App and pass to NewAccountTab below
}
```

Thread `icons` into `NewAccountTab(...)`. If the field doesn't exist yet, add `icons IconSet` to the App struct.

- [ ] **Step 8.2: Thread through `AccountTab`**

In `internal/ui/account_tab.go`, add `icons IconSet` to `AccountTab` struct and constructor. Forward to `Sidebar`, `MessageList`, `SidebarSearch` constructors.

- [ ] **Step 8.3: Migrate `Sidebar.sidebarIcon` to consume `IconSet`**

In `internal/ui/sidebar.go`, change `sidebarIcon` to a method on `Sidebar` (or take `IconSet` as a parameter):

```go
func (s Sidebar) sidebarIcon(cf mail.ClassifiedFolder) string {
    switch cf.Canonical {
    case "Inbox":
        return s.icons.Inbox
    case "Drafts":
        return s.icons.Drafts
    case "Sent":
        return s.icons.Sent
    case "Archive":
        return s.icons.Archive
    case "Spam":
        return s.icons.Spam
    case "Trash":
        return s.icons.Trash
    }
    lower := strings.ToLower(cf.Folder.Name)
    switch {
    case strings.Contains(lower, "notification"):
        return s.icons.Notification
    case strings.Contains(lower, "remind"):
        return s.icons.Reminder
    default:
        return s.icons.CustomFolder
    }
}
```

Add `icons IconSet` field on `Sidebar`. Update its constructor.

- [ ] **Step 8.4: Migrate `MessageList` flag glyphs**

In `internal/ui/msglist.go`, delete the `mlIconUnread`/`mlIconAnswered`/`mlIconFlagged` constants. Replace each callsite (`renderFlagCell`):

```go
case msg.Flags&mail.FlagFlagged != 0:
    glyph = m.icons.FlagFlagged
case msg.Flags&mail.FlagAnswered != 0:
    glyph = m.icons.FlagAnswered
case isUnread:
    glyph = m.icons.FlagUnread
```

Add `icons IconSet` field on `MessageList`; update constructor.

Also: in `renderRow`, the `subjectWidth` math currently calls `spuaACorrection(flag)` (now renamed `spuaCount`). Update accordingly:

```go
subjectWidth := max(1, m.width - mlFixedWidth - (spuaCellWidth-1)*spuaCount(flag))
```

- [ ] **Step 8.5: Migrate `SidebarSearch` glyph**

In `internal/ui/sidebar_search.go`, replace each `"\U000F0349"` literal with `s.icons.Search`. Add `icons IconSet` field.

- [ ] **Step 8.6: Update test constructors**

Find every `NewApp(...)`, `NewAccountTab(...)`, `NewSidebar(...)`, `NewMessageList(...)`, `NewSidebarSearch(...)` callsite in tests. Pass `FancyIcons` (preserves existing fixture expectations) and call `SetSPUACellWidth(2)` in test setup if a test asserts byte-exact rendering of icon-bearing rows.

Pattern for an `*_test.go` setup helper:

```go
func init() {
    SetSPUACellWidth(2) // legacy fixture expectation
}
```

- [ ] **Step 8.7: Run the full UI suite**

Run: `go test ./internal/ui/ -count=1`
Expected: PASS. Any width-equality test that fails likely needs explicit `SetSPUACellWidth(2)` in its setup.

- [ ] **Step 8.8: `make check`**

Run: `make check`
Expected: vet + tests pass.

- [ ] **Step 8.9: Commit**

```bash
git add internal/ui/
git commit -m "$(cat <<'EOF'
Pass: thread IconSet through App / AccountTab / Sidebar / MessageList

Renderers consume m.icons fields instead of hardcoded SPUA-A literals.
mlIcon* constants and the inline U+F03... literals are gone; the only
icon literals in the tree now live in icons.go (SimpleIcons, FancyIcons).
Tests set spuaCellWidth=2 via init() to preserve fancy-mode fixture
expectations; mode-parameterized regression tests follow in Task 9.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 9: Parameterize width regression tests across modes

**Files:**
- Modify: `internal/ui/app_test.go`
- Modify: `internal/ui/msglist_test.go`

- [ ] **Step 9.1: Identify the existing width-equality tests**

Run: `grep -nE "RowWidth|RightBorder|displayCells.*want" internal/ui/app_test.go internal/ui/msglist_test.go`

The two tests of interest are `TestApp_RightBorderAlignment` (or equivalent) and `TestMessageList_RowWidthEqualAcrossReadStates` (or equivalent — exact names per current source).

- [ ] **Step 9.2: Add a mode-loop wrapper**

Wrap each affected test in a sub-test loop over three configurations:

```go
for _, mode := range []struct {
    name   string
    width  int
    iconSet IconSet
}{
    {"simple_w1", 1, SimpleIcons},
    {"fancy_w1", 1, FancyIcons},
    {"fancy_w2", 2, FancyIcons},
} {
    t.Run(mode.name, func(t *testing.T) {
        SetSPUACellWidth(mode.width)
        defer SetSPUACellWidth(2) // restore the package init() default

        // ... existing test body, with iconSet plumbed into NewApp/NewMessageList ...
    })
}
```

For each mode, every assembled row must satisfy `displayCells(line) == m.width` exactly at terminal widths 80, 100, 120, 160.

- [ ] **Step 9.3: Run width regressions**

Run: `go test ./internal/ui/ -run "TestApp_RightBorder|TestMessageList_RowWidth" -v -count=1`
Expected: PASS for all 3 sub-modes × 4 widths. Each failure is specific: prints mode name + width + offending line.

- [ ] **Step 9.4: `make check`**

Run: `make check`
Expected: vet + full test suite pass.

- [ ] **Step 9.5: Commit**

```bash
git add internal/ui/app_test.go internal/ui/msglist_test.go
git commit -m "$(cat <<'EOF'
Pass: parameterize width-equality tests across icon modes

TestApp_RightBorderAlignment and TestMessageList_RowWidthEqualAcrossReadStates
now run in (simple_w1, fancy_w1, fancy_w2) configurations. Locks in the
invariant that displayCells(row) == m.width regardless of how
SetSPUACellWidth was called at startup.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 10: Add `[ui] icons` config field

**Files:**
- Modify: `internal/config/ui.go`
- Modify: `internal/config/ui_test.go`

- [ ] **Step 10.1: Write the failing test**

Add to `internal/config/ui_test.go`:

```go
func TestLoadUI_Icons(t *testing.T) {
	tests := []struct {
		name    string
		toml    string
		want    string
		wantErr bool
	}{
		{"default when missing", `[ui]`, "auto", false},
		{"explicit auto",        `[ui]` + "\n" + `icons = "auto"`, "auto", false},
		{"simple",               `[ui]` + "\n" + `icons = "simple"`, "simple", false},
		{"fancy",                `[ui]` + "\n" + `icons = "fancy"`, "fancy", false},
		{"invalid",              `[ui]` + "\n" + `icons = "blah"`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempUI(t, tt.toml)
			cfg, err := LoadUI(path)
			gotErr := err != nil
			if gotErr != tt.wantErr {
				t.Fatalf("LoadUI err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && cfg.Icons != tt.want {
				t.Errorf("cfg.Icons = %q, want %q", cfg.Icons, tt.want)
			}
		})
	}
}
```

(`writeTempUI` likely already exists in `ui_test.go`; if not, add a small helper that writes `tt.toml` to a temp file and returns the path.)

- [ ] **Step 10.2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestLoadUI_Icons -v`
Expected: build failure (`cfg.Icons` undefined).

- [ ] **Step 10.3: Add the field**

In `internal/config/ui.go`:

```go
type UIConfig struct {
    Threading bool
    Folders   map[string]FolderConfig
    // Icons is the iconography mode: "auto" (default), "simple", or
    // "fancy". See ADR-0084.
    Icons string
}

func DefaultUIConfig() UIConfig {
    return UIConfig{
        Threading: true,
        Folders:   map[string]FolderConfig{},
        Icons:     "auto",
    }
}

type rawUI struct {
    Threading *bool                   `toml:"threading"`
    Folders   map[string]rawFolderCfg `toml:"folders"`
    Icons     string                  `toml:"icons"`
}
```

In `LoadUI`, after the existing `Threading` plumbing:

```go
if raw.UI.Icons != "" {
    switch raw.UI.Icons {
    case "auto", "simple", "fancy":
        out.Icons = raw.UI.Icons
    default:
        return UIConfig{}, fmt.Errorf("ui.icons: invalid value %q (want \"auto\", \"simple\", or \"fancy\")", raw.UI.Icons)
    }
}
```

- [ ] **Step 10.4: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestLoadUI_Icons -v`
Expected: PASS, all 5 cases.

- [ ] **Step 10.5: Commit**

```bash
git add internal/config/ui.go internal/config/ui_test.go
git commit -m "$(cat <<'EOF'
Pass: [ui] icons config field with auto/simple/fancy validation

Default "auto". Unknown values reject at load time with a clear error.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 11: Wire startup mode resolution in `cmd/poplar/root.go`

**Files:**
- Modify: `cmd/poplar/root.go`

- [ ] **Step 11.1: Resolve mode + cell width before tea**

In `runRoot`, after `config.LoadUI` and before `ui.NewApp`:

```go
hasNF := term.HasNerdFont()
probe := term.MeasureSPUACells()
mode, cellWidth := term.Resolve(uiCfg.Icons, hasNF, probe)

iconSet := ui.SimpleIcons
if mode == term.IconModeFancy {
    iconSet = ui.FancyIcons
}
ui.SetSPUACellWidth(cellWidth)

app := ui.NewApp(t, backend, uiCfg, iconSet)
```

Add the import: `"github.com/glw907/poplar/internal/term"`.

- [ ] **Step 11.2: Build and run the existing binary**

Run:
```bash
go build ./...
make install
```
Expected: clean build. `~/.local/bin/poplar` updated.

- [ ] **Step 11.3: Smoke-test on the workstation terminal**

Launch poplar in kitty manually. Visually verify icons render and borders align. (No automated assertion — this is the live smoke check; the matrix verification is Task 14.)

- [ ] **Step 11.4: `make check`**

Run: `make check`
Expected: PASS.

- [ ] **Step 11.5: Commit**

```bash
git add cmd/poplar/root.go
git commit -m "$(cat <<'EOF'
Pass: startup resolves icon mode + cell width before tea

cmd/poplar/root.go now calls term.HasNerdFont + term.MeasureSPUACells +
term.Resolve, sets ui.SetSPUACellWidth, and threads the resolved
IconSet into ui.NewApp. The probe is the only pre-tea I/O round-trip;
both detection paths fall back safely on failure.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 12: `poplar diagnose` subcommand

**Files:**
- Create: `cmd/poplar/diagnose.go`
- Modify: `cmd/poplar/main.go`

- [ ] **Step 12.1: Implement the subcommand**

Create `cmd/poplar/diagnose.go`:

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/glw907/poplar/internal/config"
	"github.com/glw907/poplar/internal/term"
	"github.com/glw907/poplar/internal/ui"
	"github.com/spf13/cobra"
	xterm "golang.org/x/term"
)

func newDiagnoseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Print terminal + font detection state and resolved icon mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiagnose()
		},
	}
}

func runDiagnose() error {
	fmt.Println("Terminal:")
	fmt.Printf("  TERM           = %s\n", os.Getenv("TERM"))
	fmt.Printf("  COLORTERM      = %s\n", os.Getenv("COLORTERM"))
	fmt.Printf("  is_terminal    = %v\n", xterm.IsTerminal(int(os.Stdout.Fd())))
	fmt.Println()

	fmt.Println("Fonts:")
	hasNF := term.HasNerdFont()
	fmt.Printf("  has_nerd_font  = %v\n", hasNF)
	fmt.Printf("  source         = sysfont\n")
	fmt.Println()

	fmt.Println("Probe:")
	start := time.Now()
	w := term.MeasureSPUACells()
	dur := time.Since(start)
	fmt.Printf("  cell_width     = %d  (0 = probe failed)\n", w)
	fmt.Printf("  duration       = %s\n", dur.Round(100*time.Microsecond))
	fmt.Println()

	configPath, err := defaultConfigPath()
	cfgIcons := "auto"
	if err == nil {
		if uiCfg, err := config.LoadUI(configPath); err == nil {
			cfgIcons = uiCfg.Icons
		}
	}
	mode, cellWidth := term.Resolve(cfgIcons, hasNF, w)

	fmt.Println("Resolved:")
	fmt.Printf("  config.icons   = %s\n", cfgIcons)
	fmt.Printf("  effective_mode = %s\n", mode)
	fmt.Printf("  spua_cell_w    = %d\n", cellWidth)
	iconSet := "SimpleIcons"
	if mode == term.IconModeFancy {
		iconSet = "FancyIcons"
	}
	fmt.Printf("  icon_set       = %s\n", iconSet)

	// Suppress unused warning if SimpleIcons is unreferenced in this file:
	_ = ui.SimpleIcons
	return nil
}
```

- [ ] **Step 12.2: Register the subcommand**

In `cmd/poplar/main.go`:

```go
func main() {
    cmd := newRootCmd()
    cmd.AddCommand(newThemesCmd())
    cmd.AddCommand(newConfigCmd())
    cmd.AddCommand(newDiagnoseCmd())
    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

- [ ] **Step 12.3: Build and run**

Run:
```bash
go build ./...
./poplar diagnose
```

Expected: output similar to:

```
Terminal:
  TERM           = xterm-kitty
  COLORTERM      = truecolor
  is_terminal    = true

Fonts:
  has_nerd_font  = true
  source         = sysfont

Probe:
  cell_width     = 1
  duration       = 1.4ms

Resolved:
  config.icons   = auto
  effective_mode = fancy
  spua_cell_w    = 1
  icon_set       = FancyIcons
```

The exact values will vary per workstation. The output's existence and shape is what matters.

- [ ] **Step 12.4: `make check`**

Run: `make check`
Expected: PASS.

- [ ] **Step 12.5: Commit**

```bash
git add cmd/poplar/diagnose.go cmd/poplar/main.go
git commit -m "$(cat <<'EOF'
Pass: poplar diagnose subcommand

Empirical receipt for icon-mode detection: prints TERM env, font
detection result, CPR probe outcome + duration, and the resolved
(mode, spua_cell_w, icon_set) triple. Used as the gate output for
the manual visual matrix and as a future user-troubleshooting tool.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 13: Write ADR-0084; mark 0079 superseded; mark 0083 narrowed

**Files:**
- Create: `docs/poplar/decisions/0084-icon-mode-policy-with-runtime-probe.md`
- Modify: `docs/poplar/decisions/0079-display-cells-icon-width.md`
- Modify: `docs/poplar/decisions/0083-displaycells-everywhere-no-lipgloss-join.md`

- [ ] **Step 13.1: Write ADR-0084**

Create `docs/poplar/decisions/0084-icon-mode-policy-with-runtime-probe.md`:

```markdown
---
title: Three-mode iconography with Nerd-Font autodetection and CPR cell-width probe
status: accepted
date: 2026-04-27
---

## Context

ADR-0079 introduced `displayCells = lipgloss.Width(s) + 1*count(SPUA-A)`
on the premise that "every modern terminal renders SPUA-A as 2 cells."
That premise is false. On the workstation's actual configuration —
kitty + JetBrainsMonoNL Nerd Font via symbol_map — SPUA-A glyphs render
at 1 cell. The +1 has been an unforced overcount since the helper was
introduced; the alignment defect that ADR-0079 claimed to close
(BACKLOG #16) was never visually verified by the user, and the same
right-border jitter that motivated Pass 4.1 F2 is the same defect
inverted.

Empirical research (see brainstorm transcript 2026-04-27): only
~10–15% of fresh-install Linux/macOS users get Nerd Fonts working
zero-config. wezterm/ghostty/kitty bundle Nerd Font fallbacks; gnome-
terminal/Terminal.app/alacritty/foot do not. No mainstream Linux
distro ships a Nerd Font in its default packages. SPUA-A's actual
rendered cell width is determined by the runtime triple of terminal ×
font × symbol_map config — no static policy spans the matrix.

## Decision

Three-mode iconography with autodetection and a runtime probe.

**Config field.** `[ui] icons = "auto" | "simple" | "fancy"`, default
`"auto"`. Validated at load time.

**Mode resolution at startup** (`cmd/poplar/root.go`, before
`tea.NewProgram`):

1. `term.HasNerdFont()` — sysfont enumeration of installed font
   families; substring match for `"nerd font"` or `" nf"` suffix.
2. `term.MeasureSPUACells()` — DSR/CPR probe via `/dev/tty`: write
   `ESC[6n`, record column, write SPUA-A glyph, write `ESC[6n` again,
   compute delta. 200ms timeout; returns 0 on failure.
3. `term.Resolve(cfg, hasNerdFont, probe)` — pure decision returning
   `(IconMode, spuaCellWidth)`. `auto` picks fancy iff `hasNerdFont`,
   else simple. `simple` always returns `(simple, 1)`. `fancy` always
   returns `(fancy, probe)` with fallback to 2 on probe failure.

**Two icon tables.** `internal/ui/icons.go` defines `IconSet` plus
`SimpleIcons` (Unicode Narrow-class — every rune passes
`lipgloss.Width == 1`) and `FancyIcons` (SPUA-A — every rune in
`[U+F0000, U+FFFFD]`). Class invariants enforced by tests.

**Width math.** `displayCells(s) = lipgloss.Width(s) + (spuaCellWidth-1)
* spuaCount(s)`. In simple mode, `spuaCellWidth = 1` and the helper
degenerates to `lipgloss.Width`. The `displayTruncate` loop terminates
in `(spuaCellWidth-1)*spuaCount(s)` iterations — zero in simple mode.

**Composition.** ADR-0083's `lipgloss.JoinHorizontal`/`JoinVertical`
ban remains in effect when `spuaCellWidth != 1`. In simple mode the
restriction is technically lifted, but the existing manual row-by-row
join code in `AccountTab.View` and `App.renderFrame` is kept unchanged
in this pass; it is correct under both width regimes. Reverting to
`Join*` is a future cleanup pass.

**Limitation, called out explicitly.** Font *presence* is a proxy for
"will glyphs render," not a guarantee. A user with a Nerd Font
installed but a terminal configured to use a non-NF font will see
tofu boxes in fancy mode. We cannot disambiguate
tofu-rendered-at-1-cell from narrow-NF-rendered-at-1-cell via DSR
alone — the cursor moves the same amount in both. The `simple`
override exists precisely for this case.

**`poplar diagnose`.** A new cobra subcommand prints the terminal
environment, font-detection result, probe outcome + duration, and the
resolved `(mode, spuaCellWidth, icon_set)` triple. It is the empirical
receipt that prevents the "declared-fixed-without-verification" failure
mode that produced this ADR's predecessor. Used as the gate output for
the manual visual matrix (`docs/poplar/testing/icon-modes.md`).

## Consequences

- **Supersedes ADR-0079.** The "+1 always" rule is wrong. The new helper
  is parameterized on a measured runtime value.
- **Narrows ADR-0083.** The `Join*` ban applies only when
  `spuaCellWidth != 1`. The discipline is preserved at the codebase
  level (no changes to existing manual joins) and the documentation
  reflects the narrower scope.
- **Default works on every fresh install.** No font requirement. Users
  who have a Nerd Font installed get fancy iconography automatically.
- **`make check` enforces both modes.** Width-equality regression tests
  run in `(simple, w=1)`, `(fancy, w=1)`, and `(fancy, w=2)`
  configurations.
- **Manual matrix is required.** No row in `docs/poplar/testing/icon-modes.md`
  may be left unchecked before shipping the pass. The "fix declared
  without visual verification" failure mode that produced ADR-0079 is
  the case study justifying this rule.
- **One new dependency** (`github.com/adrg/sysfont`, MIT, pure-Go),
  one vendored snippet (~50 LOC adapted from
  `github.com/hymkor/go-cursorposition`, MIT, with attribution).
```

- [ ] **Step 13.2: Mark ADR-0079 superseded**

In `docs/poplar/decisions/0079-display-cells-icon-width.md`, change the frontmatter:

```yaml
---
title: displayCells helper for Nerd Font icon width
status: superseded by 0084
date: 2026-04-26
---
```

Append at end of file:

```markdown

## Superseded

This ADR's premise — "every modern terminal renders SPUA-A as 2 cells"
— was never verified. The fix it claimed to land for BACKLOG #16 was
declared-fixed without user visual confirmation; the same jitter
defect persisted, just inverted. ADR-0084 replaces the static "+1"
correction with a runtime CPR probe. See 0084 for the corrected model.
```

- [ ] **Step 13.3: Mark ADR-0083 narrowed**

In `docs/poplar/decisions/0083-displaycells-everywhere-no-lipgloss-join.md`:

```yaml
---
title: displayCells/displayTruncate everywhere; no lipgloss.Join* on SPUA-A rows
status: narrowed by 0084
date: 2026-04-27
---
```

Append at end:

```markdown

## Narrowed by 0084

The discipline now applies only when `spuaCellWidth != 1`. In simple
mode (the default for systems without a Nerd Font installed),
`lipgloss.Width` is canonical and `Join*` is safe. The existing manual
row-by-row join code is kept in this pass — it is correct under both
width regimes — but a future cleanup pass may revert simple-mode call
sites to `Join*`.
```

- [ ] **Step 13.4: Commit**

```bash
git add docs/poplar/decisions/0084-icon-mode-policy-with-runtime-probe.md \
        docs/poplar/decisions/0079-display-cells-icon-width.md \
        docs/poplar/decisions/0083-displaycells-everywhere-no-lipgloss-join.md
git commit -m "$(cat <<'EOF'
ADR-0084: three-mode iconography with NF autodetect + CPR probe

Supersedes 0079 (premise was wrong); narrows 0083 (Join* ban applies
only when spuaCellWidth != 1). Includes the institutional record on
why ADR-0079's fix was misdiagnosed.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 14: Manual visual verification matrix

**Files:**
- Create: `docs/poplar/testing/icon-modes.md`

This task is the gate for the pass. No row may be left unchecked.

- [ ] **Step 14.1: Capture diagnose output for the workstation**

Run:
```bash
~/.local/bin/poplar diagnose
```

Save the output. Note `effective_mode`, `spua_cell_w`, and `has_nerd_font`.

- [ ] **Step 14.2: Visual verify in kitty + workstation font (matrix row #1)**

Per `.claude/docs/tmux-testing.md`:

```bash
tmux new-session -d -s poplar-test -x 200 -y 50
tmux send-keys -t poplar-test 'poplar' Enter
sleep 2
tmux capture-pane -t poplar-test -pe > /tmp/poplar-snapshot.txt
tmux kill-session -t poplar-test
```

Inspect `/tmp/poplar-snapshot.txt` (or open it in the user's editor of choice). Verify:
1. Right border `│` is at the same column on every row.
2. Sidebar folder rows render with their fancy icons (not tofu).
3. Message-list flag column is uniform width.

- [ ] **Step 14.3: Repeat for as many matrix rows as feasible on this workstation**

The full matrix:

| #  | Terminal       | Font                                  | `icons` cfg | Expected mode | Expected `spua_cell_w` |
|----|---|---|---|---|---|
| 1  | kitty          | JetBrainsMonoNL + symbol_map (default) | auto        | fancy         | 1                       |
| 2  | kitty          | Hack Nerd Font (Mono)                  | auto        | fancy         | 2                       |
| 3  | kitty          | DejaVu Mono (no NF)                    | auto        | simple        | 1                       |
| 4  | kitty          | JetBrainsMonoNL — `simple` forced       | simple      | simple        | 1                       |
| 5  | kitty          | DejaVu Mono — `fancy` forced            | fancy       | fancy         | (probe-driven)         |
| 6  | gnome-terminal | system default                          | auto        | simple        | 1                       |
| 7  | alacritty      | system default                          | auto        | (depends)     | 1                       |
| 8  | tmux ⇒ kitty (#1 setup) | —                              | auto        | fancy         | 1                       |

Rows that cannot be exercised on the workstation (e.g., no alacritty installed) are marked `n/a — deferred` with a note.

- [ ] **Step 14.4: Document results**

Create `docs/poplar/testing/icon-modes.md`:

```markdown
# Icon-Mode Verification Matrix

Manual visual verification gate for ADR-0084. Every row must be
exercised before a pass that touches iconography ships. Rows that
cannot be exercised on the current workstation are marked
`n/a — deferred` with a note.

## Workstation: <hostname>, <date>

### Row 1: kitty + JetBrainsMonoNL + symbol_map (auto)

`poplar diagnose` output:

```
<paste output>
```

Visual verification:
- [x] Right border aligned across all rows
- [x] Fancy icons render (not tofu)
- [x] Flag column uniform width

Snapshot: `/tmp/poplar-snapshot-row1.txt` (or commit a reduced version
in this directory if useful for future pass reviewers).

### Row 2: kitty + Hack Nerd Font (auto)

(Repeat the format. Mark `n/a — deferred` for any row not exercised.)

...
```

Each row gets a section. Where a row is exercised, paste the diagnose
output, list the visual checks, and either embed or path-reference a
snapshot.

- [ ] **Step 14.5: Commit**

```bash
git add docs/poplar/testing/icon-modes.md
git commit -m "$(cat <<'EOF'
Pass: icon-mode verification matrix doc

Gating manual-test record per ADR-0084. Rows exercised on this
workstation include diagnose output and visual-check confirmation;
deferred rows are marked with reason. Future passes that touch
iconography re-run this matrix.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 15: Update invariants and conventions docs

**Files:**
- Modify: `docs/poplar/invariants.md`
- Modify: `docs/poplar/bubbletea-conventions.md`

- [ ] **Step 15.1: Update `invariants.md`**

In the `Architecture` section, find the bullet that references `displayCells` and the SPUA-A correction. Replace with text reflecting the runtime-probe model. Find the "Decision index" table; update the row for ADR-0079 to note `superseded by 0084`, the row for ADR-0083 to note `narrowed by 0084`, and add a new row:

```markdown
| Icon-mode policy: NF autodetect + CPR probe + simple/fancy tables | 0084 |
```

Add an explicit invariant under `Architecture`:

```markdown
- Icon mode is resolved once at startup. `cmd/poplar/root.go` calls
  `term.HasNerdFont`, `term.MeasureSPUACells`, and `term.Resolve` to
  produce `(IconMode, spuaCellWidth)`. `ui.SetSPUACellWidth` is
  called before `tea.NewProgram`. The resolved `IconSet` is threaded
  into `ui.NewApp`. No runtime mode toggling.
- `internal/ui/icons.go` is the only place icon literals live.
  `SimpleIcons` runes are East Asian Width Na/N (lipgloss.Width == 1).
  `FancyIcons` runes are in `[U+F0000, U+FFFFD]`. Both class invariants
  are unit-tested.
```

- [ ] **Step 15.2: Update `bubbletea-conventions.md`**

Find the section that bans `lipgloss.JoinHorizontal`/`JoinVertical` for
SPUA-A rows. Add a sentence:

```markdown
This restriction applies when `spuaCellWidth != 1` (i.e., fancy mode
on Mono Nerd Font terminals). In simple mode and on narrow-Nerd-Font
terminals the ban is technically inert, but the existing manual
row-by-row join code is kept under both regimes — see ADR-0084.
```

- [ ] **Step 15.3: `make check`**

Run: `make check`
Expected: PASS.

- [ ] **Step 15.4: Commit**

```bash
git add docs/poplar/invariants.md docs/poplar/bubbletea-conventions.md
git commit -m "$(cat <<'EOF'
docs: invariants + conventions reflect ADR-0084

Replace ADR-0079 invariants with the runtime-probe model; narrow the
Join* ban scope; mark 0079 superseded and 0083 narrowed in the
decision index.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 16: BACKLOG.md — close #20, retroactive note on #16

**Files:**
- Modify: `BACKLOG.md`

- [ ] **Step 16.1: Close #20**

Change the `#20` line from `- [ ]` to `- [x]` and append a resolution note. The line currently reads:

```
- [ ] **#20** SPUA-A cell-width policy: needs robust cross-terminal solution `#bug` `#poplar` `#bubbletea-norms` *(2026-04-27)*
```

Replace with:

```
- [x] **#20** ~~SPUA-A cell-width policy: needs robust cross-terminal solution~~ `#bug` `#poplar` `#bubbletea-norms` *(2026-04-27)*
  Resolved 2026-04-27 by ADR-0084 / pass `2026-04-27-spua-cell-width-policy`. Three-mode iconography (`[ui] icons = "auto" | "simple" | "fancy"`, default auto) with sysfont-based Nerd Font detection and CPR cell-width probe. ADR-0079 superseded; ADR-0083 narrowed. New `poplar diagnose` subcommand records the empirical receipt; manual matrix in `docs/poplar/testing/icon-modes.md`.
```

- [ ] **Step 16.2: Retroactive note on #16**

Find the `#16` line. It is currently marked `[x]` with a Pass 4 audit-A1 resolution note. Append (do not replace):

```
  Re-audit 2026-04-27: the original "1-cell undercount" framing was workstation-specific (kitty + JetBrainsMonoNL + symbol_map), not universal as ADR-0079 claimed. The `displayCells +1` fix landed here was an over-correction whose visible defect was masked until Pass 4.1 F2. ADR-0084 replaces the static rule with a runtime probe; see that ADR's "Context" section for the institutional record.
```

- [ ] **Step 16.3: Commit**

```bash
git add BACKLOG.md
git commit -m "$(cat <<'EOF'
Backlog: close #20, retroactive re-audit note on #16

#20 resolved by ADR-0084 (three-mode iconography with runtime probe).
#16 gets a re-audit note pointing at the new ADR's case-study
explanation: the original fix was workstation-specific, not universal.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 17: Pass-end ritual

**Files:**
- Modify: `docs/poplar/STATUS.md` (per the `poplar-pass` skill)

- [ ] **Step 17.1: Invoke `poplar-pass` skill**

Tell Claude: "Finish pass."

The skill will:
- Verify ADRs land in `docs/poplar/decisions/`.
- Verify `invariants.md` reflects the new state.
- Update `STATUS.md` with the next pass's starter prompt.
- Archive this plan under `docs/superpowers/plans/archive/` (or per project convention).
- Run `make install` and verify the binary lands in `~/.local/bin/`.
- Commit + push.

- [ ] **Step 17.2: Verify `~/.local/bin/poplar` is updated**

Run:
```bash
~/.local/bin/poplar diagnose
```

Expected: same output shape as Task 12.3, reflecting the workstation's actual detection.

---

## Self-Review Notes

**Spec coverage:**
- Three-mode resolution → Task 5 (`Resolve`) + Task 11 (startup wiring).
- IconSet + SimpleIcons + FancyIcons → Task 6.
- `displayCells` parameterized on `spuaCellWidth` → Task 7.
- Manual joins kept (out of scope for revert) → Task 8 explicit note; Task 13 ADR.
- `poplar diagnose` → Task 12.
- Five-layer testing → Tasks 2/4 (unit), 4 (pty round-trip), 9 (regression parameterization), 14 (manual matrix), 13 (#16 re-audit in ADR-0084).
- ADR-0084 + 0079 superseded + 0083 narrowed → Task 13.
- BACKLOG.md updates → Task 16.
- New deps with license attribution → Task 1 (sysfont, pty), Task 4 (probe attribution).

**Type consistency:**
- `IconMode` defined Task 5; consumed Tasks 11, 12.
- `IconSet` defined Task 6; consumed Tasks 8, 11, 12.
- `SetSPUACellWidth` defined Task 7; called Tasks 8 (test setup), 9 (parameterized tests), 11 (startup), 12 (diagnose indirectly via Resolve).
- `term.HasNerdFont`, `term.MeasureSPUACells`, `term.Resolve` defined in Tasks 2, 4, 5; consumed Tasks 11, 12.

**Out of scope (per spec):**
- Reverting manual row-joins to `lipgloss.Join*` in simple mode.
- Per-icon visual tuning of the SimpleIcons table beyond the Narrow-class invariant.
- Windows.

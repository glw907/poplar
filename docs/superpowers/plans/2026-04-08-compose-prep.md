# compose-prep Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a stdin/stdout Go binary that normalizes aerc compose buffers, replacing ~250 lines of fragile Lua with tested, correct RFC 2822 processing.

**Architecture:** New binary `cmd/compose-prep/` with business logic in `internal/compose/`. Five pipeline steps (unfold, strip brackets, fold addresses, inject Cc/Bcc, reflow quoted text) run sequentially on header and body sections. Each step is an unexported function in its own file with unit tests. The `Prepare` function orchestrates the pipeline and is the only exported function. Uses `go-runewidth` for display-width-correct line wrapping.

**Tech Stack:** Go 1.25+, cobra, go-runewidth, net/mail (stdlib)

**Spec:** `docs/superpowers/specs/2026-04-08-compose-prep-design.md`

**Go conventions:** Read and follow `~/.claude/docs/go-conventions.md` for all Go code. Key rules: no unnecessary interfaces, flags in a struct, `SilenceUsage: true`, `fmt.Errorf("context: %w", err)`, table-driven tests, no assertion libraries, `make check` must pass.

**Dual config requirement:** The Lua integration task (Task 8) modifies `.config/nvim-mail/init.lua`. Apply the same change to BOTH:
- Project repo: `.config/nvim-mail/init.lua`
- Personal dotfiles: `~/.dotfiles/beautiful-aerc/.config/nvim-mail/init.lua`

---

### Task 1: Project Scaffolding and Pass-Through

**Files:**
- Create: `cmd/compose-prep/main.go`
- Create: `cmd/compose-prep/root.go`
- Create: `internal/compose/prepare.go`
- Modify: `Makefile`
- Modify: `go.mod` (via `go get`)

- [ ] **Step 1: Add go-runewidth dependency**

```bash
cd ~/Projects/beautiful-aerc
go get github.com/mattn/go-runewidth
```

- [ ] **Step 2: Create `cmd/compose-prep/main.go`**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 3: Create `cmd/compose-prep/root.go`**

```go
package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/glw907/beautiful-aerc/internal/compose"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type flags struct {
	noCcBcc bool
	debug   bool
}

func newRootCmd() *cobra.Command {
	f := flags{}

	cmd := &cobra.Command{
		Use:          "compose-prep",
		Short:        "Normalize aerc compose buffers for editing",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(f)
		},
	}

	cmd.Flags().BoolVar(&f.noCcBcc, "no-cc-bcc", false, "do not inject empty Cc/Bcc headers")
	cmd.Flags().BoolVar(&f.debug, "debug", false, "write diagnostic messages to stderr")

	return cmd
}

func run(f flags) error {
	if f.debug {
		log.SetPrefix("compose-prep: ")
		log.SetFlags(0)
	} else {
		log.SetOutput(io.Discard)
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("no input (pipe a compose buffer to stdin)")
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	opts := compose.Options{
		InjectCcBcc: !f.noCcBcc,
	}

	output := compose.Prepare(input, opts)

	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Create `internal/compose/prepare.go` with pass-through stub**

```go
package compose

const maxWidth = 72

// Options controls compose-prep behavior.
type Options struct {
	InjectCcBcc bool
}

// Prepare normalizes an aerc compose buffer. On any processing error,
// the original input is returned unchanged.
func Prepare(input []byte, opts Options) []byte {
	return input
}
```

- [ ] **Step 5: Update Makefile**

Add `compose-prep` to the `build`, `install`, and `clean` targets. The full Makefile should be:

```makefile
build:
	go build -o mailrender ./cmd/mailrender
	go build -o pick-link ./cmd/pick-link
	go build -o fastmail-cli ./cmd/fastmail-cli
	go build -o tidytext ./cmd/tidytext
	go build -o compose-prep ./cmd/compose-prep

test:
	go test ./...

vet:
	go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

install:
	GOBIN=$(HOME)/.local/bin go install ./cmd/mailrender
	GOBIN=$(HOME)/.local/bin go install ./cmd/pick-link
	GOBIN=$(HOME)/.local/bin go install ./cmd/fastmail-cli
	GOBIN=$(HOME)/.local/bin go install ./cmd/tidytext
	GOBIN=$(HOME)/.local/bin go install ./cmd/compose-prep

check: vet test

clean:
	rm -f mailrender pick-link fastmail-cli tidytext compose-prep

.PHONY: build test vet lint install check clean
```

- [ ] **Step 6: Verify build and pass-through**

```bash
cd ~/Projects/beautiful-aerc
make build
echo -e "From: alice@dom\nTo: bob@dom\nSubject: Hi\n\nHello" | ./compose-prep
```

Expected: the input is echoed back unchanged (pass-through stub).

```bash
make check
```

Expected: all tests pass (no new tests yet, but existing tests must not break).

- [ ] **Step 7: Commit**

```bash
git add cmd/compose-prep/ internal/compose/prepare.go Makefile go.mod go.sum
git commit -m "Scaffold compose-prep binary with pass-through stub

New stdin/stdout binary for normalizing aerc compose buffers.
Currently passes input through unchanged — pipeline steps will
be added in subsequent commits."
```

---

### Task 2: Header Unfolding

**Files:**
- Create: `internal/compose/unfold.go`
- Create: `internal/compose/unfold_test.go`
- Modify: `internal/compose/prepare.go`

- [ ] **Step 1: Write failing tests in `internal/compose/unfold_test.go`**

```go
package compose

import (
	"testing"
)

func TestUnfoldHeaders(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "no continuation lines",
			input: []string{"From: alice@dom", "To: bob@dom", "Subject: Hello"},
			want:  []string{"From: alice@dom", "To: bob@dom", "Subject: Hello"},
		},
		{
			name:  "space continuation",
			input: []string{"To: alice@dom,", " bob@dom"},
			want:  []string{"To: alice@dom, bob@dom"},
		},
		{
			name:  "tab continuation",
			input: []string{"To: alice@dom,", "\tbob@dom"},
			want:  []string{"To: alice@dom, bob@dom"},
		},
		{
			name:  "multiple continuations",
			input: []string{"To: alice@dom,", " bob@dom,", " charlie@dom"},
			want:  []string{"To: alice@dom, bob@dom, charlie@dom"},
		},
		{
			name:  "mixed headers and continuations",
			input: []string{"From: alice@dom", "To: bob@dom,", " charlie@dom", "Subject: Hi"},
			want:  []string{"From: alice@dom", "To: bob@dom, charlie@dom", "Subject: Hi"},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "continuation with extra whitespace",
			input: []string{"To: alice@dom,", "   bob@dom"},
			want:  []string{"To: alice@dom, bob@dom"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := unfoldHeaders(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("unfoldHeaders() returned %d lines, want %d\ngot:  %q\nwant: %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd ~/Projects/beautiful-aerc
go test ./internal/compose/ -run TestUnfoldHeaders -v
```

Expected: FAIL — `unfoldHeaders` is not defined.

- [ ] **Step 3: Implement `internal/compose/unfold.go`**

```go
package compose

import "strings"

// unfoldHeaders joins RFC 2822 continuation lines (lines starting with
// space or tab) onto the preceding line with a single space.
func unfoldHeaders(headers []string) []string {
	var result []string
	for _, line := range headers {
		if len(result) > 0 && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			result[len(result)-1] += " " + strings.TrimLeft(line, " \t")
		} else {
			result = append(result, line)
		}
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/compose/ -run TestUnfoldHeaders -v
```

Expected: PASS — all 7 cases.

- [ ] **Step 5: Wire unfold into Prepare**

Replace the `Prepare` function in `internal/compose/prepare.go`:

```go
package compose

import (
	"log"
	"strings"
)

// Options controls compose-prep behavior.
type Options struct {
	InjectCcBcc bool
}

// Prepare normalizes an aerc compose buffer. On any processing error,
// the original input is returned unchanged.
func Prepare(input []byte, opts Options) []byte {
	text := strings.ReplaceAll(string(input), "\r\n", "\n")
	lines := strings.Split(text, "\n")

	// strings.Split on trailing \n produces empty last element
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Find header/body boundary (first blank line)
	boundary := -1
	for i, line := range lines {
		if line == "" {
			boundary = i
			break
		}
	}

	if boundary < 0 {
		log.Println("no header/body boundary found, passing through")
		return input
	}

	headers := lines[:boundary]
	body := lines[boundary+1:]

	headers = unfoldHeaders(headers)

	var result []string
	result = append(result, headers...)
	result = append(result, "")
	result = append(result, body...)

	return []byte(strings.Join(result, "\n") + "\n")
}
```

- [ ] **Step 6: Run all tests**

```bash
make check
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/compose/unfold.go internal/compose/unfold_test.go internal/compose/prepare.go
git commit -m "Add header unfolding to compose-prep pipeline

Join RFC 2822 continuation lines (space/tab prefix) onto the
preceding header line with a single space."
```

---

### Task 3: Bracket Stripping

**Files:**
- Create: `internal/compose/bracket.go`
- Create: `internal/compose/bracket_test.go`
- Modify: `internal/compose/prepare.go`

- [ ] **Step 1: Write failing tests in `internal/compose/bracket_test.go`**

```go
package compose

import (
	"testing"
)

func TestStripBrackets(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "bare address after colon",
			input: []string{"From: <alice@dom>"},
			want:  []string{"From: alice@dom"},
		},
		{
			name:  "bare address after comma",
			input: []string{"To: Bob <bob@dom>, <charlie@dom>"},
			want:  []string{"To: Bob <bob@dom>, charlie@dom"},
		},
		{
			name:  "named address preserved",
			input: []string{"To: Alice <alice@dom>"},
			want:  []string{"To: Alice <alice@dom>"},
		},
		{
			name:  "non-address header untouched",
			input: []string{"Subject: <important>"},
			want:  []string{"Subject: <important>"},
		},
		{
			name:  "date header untouched",
			input: []string{"Date: Mon, 1 Jan 2026 12:00:00 +0000"},
			want:  []string{"Date: Mon, 1 Jan 2026 12:00:00 +0000"},
		},
		{
			name:  "quoted display name with comma",
			input: []string{`To: "Smith, John" <john@dom>`},
			want:  []string{`To: "Smith, John" <john@dom>`},
		},
		{
			name:  "multiple bare addresses",
			input: []string{"To: <alice@dom>, <bob@dom>"},
			want:  []string{"To: alice@dom, bob@dom"},
		},
		{
			name:  "mixed bare and named",
			input: []string{"To: Alice <alice@dom>, <bob@dom>, Charlie <charlie@dom>"},
			want:  []string{"To: alice@dom, bob@dom, Charlie <charlie@dom>"},
		},
		{
			name:  "empty value passes through",
			input: []string{"To:"},
			want:  []string{"To:"},
		},
		{
			name:  "empty value with space passes through",
			input: []string{"To: "},
			want:  []string{"To: "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripBrackets(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("stripBrackets() returned %d lines, want %d\ngot:  %q\nwant: %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/compose/ -run TestStripBrackets -v
```

Expected: FAIL — `stripBrackets` is not defined.

- [ ] **Step 3: Implement `internal/compose/bracket.go`**

```go
package compose

import (
	"net/mail"
	"strings"
)

// addressHeaders lists headers that contain email addresses.
var addressHeaders = map[string]bool{
	"from": true, "to": true, "cc": true, "bcc": true,
}

// splitHeader splits "Key: value" into key, value, ok.
func splitHeader(line string) (string, string, bool) {
	idx := strings.Index(line, ":")
	if idx < 1 {
		return "", "", false
	}
	return line[:idx], strings.TrimSpace(line[idx+1:]), true
}

// formatAddr formats a mail.Address for display. Bare addresses (no
// name) are returned without angle brackets. Named addresses use the
// standard RFC 5322 format.
func formatAddr(a *mail.Address) string {
	if a.Name == "" {
		return a.Address
	}
	return a.String()
}

// stripBrackets removes angle brackets from bare email addresses on
// address headers (From, To, Cc, Bcc). Named addresses are untouched.
// If parsing fails, the line passes through unchanged.
func stripBrackets(headers []string) []string {
	result := make([]string, len(headers))
	for i, line := range headers {
		key, value, ok := splitHeader(line)
		if !ok || !addressHeaders[strings.ToLower(key)] || strings.TrimSpace(value) == "" {
			result[i] = line
			continue
		}
		addrs, err := mail.ParseAddressList(value)
		if err != nil {
			result[i] = line
			continue
		}
		parts := make([]string, len(addrs))
		for j, a := range addrs {
			parts[j] = formatAddr(a)
		}
		result[i] = key + ": " + strings.Join(parts, ", ")
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/compose/ -run TestStripBrackets -v
```

Expected: PASS — all 10 cases. If the "mixed bare and named" case fails because `net/mail` preserves the name `Alice` differently, adjust the expected value. The `net/mail` package parses `Alice <alice@dom>` with `Name: "Alice"` and `addr.String()` returns `"Alice" <alice@dom>` (with quotes). If this happens, update the test:

```go
		{
			name:  "mixed bare and named",
			input: []string{"To: Alice <alice@dom>, <bob@dom>, Charlie <charlie@dom>"},
			want:  []string{`To: "Alice" <alice@dom>, bob@dom, "Charlie" <charlie@dom>`},
		},
```

Check what `net/mail` actually produces and match the test to its output. The goal is bracket stripping from bare addresses — named address formatting is whatever `net/mail` emits.

- [ ] **Step 5: Wire stripBrackets into Prepare**

In `internal/compose/prepare.go`, add the `stripBrackets` call after `unfoldHeaders`:

```go
	headers = unfoldHeaders(headers)
	headers = stripBrackets(headers)
```

- [ ] **Step 6: Run all tests**

```bash
make check
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/compose/bracket.go internal/compose/bracket_test.go internal/compose/prepare.go
git commit -m "Add bracket stripping to compose-prep pipeline

Use net/mail.ParseAddressList for correct RFC 5322 address parsing.
Bare <email> addresses have brackets removed; named addresses are
preserved. Non-address headers pass through unchanged."
```

---

### Task 4: Address Folding

**Files:**
- Create: `internal/compose/fold.go`
- Create: `internal/compose/fold_test.go`
- Modify: `internal/compose/prepare.go`

- [ ] **Step 1: Write failing tests in `internal/compose/fold_test.go`**

```go
package compose

import (
	"testing"
)

func TestFoldAddresses(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "short list stays on one line",
			input: []string{"To: alice@dom, bob@dom"},
			want:  []string{"To: alice@dom, bob@dom"},
		},
		{
			name: "long list wraps at 72 columns",
			input: []string{
				"To: Alice Example <alice@example.com>, Bob Example <bob@example.com>, Charlie Example <charlie@example.com>",
			},
			want: []string{
				`To: "Alice Example" <alice@example.com>,`,
				`    "Bob Example" <bob@example.com>,`,
				`    "Charlie Example" <charlie@example.com>`,
			},
		},
		{
			name:  "single recipient unchanged",
			input: []string{"To: alice@example.com"},
			want:  []string{"To: alice@example.com"},
		},
		{
			name:  "non-address header untouched",
			input: []string{"Subject: This is a very long subject line that exceeds seventy-two characters easily"},
			want:  []string{"Subject: This is a very long subject line that exceeds seventy-two characters easily"},
		},
		{
			name:  "Cc indent matches key length",
			input: []string{
				"Cc: Alice Example <alice@example.com>, Bob Example <bob@example.com>, Charlie Example <charlie@example.com>",
			},
			want: []string{
				`Cc: "Alice Example" <alice@example.com>,`,
				`    "Bob Example" <bob@example.com>,`,
				`    "Charlie Example" <charlie@example.com>`,
			},
		},
		{
			name:  "Bcc indent matches key length",
			input: []string{
				"Bcc: Alice Example <alice@example.com>, Bob Example <bob@example.com>, Charlie Example <charlie@example.com>",
			},
			want: []string{
				`Bcc: "Alice Example" <alice@example.com>,`,
				`     "Bob Example" <bob@example.com>,`,
				`     "Charlie Example" <charlie@example.com>`,
			},
		},
		{
			name:  "empty To passes through",
			input: []string{"To:"},
			want:  []string{"To:"},
		},
		{
			name:  "From header not folded",
			input: []string{"From: alice@example.com"},
			want:  []string{"From: alice@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := foldAddresses(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("foldAddresses() returned %d lines, want %d\ngot:  %q\nwant: %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d:\n  got:  %q\n  want: %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
```

**Note:** The expected output for long-list tests uses `"Alice Example"` with quotes because `net/mail`'s `Address.String()` quotes multi-word display names. Run the tests after implementation to see the exact output `net/mail` produces, and adjust expected values to match. The important thing is that lines wrap at or below 72 columns and indent aligns with the first address.

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/compose/ -run TestFoldAddresses -v
```

Expected: FAIL — `foldAddresses` is not defined.

- [ ] **Step 3: Implement `internal/compose/fold.go`**

```go
package compose

import (
	"net/mail"
	"strings"

	"github.com/mattn/go-runewidth"
)

// foldAddresses wraps To, Cc, and Bcc headers at recipient boundaries
// to fit within 72 columns. Continuation lines are indented to align
// under the first address. Single-recipient and non-address headers
// pass through unchanged.
func foldAddresses(headers []string) []string {
	foldable := map[string]bool{"to": true, "cc": true, "bcc": true}

	var result []string
	for _, line := range headers {
		key, value, ok := splitHeader(line)
		if !ok || !foldable[strings.ToLower(key)] || strings.TrimSpace(value) == "" {
			result = append(result, line)
			continue
		}
		addrs, err := mail.ParseAddressList(value)
		if err != nil || len(addrs) < 2 {
			result = append(result, line)
			continue
		}

		indent := strings.Repeat(" ", len(key)+2)
		formatted := make([]string, len(addrs))
		for i, a := range addrs {
			formatted[i] = formatAddr(a)
		}

		cur := key + ": " + formatted[0]
		for j := 1; j < len(formatted); j++ {
			candidate := cur + ", " + formatted[j]
			if runewidth.StringWidth(candidate) <= maxWidth {
				cur = candidate
			} else {
				result = append(result, cur+",")
				cur = indent + formatted[j]
			}
		}
		result = append(result, cur)
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/compose/ -run TestFoldAddresses -v
```

Expected: PASS. If any test fails due to `net/mail` quoting differences, adjust the expected values. Verify that wrapped lines are all <= 72 columns wide (check `runewidth.StringWidth` on each output line).

- [ ] **Step 5: Wire foldAddresses into Prepare**

In `internal/compose/prepare.go`, add after `stripBrackets`:

```go
	headers = unfoldHeaders(headers)
	headers = stripBrackets(headers)
	headers = foldAddresses(headers)
```

- [ ] **Step 6: Run all tests**

```bash
make check
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/compose/fold.go internal/compose/fold_test.go internal/compose/prepare.go
git commit -m "Add address folding to compose-prep pipeline

Wrap To/Cc/Bcc headers at recipient boundaries to fit within
72 columns. Uses go-runewidth for correct display-width
measurement of international names."
```

---

### Task 5: Cc/Bcc Injection

**Files:**
- Create: `internal/compose/inject.go`
- Create: `internal/compose/inject_test.go`
- Modify: `internal/compose/prepare.go`

- [ ] **Step 1: Write failing tests in `internal/compose/inject_test.go`**

```go
package compose

import (
	"testing"
)

func TestInjectCcBcc(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "both missing inserted after To",
			input: []string{"From: alice@dom", "To: bob@dom", "Subject: Hi"},
			want:  []string{"From: alice@dom", "To: bob@dom", "Cc:", "Bcc:", "Subject: Hi"},
		},
		{
			name:  "Cc present Bcc missing",
			input: []string{"From: alice@dom", "To: bob@dom", "Cc: charlie@dom", "Subject: Hi"},
			want:  []string{"From: alice@dom", "To: bob@dom", "Cc: charlie@dom", "Bcc:", "Subject: Hi"},
		},
		{
			name:  "both present no change",
			input: []string{"From: alice@dom", "To: bob@dom", "Cc:", "Bcc:", "Subject: Hi"},
			want:  []string{"From: alice@dom", "To: bob@dom", "Cc:", "Bcc:", "Subject: Hi"},
		},
		{
			name: "To with continuation lines",
			input: []string{
				"From: alice@dom",
				"To: bob@dom,",
				"    charlie@dom",
				"Subject: Hi",
			},
			want: []string{
				"From: alice@dom",
				"To: bob@dom,",
				"    charlie@dom",
				"Cc:",
				"Bcc:",
				"Subject: Hi",
			},
		},
		{
			name:  "no To header",
			input: []string{"From: alice@dom", "Subject: Hi"},
			want:  []string{"From: alice@dom", "Subject: Hi"},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "empty To",
			input: []string{"From: alice@dom", "To:", "Subject: Hi"},
			want:  []string{"From: alice@dom", "To:", "Cc:", "Bcc:", "Subject: Hi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectCcBcc(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("injectCcBcc() returned %d lines, want %d\ngot:  %q\nwant: %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/compose/ -run TestInjectCcBcc -v
```

Expected: FAIL — `injectCcBcc` is not defined.

- [ ] **Step 3: Implement `internal/compose/inject.go`**

```go
package compose

import "strings"

// injectCcBcc inserts empty Cc: and Bcc: headers after the To: block
// if they are not already present.
func injectCcBcc(headers []string) []string {
	hasCc, hasBcc := false, false
	toEnd := -1

	for i, line := range headers {
		key, _, ok := splitHeader(line)
		if ok {
			switch strings.ToLower(key) {
			case "cc":
				hasCc = true
			case "bcc":
				hasBcc = true
			case "to":
				toEnd = i
			}
		} else if toEnd >= 0 && i == toEnd+1 && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// Continuation line immediately following To: block
			toEnd = i
		}
	}

	if hasCc && hasBcc {
		return headers
	}
	if toEnd < 0 {
		return headers
	}

	// Scan forward from To: for continuation lines
	for j := toEnd + 1; j < len(headers); j++ {
		if len(headers[j]) > 0 && (headers[j][0] == ' ' || headers[j][0] == '\t') {
			toEnd = j
		} else {
			break
		}
	}

	result := make([]string, 0, len(headers)+2)
	result = append(result, headers[:toEnd+1]...)
	if !hasCc {
		result = append(result, "Cc:")
	}
	if !hasBcc {
		result = append(result, "Bcc:")
	}
	result = append(result, headers[toEnd+1:]...)
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/compose/ -run TestInjectCcBcc -v
```

Expected: PASS — all 7 cases.

- [ ] **Step 5: Wire injectCcBcc into Prepare**

In `internal/compose/prepare.go`, add conditionally after `foldAddresses`:

```go
	headers = unfoldHeaders(headers)
	headers = stripBrackets(headers)
	headers = foldAddresses(headers)
	if opts.InjectCcBcc {
		headers = injectCcBcc(headers)
	}
```

- [ ] **Step 6: Run all tests**

```bash
make check
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/compose/inject.go internal/compose/inject_test.go internal/compose/prepare.go
git commit -m "Add Cc/Bcc injection to compose-prep pipeline

Insert empty Cc: and Bcc: headers after the To: block when absent.
Controlled by Options.InjectCcBcc (default on, --no-cc-bcc to skip)."
```

---

### Task 6: Quoted Text Reflow

**Files:**
- Create: `internal/compose/reflow.go`
- Create: `internal/compose/reflow_test.go`
- Modify: `internal/compose/prepare.go`

- [ ] **Step 1: Write failing tests in `internal/compose/reflow_test.go`**

```go
package compose

import (
	"testing"
)

func TestQuotePrefix(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "single level", input: "> hello", want: "> "},
		{name: "double level spaced", input: "> > hello", want: "> > "},
		{name: "double level compact", input: ">> hello", want: ">> "},
		{name: "triple level", input: "> > > hello", want: "> > > "},
		{name: "no prefix", input: "hello", want: ""},
		{name: "just arrow", input: ">text", want: ">"},
		{name: "empty line", input: "", want: ""},
		{name: "blank quoted", input: "> ", want: "> "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quotePrefix(tt.input)
			if got != tt.want {
				t.Errorf("quotePrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestQuoteDepth(t *testing.T) {
	tests := []struct {
		prefix string
		want   int
	}{
		{"> ", 1},
		{"> > ", 2},
		{">> ", 2},
		{"> > > ", 3},
		{">", 1},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := quoteDepth(tt.prefix)
			if got != tt.want {
				t.Errorf("quoteDepth(%q) = %d, want %d", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestReflowQuoted(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name: "reflow ragged quoted lines",
			input: []string{
				"> This is a long quoted line that was wrapped at an odd",
				"> point by the original sender's email client and",
				"> looks ragged.",
			},
			want: []string{
				"> This is a long quoted line that was wrapped at an odd point by the",
				"> original sender's email client and looks ragged.",
			},
		},
		{
			name: "preserve blank quoted lines as paragraph breaks",
			input: []string{
				"> First paragraph.",
				">",
				"> Second paragraph.",
			},
			want: []string{
				"> First paragraph.",
				">",
				"> Second paragraph.",
			},
		},
		{
			name: "nested quotes reflowed independently",
			input: []string{
				"> > Inner quote that is too long and should be",
				"> > reflowed by the tool.",
				"> Outer quote.",
			},
			want: []string{
				"> > Inner quote that is too long and should be reflowed by the tool.",
				"> Outer quote.",
			},
		},
		{
			name:  "unquoted body lines untouched",
			input: []string{"Hello there.", "This is my reply."},
			want:  []string{"Hello there.", "This is my reply."},
		},
		{
			name: "decorative line preserved",
			input: []string{
				"> ----------",
				"> Some text.",
			},
			want: []string{
				"> ----------",
				"> Some text.",
			},
		},
		{
			name:  "empty body",
			input: []string{},
			want:  []string{},
		},
		{
			name: "mixed quoted and unquoted",
			input: []string{
				"My reply.",
				"",
				"> Quoted text that should be",
				"> reflowed into one line.",
			},
			want: []string{
				"My reply.",
				"",
				"> Quoted text that should be reflowed into one line.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reflowQuoted(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("reflowQuoted() returned %d lines, want %d\ngot:  %q\nwant: %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d:\n  got:  %q\n  want: %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/compose/ -run "TestQuotePrefix|TestQuoteDepth|TestReflowQuoted" -v
```

Expected: FAIL — functions not defined.

- [ ] **Step 3: Implement `internal/compose/reflow.go`**

```go
package compose

import (
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// quotePrefix extracts the leading "> > " style prefix from a line.
// Returns empty string if the line is not quoted.
func quotePrefix(line string) string {
	if len(line) == 0 || line[0] != '>' {
		return ""
	}
	i := 1
	for i < len(line) && (line[i] == '>' || line[i] == ' ') {
		i++
	}
	return line[:i]
}

// quoteDepth counts the number of '>' characters in a prefix.
func quoteDepth(prefix string) int {
	n := 0
	for _, c := range prefix {
		if c == '>' {
			n++
		}
	}
	return n
}

// canonicalPrefix builds a normalized prefix for a given depth:
// depth 1 = "> ", depth 2 = "> > ", etc.
func canonicalPrefix(depth int) string {
	if depth <= 0 {
		return ""
	}
	return strings.Repeat("> ", depth)
}

// hasAlphanumeric returns true if s contains at least one letter or digit.
func hasAlphanumeric(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// wrapText wraps text to fit within width columns when prepended with
// prefix. Breaks only at spaces. Returns a slice of lines, each
// starting with prefix.
func wrapText(text, prefix string, width int) []string {
	avail := width - runewidth.StringWidth(prefix)
	if avail <= 0 {
		return []string{prefix + text}
	}

	var result []string
	for runewidth.StringWidth(text) > avail {
		breakIdx := -1
		w := 0
		for i, r := range text {
			w += runewidth.RuneWidth(r)
			if w > avail {
				break
			}
			if r == ' ' {
				breakIdx = i
			}
		}
		if breakIdx < 0 {
			break // word exceeds available width, emit as-is
		}
		result = append(result, prefix+strings.TrimRight(text[:breakIdx+1], " "))
		text = strings.TrimLeft(text[breakIdx+1:], " ")
	}
	if len(text) > 0 {
		result = append(result, prefix+text)
	}
	return result
}

// reflowQuoted joins consecutive quoted lines at the same depth into
// paragraphs and re-wraps them at 72 columns. Blank quoted lines and
// decorative lines (no alphanumeric content) are preserved as breaks.
// Unquoted lines pass through unchanged.
func reflowQuoted(body []string) []string {
	var result []string
	i := 0
	for i < len(body) {
		prefix := quotePrefix(body[i])
		if prefix == "" {
			result = append(result, body[i])
			i++
			continue
		}

		depth := quoteDepth(prefix)
		canon := canonicalPrefix(depth)
		text := strings.TrimSpace(body[i][len(prefix):])

		// Blank quoted line — preserve as paragraph break
		if text == "" {
			result = append(result, strings.TrimRight(canon, " "))
			i++
			continue
		}

		// Decorative line (no letters or digits) — preserve as-is
		if !hasAlphanumeric(text) {
			avail := maxWidth - runewidth.StringWidth(canon)
			if runewidth.StringWidth(text) > avail {
				text = runewidth.Truncate(text, avail, "")
			}
			result = append(result, canon+text)
			i++
			continue
		}

		// Join consecutive lines at the same quote depth
		j := i + 1
		for j < len(body) {
			np := quotePrefix(body[j])
			if quoteDepth(np) != depth {
				break
			}
			nt := strings.TrimSpace(body[j][len(np):])
			if nt == "" || !hasAlphanumeric(nt) {
				break
			}
			text = strings.TrimRight(text, " ") + " " + strings.TrimLeft(nt, " ")
			j++
		}

		text = strings.TrimSpace(text)
		wrapped := wrapText(text, canon, maxWidth)
		result = append(result, wrapped...)
		i = j
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/compose/ -run "TestQuotePrefix|TestQuoteDepth|TestReflowQuoted" -v
```

Expected: PASS — all cases. If reflow line breaks differ from expected (e.g., different word-break positions), adjust the expected output to match the actual 72-column wrapping.

- [ ] **Step 5: Wire reflowQuoted into Prepare**

In `internal/compose/prepare.go`, add after the header pipeline:

```go
	headers = unfoldHeaders(headers)
	headers = stripBrackets(headers)
	headers = foldAddresses(headers)
	if opts.InjectCcBcc {
		headers = injectCcBcc(headers)
	}

	body = reflowQuoted(body)
```

- [ ] **Step 6: Run all tests**

```bash
make check
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/compose/reflow.go internal/compose/reflow_test.go internal/compose/prepare.go
git commit -m "Add quoted text reflow to compose-prep pipeline

Join consecutive quoted lines at the same depth into paragraphs
and re-wrap at 72 columns. Blank quoted lines and decorative lines
are preserved. Uses go-runewidth for display-width measurement."
```

---

### Task 7: Integration Tests

**Files:**
- Create: `internal/compose/prepare_test.go`

- [ ] **Step 1: Write integration tests in `internal/compose/prepare_test.go`**

```go
package compose

import (
	"strings"
	"testing"
)

func TestPrepare(t *testing.T) {
	tests := []struct {
		name  string
		input string
		opts  Options
		want  string
	}{
		{
			name: "new compose with empty To",
			input: join(
				"From: geoff@907.life",
				"To:",
				"Subject:",
				"",
				"",
			),
			opts: Options{InjectCcBcc: true},
			want: join(
				"From: geoff@907.life",
				"To:",
				"Cc:",
				"Bcc:",
				"Subject:",
				"",
				"",
			),
		},
		{
			name: "reply with quoted text",
			input: join(
				"From: geoff@907.life",
				"To: Alice <alice@example.com>",
				"Subject: Re: Hello",
				"",
				"",
				"> This is a quoted line that was wrapped oddly by the",
				"> original sender's client and should be",
				"> reflowed.",
			),
			opts: Options{InjectCcBcc: true},
			want: join(
				"From: geoff@907.life",
				`To: "Alice" <alice@example.com>`,
				"Cc:",
				"Bcc:",
				"Subject: Re: Hello",
				"",
				"",
				"> This is a quoted line that was wrapped oddly by the original",
				"> sender's client and should be reflowed.",
			),
		},
		{
			name: "forward with empty To and quoted text",
			input: join(
				"From: geoff@907.life",
				"To:",
				"Subject: Fwd: News",
				"",
				"",
				"> Forwarded content.",
			),
			opts: Options{InjectCcBcc: true},
			want: join(
				"From: geoff@907.life",
				"To:",
				"Cc:",
				"Bcc:",
				"Subject: Fwd: News",
				"",
				"",
				"> Forwarded content.",
			),
		},
		{
			name: "multi-recipient folding",
			input: join(
				"From: geoff@907.life",
				"To: Alice <alice@example.com>, Bob <bob@example.com>, Charlie <charlie@example.com>",
				"Subject: Group",
				"",
				"Hello everyone.",
			),
			opts: Options{InjectCcBcc: true},
			want: join(
				"From: geoff@907.life",
				`To: "Alice" <alice@example.com>, "Bob" <bob@example.com>,`,
				`    "Charlie" <charlie@example.com>`,
				"Cc:",
				"Bcc:",
				"Subject: Group",
				"",
				"Hello everyone.",
			),
		},
		{
			name: "folded continuation header unfolded first",
			input: join(
				"From: geoff@907.life",
				"To: alice@example.com,",
				" bob@example.com",
				"Subject: Hi",
				"",
				"Body.",
			),
			opts: Options{InjectCcBcc: true},
			want: join(
				"From: geoff@907.life",
				"To: alice@example.com, bob@example.com",
				"Cc:",
				"Bcc:",
				"Subject: Hi",
				"",
				"Body.",
			),
		},
		{
			name: "no-cc-bcc flag",
			input: join(
				"From: geoff@907.life",
				"To: alice@dom",
				"Subject: Hi",
				"",
				"Body.",
			),
			opts: Options{InjectCcBcc: false},
			want: join(
				"From: geoff@907.life",
				"To: alice@dom",
				"Subject: Hi",
				"",
				"Body.",
			),
		},
		{
			name:  "malformed input no blank line",
			input: "From: alice@dom\nTo: bob@dom\nBody text\n",
			opts:  Options{InjectCcBcc: true},
			want:  "From: alice@dom\nTo: bob@dom\nBody text\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(Prepare([]byte(tt.input), tt.opts))
			if got != tt.want {
				t.Errorf("Prepare() mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, tt.want)
			}
		})
	}
}

// join builds a newline-terminated string from lines.
func join(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
```

**Note:** The expected values for `net/mail` formatting (e.g., `"Alice"` vs `Alice`) need to match what `net/mail.Address.String()` actually produces. After running the tests, adjust any mismatched expected values. The integration tests verify the full pipeline produces correct output — exact quoting style depends on `net/mail`.

- [ ] **Step 2: Run integration tests**

```bash
go test ./internal/compose/ -run TestPrepare -v
```

Expected: PASS. If any expected values need adjusting for `net/mail` formatting, fix them now.

- [ ] **Step 3: Run full test suite**

```bash
make check
```

Expected: PASS.

- [ ] **Step 4: Manual verification with compose-prep binary**

```bash
make build
printf "From: geoff@907.life\nTo: <alice@example.com>,\n bob@example.com\nSubject: Test\n\n> This is quoted text that was\n> wrapped oddly.\n" | ./compose-prep
```

Expected output (approximately):
```
From: geoff@907.life
To: alice@example.com, bob@example.com
Cc:
Bcc:
Subject: Test

> This is quoted text that was wrapped oddly.
```

- [ ] **Step 5: Commit**

```bash
git add internal/compose/prepare_test.go
git commit -m "Add integration tests for compose-prep pipeline

Full-buffer tests covering new compose, reply, forward,
multi-recipient folding, continuation unfolding, no-cc-bcc flag,
and malformed input pass-through."
```

---

### Task 8: Lua Integration

**Files:**
- Modify: `.config/nvim-mail/init.lua`
- Modify: `~/.dotfiles/beautiful-aerc/.config/nvim-mail/init.lua`

- [ ] **Step 1: Install the binary**

```bash
cd ~/Projects/beautiful-aerc
make install
```

Verify `compose-prep` is in `~/.local/bin/`:

```bash
which compose-prep
```

Expected: `/home/glw907/.local/bin/compose-prep`

- [ ] **Step 2: Identify the Lua code to replace in the project repo**

In `.config/nvim-mail/init.lua`, locate these sections:

1. **Helper functions** `get_quote_prefix`, `normalize_prefix`, `wrap_text`, `reflow_quoted` — these are defined before the VimEnter autocmd. Find them by searching for `local function get_quote_prefix`.

2. **VimEnter pipeline** — inside the second `VimEnter` autocmd callback, find the block from `local lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)` through the line just before `-- Insert blank lines after the header block` (or similar separator comment). This includes:
   - Unfold loop
   - Bracket stripping
   - Address re-folding
   - Cc/Bcc injection
   - Reflow call

- [ ] **Step 3: Delete helper functions from the project repo init.lua**

Delete the four functions: `get_quote_prefix`, `normalize_prefix`, `wrap_text`, and `reflow_quoted` (approximately lines 103–200 — verify exact range by reading the file). These are replaced by the Go binary.

- [ ] **Step 4: Replace the VimEnter pipeline in the project repo init.lua**

Inside the second VimEnter autocmd callback, replace the entire pipeline block (from reading buffer lines through the reflow call) with:

```lua
      -- Normalize headers and reflow quoted text via compose-prep.
      -- If compose-prep is not installed or fails, the buffer is left
      -- unchanged — usable but not pretty.
      local raw_lines = vim.api.nvim_buf_get_lines(0, 0, -1, false)
      local result = vim.fn.systemlist("compose-prep", raw_lines)
      if vim.v.shell_error == 0 and #result > 0 then
        vim.api.nvim_buf_set_lines(0, 0, -1, false, result)
      end
```

The code AFTER this (blank line insertion, extmark separators, cursor positioning, `startinsert`) must remain unchanged.

**Important:** The variable names used after the pipeline may reference `result` or `header_end`. After the replacement:
- `result` is now the output of `systemlist` — it is the full buffer (headers + blank line + body), not just headers.
- If subsequent code references `header_end`, it needs to be recalculated by scanning `result` for the first blank line. Add this after the `systemlist` block:

```lua
      -- Find header/body boundary for separator placement
      local header_end = 1
      for i, line in ipairs(result) do
        if line == "" then
          header_end = i
          break
        end
      end
```

Verify the existing separator and cursor positioning code works with this new `header_end` value.

- [ ] **Step 5: Apply the same changes to the dotfiles copy**

Apply the identical deletions and replacements to `~/.dotfiles/beautiful-aerc/.config/nvim-mail/init.lua`. The line numbers will be offset (~15 lines) due to extra comments at the top of the dotfiles copy.

- [ ] **Step 6: Test new compose**

Restart aerc. Press `C` to compose a new message. Verify:
- Headers are normalized (From, To, Cc, Bcc, Subject visible)
- Empty Cc: and Bcc: headers are present
- Cursor is on the To: line after "To: " in insert mode
- Separator extmarks render above and below headers

- [ ] **Step 7: Test reply**

In aerc, press `r` to reply to a message. Verify:
- Quoted text is reflowed at 72 columns
- Bare `<email>` brackets are stripped from headers
- Long recipient lists are folded
- Cursor is in the body between separator and quoted text

- [ ] **Step 8: Test graceful degradation**

Temporarily rename the binary and compose a new message:

```bash
mv ~/.local/bin/compose-prep ~/.local/bin/compose-prep.bak
```

Open a compose in aerc. Verify the editor opens with the raw buffer (no normalization, but no crash). Restore:

```bash
mv ~/.local/bin/compose-prep.bak ~/.local/bin/compose-prep
```

- [ ] **Step 9: Commit**

```bash
cd ~/Projects/beautiful-aerc
git add .config/nvim-mail/init.lua
git commit -m "Replace Lua buffer prep with compose-prep binary

Delete ~150 lines of regex-based RFC 2822 processing from init.lua.
The VimEnter autocmd now calls compose-prep via systemlist for
header normalization and quoted text reflow. Falls back to raw
buffer if the binary is not installed."
```

---

### Task 9: Documentation

**Files:**
- Modify: `CLAUDE.md`
- Modify: `README.md`

- [ ] **Step 1: Update CLAUDE.md**

In `CLAUDE.md`, find the `## Project Structure` section and add `compose-prep` to the tree:

```
cmd/compose-prep/      CLI wiring: compose buffer normalizer (cobra)
```

Add it alphabetically after `cmd/pick-link/` or in the appropriate position.

In the `## Build` section, `compose-prep` is already covered by `make build` / `make install` (builds all binaries). No change needed.

Add a new section after `## tidytext`:

```markdown
## compose-prep

Stdin/stdout buffer normalizer for the nvim-mail compose editor.
Reads an aerc compose buffer from stdin, normalizes headers and
reflows quoted text, writes the result to stdout. Called by
nvim-mail's VimEnter autocmd via `vim.fn.systemlist`.

### Pipeline

1. **Unfold** RFC 2822 continuation lines (space/tab prefix)
2. **Strip brackets** from bare `<email>` addresses (uses `net/mail`)
3. **Fold** To/Cc/Bcc at 72-column recipient boundaries
4. **Inject** empty Cc:/Bcc: headers when absent
5. **Reflow** quoted text paragraphs at 72 columns

### Flags

    --no-cc-bcc    Suppress empty Cc/Bcc header injection
    --debug        Write diagnostic messages to stderr

### Error Behavior

On any processing error, the original input is passed through
unchanged. Exit code is always 0. The compose window always opens.
```

- [ ] **Step 2: Update README.md**

In the `### nvim-mail` section, update the feature list. Find the bullet about Telescope contact picker or address header reformatting and add/update:

```markdown
- Compose buffer normalization via `compose-prep` — RFC 2822 header unfolding, bare bracket stripping, address folding at 72 columns, Cc/Bcc header injection, and quoted text reflow. Falls back gracefully if the binary is not installed.
```

In the Optional prerequisites section, add:

```markdown
- **compose-prep** — compose buffer normalizer (installed via `make install` from this repo)
```

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md README.md
git commit -m "Document compose-prep in CLAUDE.md and README.md"
```

---

### Summary of Files Created/Modified

**Created:**
- `cmd/compose-prep/main.go`
- `cmd/compose-prep/root.go`
- `internal/compose/prepare.go`
- `internal/compose/prepare_test.go`
- `internal/compose/unfold.go`
- `internal/compose/unfold_test.go`
- `internal/compose/bracket.go`
- `internal/compose/bracket_test.go`
- `internal/compose/fold.go`
- `internal/compose/fold_test.go`
- `internal/compose/inject.go`
- `internal/compose/inject_test.go`
- `internal/compose/reflow.go`
- `internal/compose/reflow_test.go`

**Modified:**
- `Makefile`
- `go.mod` / `go.sum`
- `.config/nvim-mail/init.lua`
- `~/.dotfiles/beautiful-aerc/.config/nvim-mail/init.lua`
- `CLAUDE.md`
- `README.md`

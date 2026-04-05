# Save & Fix Corpus Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `save` subcommand and keybinding to flag emails with rendering issues, plus project-level Claude infrastructure for batch fixing.

**Architecture:** New `save` cobra subcommand reads stdin, sniffs HTML vs plain text, writes to `corpus/` with timestamped filename. Claude skill `fix-corpus` drives batch triage/fix workflow. Project `.claude/` directory holds memories, skills, and docs so the project is self-contained.

**Tech Stack:** Go 1.26, cobra, existing palette resolution pattern

---

### Task 1: `save` subcommand -- corpus directory resolution

**Files:**
- Create: `internal/corpus/corpus.go`
- Create: `internal/corpus/corpus_test.go`

This task adds the corpus directory finder, following the same resolution pattern as `palette.FindPath()` in `internal/palette/palette.go`.

- [ ] **Step 1: Write failing tests for FindDir**

```go
package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (envVal string, binHint string)
		wantErr bool
	}{
		{
			"env override",
			func(t *testing.T) (string, string) {
				dir := t.TempDir()
				corpus := filepath.Join(dir, "corpus")
				os.MkdirAll(corpus, 0755)
				return dir, ""
			},
			false,
		},
		{
			"relative to binary hint",
			func(t *testing.T) (string, string) {
				dir := t.TempDir()
				// Simulate: project/.config/aerc -> ../../corpus
				aercDir := filepath.Join(dir, ".config", "aerc")
				os.MkdirAll(aercDir, 0755)
				corpus := filepath.Join(dir, "corpus")
				os.MkdirAll(corpus, 0755)
				return "", aercDir
			},
			false,
		},
		{
			"creates corpus dir if missing",
			func(t *testing.T) (string, string) {
				dir := t.TempDir()
				aercDir := filepath.Join(dir, ".config", "aerc")
				os.MkdirAll(aercDir, 0755)
				// corpus/ does not exist yet
				return "", aercDir
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVal, binHint := tt.setup(t)
			if envVal != "" {
				t.Setenv("AERC_CONFIG", envVal)
			} else {
				t.Setenv("AERC_CONFIG", "")
			}
			dir, err := FindDir(binHint)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, err := os.Stat(dir); err != nil {
				t.Errorf("corpus dir does not exist: %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/corpus/ -v`
Expected: FAIL (package does not exist)

- [ ] **Step 3: Implement FindDir**

```go
package corpus

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindDir locates or creates the corpus directory. Resolution order:
// 1. $AERC_CONFIG/../../corpus/ (env override)
// 2. configHint/../../corpus/ (caller-supplied aerc config path)
// 3. ~/.config/aerc/../../corpus/ (default)
// Creates the directory if it does not exist.
func FindDir(configHint string) (string, error) {
	var candidates []string

	if aercConfig := os.Getenv("AERC_CONFIG"); aercConfig != "" {
		candidates = append(candidates, filepath.Join(aercConfig, "..", "..", "corpus"))
	}

	if configHint != "" {
		candidates = append(candidates, filepath.Join(configHint, "..", "..", "corpus"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "aerc", "..", "..", "corpus"))
	}

	// Check for existing corpus dir first.
	for _, c := range candidates {
		c = filepath.Clean(c)
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}

	// Create at first candidate location.
	if len(candidates) > 0 {
		dir := filepath.Clean(candidates[0])
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("creating corpus directory %s: %w", dir, err)
		}
		return dir, nil
	}

	return "", fmt.Errorf("cannot determine corpus directory")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/corpus/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/corpus/corpus.go internal/corpus/corpus_test.go
git commit -m "Add corpus directory resolution"
```

---

### Task 2: `save` subcommand -- content sniffing and file writing

**Files:**
- Modify: `internal/corpus/corpus.go`
- Modify: `internal/corpus/corpus_test.go`

- [ ] **Step 1: Write failing tests for IsHTML and Save**

Add to `internal/corpus/corpus_test.go`:

```go
func TestIsHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"html tag", "<html><body>hello</body></html>", true},
		{"doctype", "<!DOCTYPE html><html>", true},
		{"head tag", "<head><meta charset='utf-8'></head>", true},
		{"body tag", "<body>content</body>", true},
		{"table tag", "<table><tr><td>cell</td></tr></table>", true},
		{"case insensitive", "<HTML><BODY>hello</BODY></HTML>", true},
		{"plain text", "Hello, this is a plain email.", false},
		{"markdown", "# Heading\n\nSome **bold** text.", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsHTML([]byte(tt.input))
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSave(t *testing.T) {
	t.Run("html content", func(t *testing.T) {
		dir := t.TempDir()
		content := []byte("<html><body>test</body></html>")
		path, err := Save(dir, content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Ext(path) != ".html" {
			t.Errorf("expected .html extension, got %s", filepath.Ext(path))
		}
		got, _ := os.ReadFile(path)
		if string(got) != string(content) {
			t.Errorf("content mismatch")
		}
	})

	t.Run("plain text content", func(t *testing.T) {
		dir := t.TempDir()
		content := []byte("Hello, plain text email.")
		path, err := Save(dir, content)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filepath.Ext(path) != ".txt" {
			t.Errorf("expected .txt extension, got %s", filepath.Ext(path))
		}
	})

	t.Run("empty input", func(t *testing.T) {
		dir := t.TempDir()
		_, err := Save(dir, []byte{})
		if err == nil {
			t.Error("expected error for empty input")
		}
	})

	t.Run("collision avoidance", func(t *testing.T) {
		dir := t.TempDir()
		content := []byte("plain text")
		p1, _ := Save(dir, content)
		p2, _ := Save(dir, content)
		if p1 == p2 {
			t.Error("expected different paths for same-second saves")
		}
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/corpus/ -v -run TestIsHTML`
Expected: FAIL (undefined: IsHTML)

- [ ] **Step 3: Implement IsHTML and Save**

Add to `internal/corpus/corpus.go`:

```go
import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// htmlMarkers are case-insensitive strings that indicate HTML content.
var htmlMarkers = []string{"<html", "<head", "<body", "<!doctype", "<table"}

// IsHTML returns true if data looks like HTML content.
func IsHTML(data []byte) bool {
	lower := bytes.ToLower(data[:min(len(data), 1024)])
	for _, marker := range htmlMarkers {
		if bytes.Contains(lower, []byte(marker)) {
			return true
		}
	}
	return false
}

// Save writes data to the corpus directory with a timestamped filename.
// Returns the full path of the saved file.
func Save(dir string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("no input")
	}

	ext := ".txt"
	if IsHTML(data) {
		ext = ".html"
	}

	stamp := time.Now().Format("20060102-150405")
	name := stamp + ext
	path := filepath.Join(dir, name)

	// Handle collision: append -2, -3, etc.
	if _, err := os.Stat(path); err == nil {
		for i := 2; ; i++ {
			name = fmt.Sprintf("%s-%d%s", stamp, i, ext)
			path = filepath.Join(dir, name)
			if _, err := os.Stat(path); err != nil {
				break
			}
		}
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("writing corpus file %s: %w", name, err)
	}
	return path, nil
}
```

Note: Update the import block at the top of `corpus.go` to include `bytes`, `strings`, `time`, and remove any unused imports.

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/corpus/ -v`
Expected: PASS (all tests)

- [ ] **Step 5: Run full check**

Run: `make check`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/corpus/corpus.go internal/corpus/corpus_test.go
git commit -m "Add content sniffing and file saving to corpus package"
```

---

### Task 3: `save` cobra subcommand

**Files:**
- Create: `cmd/beautiful-aerc/save.go`
- Modify: `cmd/beautiful-aerc/root.go`

- [ ] **Step 1: Create save subcommand**

Create `cmd/beautiful-aerc/save.go`:

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/glw907/beautiful-aerc/internal/corpus"
	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current email part to corpus for later analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			binPath, _ := os.Executable()
			configHint := ""
			if binPath != "" {
				resolved, err := filepath.EvalSymlinks(binPath)
				if err == nil {
					binPath = resolved
				}
				binDir := filepath.Dir(binPath)
				configHint = filepath.Join(binDir, "..", "..", ".config", "aerc")
			}

			dir, err := corpus.FindDir(configHint)
			if err != nil {
				return err
			}

			path, err := corpus.Save(dir, data)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "saved %s\n", filepath.Base(path))
			return nil
		},
	}
	return cmd
}
```

- [ ] **Step 2: Register in root command**

Modify `cmd/beautiful-aerc/root.go` -- add `cmd.AddCommand(newSaveCmd())`:

```go
func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "beautiful-aerc",
		Short:        "Themeable filters for the aerc email client",
		SilenceUsage: true,
	}
	cmd.AddCommand(newHeadersCmd())
	cmd.AddCommand(newHTMLCmd())
	cmd.AddCommand(newPlainCmd())
	cmd.AddCommand(newPickLinkCmd())
	cmd.AddCommand(newSaveCmd())
	return cmd
}
```

- [ ] **Step 3: Run check**

Run: `make check`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/beautiful-aerc/save.go cmd/beautiful-aerc/root.go
git commit -m "Add save subcommand for flagging emails to corpus"
```

---

### Task 4: E2e test for save subcommand

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Add save e2e test**

Add to `e2e/e2e_test.go`:

```go
func TestSaveHTMLFixture(t *testing.T) {
	corpusDir := filepath.Join(t.TempDir(), "corpus")

	input, err := os.ReadFile("testdata/simple.html")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	cmd := exec.Command(binary, "save")
	cmd.Stdin = bytes.NewReader(input)
	cmd.Env = append(os.Environ(),
		"AERC_CONFIG="+corpusDir+"/.config/aerc",
	)
	// Create the fake config path so FindDir resolves ../../corpus
	aercDir := filepath.Join(corpusDir, ".config", "aerc")
	os.MkdirAll(aercDir, 0755)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running save: %v\noutput: %s", err, out)
	}

	// Verify a .html file was created in the corpus dir
	matches, _ := filepath.Glob(filepath.Join(corpusDir, "corpus", "*.html"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 html file in corpus, got %d", len(matches))
	}

	saved, _ := os.ReadFile(matches[0])
	if !bytes.Equal(saved, input) {
		t.Error("saved content does not match input")
	}
}

func TestSavePlainText(t *testing.T) {
	corpusDir := filepath.Join(t.TempDir(), "corpus")
	aercDir := filepath.Join(corpusDir, ".config", "aerc")
	os.MkdirAll(aercDir, 0755)

	input := []byte("Hello, this is a plain text email.\n")

	cmd := exec.Command(binary, "save")
	cmd.Stdin = bytes.NewReader(input)
	cmd.Env = append(os.Environ(),
		"AERC_CONFIG="+corpusDir+"/.config/aerc",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running save: %v\noutput: %s", err, out)
	}

	matches, _ := filepath.Glob(filepath.Join(corpusDir, "corpus", "*.txt"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 txt file in corpus, got %d", len(matches))
	}
}
```

- [ ] **Step 2: Run tests**

Run: `make check`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add e2e/e2e_test.go
git commit -m "Add e2e tests for save subcommand"
```

---

### Task 5: Keybinding and gitignore

**Files:**
- Modify: `.config/aerc/binds.conf`
- Modify: `.gitignore`

- [ ] **Step 1: Add keybinding**

Add to the `[view]` section of `.config/aerc/binds.conf`, after the `| = :pipe<space>` line:

```ini
b = :pipe -m beautiful-aerc save<Enter>
```

- [ ] **Step 2: Add corpus to gitignore**

Add to `.gitignore`:

```
corpus/
```

- [ ] **Step 3: Commit**

```bash
git add .config/aerc/binds.conf .gitignore
git commit -m "Add save keybinding and gitignore corpus directory"
```

---

### Task 6: Project Claude infrastructure -- memories and docs

**Files:**
- Create: `.claude/memory/MEMORY.md`
- Create: `.claude/memory/pipeline_architecture.md`
- Create: `.claude/memory/problem_senders.md`
- Create: `.claude/memory/edge_case_workflow.md`
- Create: `.claude/memory/debug_methodology.md`
- Create: `.claude/docs/tmux-testing.md`

- [ ] **Step 1: Create pipeline architecture memory**

Create `.claude/memory/pipeline_architecture.md`:

```markdown
---
name: HTML filter pipeline architecture
description: Current Go pipeline stages for HTML-to-markdown email conversion, ordering, and what each stage does
type: project
---

The HTML email filter is a Go binary (`beautiful-aerc html`) that
converts HTML email to syntax-highlighted markdown for aerc's viewer.

**Pipeline stages (in order):**

1. `prepareHTML` -- strip Mozilla attributes, hidden elements
   (display:none divs), and zero-size tracking images
2. `runPandoc` -- HTML to markdown via pandoc with extensions disabled
   (-raw_html, -native_divs, -native_spans, -header_attributes,
   -bracketed_spans, -fenced_divs, -inline_code_attributes,
   -link_attributes) plus unwrap-tables.lua Lua filter
3. `html.UnescapeString` -- decode HTML entities
4. `cleanPandocArtifacts` -- trailing backslashes, escaped punctuation,
   consecutive bold markers, stray bold, superscript carets, nested
   headings, empty headings
5. `normalizeBoldMarkers` -- balance ** markers per paragraph, strip
   unpaired trailing markers
6. `normalizeLists` -- convert Unicode bullets to markdown items, strip
   excess indentation, compact loose lists
7. `normalizeWhitespace` -- NBSP, zero-width chars, blank line cleanup
8. `convertToFootnotes` -- reference-style links to numbered footnotes
9. `styleFootnotes` -- ANSI colors for footnote markers and ref section
10. `highlightMarkdown` -- ANSI colors for headings, bold, italic, rules

**Key files:**
- `internal/filter/html.go` -- pipeline stages 1, 3-7, 10
- `internal/filter/footnotes.go` -- stages 8-9
- `internal/palette/palette.go` -- color token loading
- `.config/aerc/filters/unwrap-tables.lua` -- pandoc Lua filter

**Why:** Understanding the stage ordering matters because fixes must
target the right stage. Regex cleanup after pandoc is intentional --
pandoc's markdown output has artifacts that can't be prevented by
pandoc flags alone.

**How to apply:** When debugging a rendering issue, trace the email
through each stage to find where the problem is introduced. Use
`corpus/` to save problem emails for batch fixing.
```

- [ ] **Step 2: Create problem senders memory**

Create `.claude/memory/problem_senders.md`:

```markdown
---
name: Problem sender patterns
description: Email sender types that stress the HTML filter pipeline -- useful for regression testing
type: project
---

Sender types that produce challenging HTML for the pipeline:

- **Marketing emails** -- layout tables, tracking images, nbsp padding,
  responsive duplicates in display:none divs
- **Bank of America** -- zero-size tracking pixels inline between URL
  fragments, unclosed `<strong>` tags producing consecutive bold
- **Apple receipts** -- complex nested tables, heavy inline styles
- **GitHub notifications** -- image-links, multi-line link text,
  tracking redirects with empty URLs
- **Google Calendar** -- image buttons for RSVP, empty-URL links
- **Thunderbird senders** -- `class="moz-*"` attributes pollute output
- **Yahoo mail** -- various DOCTYPE formats, non-standard structures
- **Microsoft Outlook** -- platform-specific markup, MSO conditionals
- **Newsletters** -- HTML with no paragraph breaks in text/plain part
  (reason HTML MIME part is preferred over text/plain)
- **Remind.com** -- bare domain URLs without https:// scheme
- **Callcentric** -- angle-bracket autolink URLs
- **ClouDNS** -- empty mailto: links
- **Spotify/dbrand** -- responsive HTML with duplicate content in
  display:none sections

**Why:** These patterns recur. New pipeline fixes should be tested
against this list to catch regressions.

**How to apply:** When making pipeline changes, mentally check whether
the change could affect any of these sender types. The `corpus/`
directory and `scripts/audit.sh` can test against real examples.
```

- [ ] **Step 3: Create edge case workflow memory**

Create `.claude/memory/edge_case_workflow.md`:

```markdown
---
name: Edge case fix workflow
description: When the user flags rendering issues, fix the filter and ship autonomously
type: feedback
---

When the user reports formatting issues (via corpus, pasted output,
or description), fix the beautiful-aerc filter code autonomously,
then run /ship when done. No need to ask for confirmation -- diagnose,
fix, and ship.

For batch fixes, use the /fix-corpus skill which previews each corpus
email via tmux, triages by pattern, fixes holistically, and ships.

**Why:** The user tests beautiful-aerc by reading real email and
spotting rendering problems. They want quick turnaround, not
discussion.

**How to apply:** On seeing reported issues, immediately start
diagnosing. Trace through the pipeline, fix the Go code, and ship.
```

- [ ] **Step 4: Create debug methodology memory**

Create `.claude/memory/debug_methodology.md`:

```markdown
---
name: Fix the cause not the symptoms
description: When debugging, identify the source of the problem before trying display-layer fixes
type: feedback
---

When something unwanted is showing up in output, trace it back to the
source (the tool/server producing it) rather than trying to suppress
it at every display layer.

**Why:** Wasted rounds trying to suppress issues at the display layer
when the fix was configuring the producer.

**How to apply:** Ask "what is producing this?" first. Configure the
producer, not the consumer. In the pipeline context: if pandoc
produces bad output, check if a pre-pandoc HTML cleanup can prevent
it rather than adding another post-pandoc regex.
```

- [ ] **Step 5: Create MEMORY.md index**

Create `.claude/memory/MEMORY.md`:

```markdown
# beautiful-aerc Project Memory

- [pipeline_architecture.md](pipeline_architecture.md) -- Go pipeline stages, ordering, key files
- [problem_senders.md](problem_senders.md) -- Sender types that stress the filter pipeline
- [edge_case_workflow.md](edge_case_workflow.md) -- Fix rendering issues and ship autonomously
- [debug_methodology.md](debug_methodology.md) -- Fix the cause, not the symptoms
```

- [ ] **Step 6: Create tmux testing doc**

Create `.claude/docs/tmux-testing.md`:

```markdown
# Filter Testing via tmux

Render emails through the filter and verify output without requiring
human visual inspection.

## Rendering a corpus file

```bash
# HTML email
cat corpus/20260404-143022.html | beautiful-aerc html

# Plain text email
cat corpus/20260404-143022.txt | beautiful-aerc plain
```

## Preview in tmux (simulates aerc viewer)

```bash
tmux kill-session -t test 2>/dev/null
tmux new-session -d -s test -x 80 -y 40

# Render and display
cat corpus/file.html \
  | AERC_COLUMNS=80 AERC_CONFIG=~/.config/aerc beautiful-aerc html \
  | tmux load-buffer - \
  && tmux paste-buffer -t test

# Capture for inspection
tmux capture-pane -t test -p

# Clean up
tmux kill-session -t test
```

## Batch audit

The `scripts/audit.sh` script samples HTML from the JMAP blob cache,
runs each through the filter, and writes rendered output to a
directory for review.

```bash
bash scripts/audit.sh -o audit-output/
```

## Comparing output

Strip ANSI codes for text comparison:

```bash
cat corpus/file.html \
  | AERC_COLUMNS=80 AERC_CONFIG=~/.config/aerc beautiful-aerc html \
  | sed 's/\x1b\[[0-9;]*m//g' > /tmp/rendered.txt
```
```

- [ ] **Step 7: Commit**

```bash
git add .claude/memory/ .claude/docs/
git commit -m "Add project-level Claude memories and docs"
```

---

### Task 7: Project Claude infrastructure -- skill and settings

**Files:**
- Create: `.claude/skills/fix-corpus`
- Create: `.claude/settings.json`

- [ ] **Step 1: Create fix-corpus skill**

Create `.claude/skills/fix-corpus`:

```markdown
---
name: fix-corpus
description: Batch review and fix emails with rendering issues saved in corpus/
---

# Fix Corpus

Batch triage and fix rendering issues in emails saved to `corpus/`.

## Prerequisites

- Emails saved via `beautiful-aerc save` (keybinding `b` in aerc viewer)
- At least a few corpus files to make batch processing worthwhile

## Workflow

### 1. Scan corpus

List all files in `corpus/`, grouped by type:

```bash
ls -la corpus/*.html corpus/*.txt 2>/dev/null
```

If empty, tell the user and stop.

### 2. Preview each email

For each corpus file, render it through the filter and show the output.
Use the tmux pattern from `.claude/docs/tmux-testing.md`:

```bash
tmux kill-session -t corpus-review 2>/dev/null
tmux new-session -d -s corpus-review -x 80 -y 40
```

For each file:

```bash
# Render through appropriate filter
cat corpus/<file> \
  | AERC_COLUMNS=80 AERC_CONFIG=~/.config/aerc beautiful-aerc html \
  | tmux load-buffer - \
  && tmux paste-buffer -t corpus-review
tmux capture-pane -t corpus-review -p
```

Use `beautiful-aerc html` for `.html` files and `beautiful-aerc plain`
for `.txt` files.

Show the rendered output to the user. If the issue is not obvious,
ask: "What's wrong with this one?"

### 3. Triage by pattern

After reviewing all emails, group issues by root cause. Look for
commonality before making fixes -- the holistic approach produces
better code than fixing one email at a time.

Present the grouped issues to the user for confirmation before
proceeding.

### 4. Fix pipeline code

Make changes to `internal/filter/` to address the identified patterns.
Add test cases for each fix in the corresponding `_test.go` files.

### 5. Verify

Re-render all corpus emails and confirm fixes. Also re-run the
full test suite to check for regressions:

```bash
make check
```

### 6. Quality gates

Run the standard quality passes:
- `make check` (vet + tests)
- `/go-review` (Go convention review)
- `/simplify` (code quality review)

### 7. Ship

```bash
/ship
```

This commits, pushes, and installs the updated binary.

### 8. Cleanup

Ask the user:
- Should any fixed emails become e2e test fixtures? If yes, copy to
  `e2e/testdata/` and generate golden files with
  `go test ./e2e/ -update-golden`.
- Remove fixed emails from `corpus/`?

Clean up the tmux session:

```bash
tmux kill-session -t corpus-review 2>/dev/null
```
```

- [ ] **Step 2: Create project settings.json**

Create `.claude/settings.json`:

```json
{}
```

- [ ] **Step 3: Commit**

```bash
git add .claude/skills/fix-corpus .claude/settings.json
git commit -m "Add fix-corpus skill and project settings"
```

---

### Task 8: Update CLAUDE.md and gitignore .claude artifacts

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update CLAUDE.md**

Add the following sections to the end of the existing `CLAUDE.md`:

```markdown

## Corpus

`corpus/` holds raw email parts (HTML or plain text) flagged for
rendering issues. Save emails from aerc with `b` in the viewer.
The `/fix-corpus` skill batch-processes accumulated corpus emails.

## Filter Testing

See `.claude/docs/tmux-testing.md` for patterns to render emails
through the filter and verify output via tmux.
```

- [ ] **Step 2: Run make check**

Run: `make check`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "Add corpus and testing references to CLAUDE.md"
```

---

### Task 9: Global cleanup

**Files:**
- Modify: `~/.claude/projects/-home-glw907/memory/MEMORY.md`
- Remove: `~/.claude/projects/-home-glw907/memory/project_aerc_html_pipeline.md`
- Modify: `~/.claude/CLAUDE.md` (symlinked from dotfiles)

- [ ] **Step 1: Remove stale global pipeline memory**

Delete `~/.claude/projects/-home-glw907/memory/project_aerc_html_pipeline.md`
and remove its line from `~/.claude/projects/-home-glw907/memory/MEMORY.md`.

- [ ] **Step 2: Remove aerc section from global CLAUDE.md**

Remove the `## aerc (Email)` section from `~/.claude/CLAUDE.md` (the
symlinked file at `~/.dotfiles/claude/.claude/CLAUDE.md`). This section
is approximately 6 lines starting with `## aerc (Email)` and ending
before the next `##` heading. The content now lives in the project
CLAUDE.md and project memories.

- [ ] **Step 3: Commit dotfiles changes**

```bash
cd ~/.dotfiles
git add claude/.claude/CLAUDE.md
git commit -m "Remove aerc section from global CLAUDE.md (moved to project)"
cd ~/Projects/beautiful-aerc
```

- [ ] **Step 4: Verify project context is self-contained**

Start a fresh thought: does `CLAUDE.md` + `.claude/memory/` + `.claude/docs/` + `.claude/skills/` provide everything needed to work on this project without global context? Check that:
- Pipeline architecture is documented
- Problem senders are listed
- Fix workflow is described
- Testing patterns are documented
- Build commands are in CLAUDE.md
- Go conventions pointer is in CLAUDE.md

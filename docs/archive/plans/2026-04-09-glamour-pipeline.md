# Glamour Pipeline Migration — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the pandoc-based HTML rendering pipeline with Go-native libraries (html-to-markdown + Glamour), eliminate pick-link, and add table rendering support.

**Architecture:** HTML emails are pre-cleaned, converted to markdown via html-to-markdown (with a layout table plugin), then rendered to styled ANSI output via Glamour with a custom style derived from the theme TOML. A new `to-html` subcommand uses goldmark for the sending direction. A new `markdown` subcommand outputs clean markdown for reply templates.

**Tech Stack:** `html-to-markdown/v2`, `glamour`, `goldmark` (transitive via glamour, used directly for to-html)

**Spec:** `docs/superpowers/specs/2026-04-09-glamour-pipeline-design.md`

---

## File Structure

### New files
- `internal/filter/html.go` — rewritten: new pipeline using html-to-markdown + Glamour
- `internal/filter/html_test.go` — rewritten: unit tests for new pipeline
- `internal/filter/convert.go` — html-to-markdown converter setup + layout table plugin
- `internal/filter/convert_test.go` — unit tests for converter and layout table detection
- `internal/theme/glamour.go` — theme-to-Glamour style mapping
- `internal/theme/glamour_test.go` — tests for style mapping
- `cmd/mailrender/markdown.go` — new `markdown` subcommand
- `cmd/mailrender/tohtml.go` — new `to-html` subcommand

### Modified files
- `cmd/mailrender/root.go` — register new subcommands
- `cmd/mailrender/html.go` — updated to use new `filter.HTML` signature
- `internal/filter/plain.go` — updated for new `filter.HTML` signature
- `Makefile` — remove pick-link
- `go.mod` / `go.sum` — new dependencies

### Deleted files
- `internal/filter/footnotes.go`
- `internal/filter/footnotes_test.go`
- `cmd/pick-link/root.go`
- `cmd/pick-link/main.go`
- `internal/picker/picker.go`

### Unchanged files
- `internal/filter/headers.go` — no changes
- `internal/theme/theme.go` — no changes (new code in separate file)
- `internal/theme/styleset.go` — no changes
- `cmd/mailrender/headers.go` — no changes
- `cmd/mailrender/themes.go` — no changes

---

## Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add html-to-markdown and glamour**

```bash
cd /home/glw907/Projects/beautiful-aerc
go get github.com/JohannesKaufmann/html-to-markdown/v2@latest
go get github.com/charmbracelet/glamour@latest
```

- [ ] **Step 2: Verify dependencies resolve**

```bash
go mod tidy
```

Expected: clean exit, no errors.

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "Add html-to-markdown and glamour dependencies

These replace pandoc (external binary) for HTML-to-markdown conversion
and custom regex-based markdown highlighting for ANSI rendering.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 2: Theme-to-Glamour Style Mapping

**Files:**
- Create: `internal/theme/glamour.go`
- Test: `internal/theme/glamour_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/theme/glamour_test.go
package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/glamour/ansi"
)

func writeTestTheme(t *testing.T, dir string) string {
	t.Helper()
	content := `name = "test"

[colors]
bg_base = "#2e3440"
bg_elevated = "#3b4252"
bg_selection = "#394353"
bg_border = "#49576b"
fg_base = "#d8dee9"
fg_bright = "#e5e9f0"
fg_brightest = "#eceff4"
fg_dim = "#616e88"
accent_primary = "#81a1c1"
accent_secondary = "#88c0d0"
accent_tertiary = "#8fbcbb"
color_error = "#bf616a"
color_warning = "#d08770"
color_success = "#a3be8c"
color_info = "#ebcb8b"
color_special = "#b48ead"

[tokens]
heading = { color = "color_success", bold = true }
bold = { bold = true }
italic = { italic = true }
link_text = { color = "accent_secondary" }
rule = { color = "fg_dim" }
hdr_key = { color = "accent_primary", bold = true }
hdr_value = { color = "fg_base" }
hdr_dim = { color = "fg_dim" }
`
	path := filepath.Join(dir, "test.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGlamourStyle(t *testing.T) {
	dir := t.TempDir()
	path := writeTestTheme(t, dir)
	th, err := Load(path)
	if err != nil {
		t.Fatalf("loading theme: %v", err)
	}

	style := th.GlamourStyle()

	// Document margins should be zero.
	if style.Document.Margin == nil || *style.Document.Margin != 0 {
		t.Error("document margin should be 0")
	}
	if style.Document.Indent == nil || *style.Document.Indent != 0 {
		t.Error("document indent should be 0")
	}

	// Heading should use color_success hex and bold.
	if style.H1.Color == nil || *style.H1.Color != "#a3be8c" {
		t.Errorf("H1 color = %v, want #a3be8c", style.H1.Color)
	}
	if style.H1.Bold == nil || !*style.H1.Bold {
		t.Error("H1 should be bold")
	}

	// Strong should be bold.
	if style.Strong.Bold == nil || !*style.Strong.Bold {
		t.Error("Strong should be bold")
	}

	// Emph should be italic.
	if style.Emph.Italic == nil || !*style.Emph.Italic {
		t.Error("Emph should be italic")
	}

	// Link text should use accent_secondary hex.
	if style.LinkText.Color == nil || *style.LinkText.Color != "#88c0d0" {
		t.Errorf("LinkText color = %v, want #88c0d0", style.LinkText.Color)
	}
}

func TestGlamourStyleMissingToken(t *testing.T) {
	dir := t.TempDir()
	// Theme with no heading token — should still produce a valid style.
	content := `name = "minimal"

[colors]
bg_base = "#2e3440"
bg_elevated = "#3b4252"
bg_selection = "#394353"
bg_border = "#49576b"
fg_base = "#d8dee9"
fg_bright = "#e5e9f0"
fg_brightest = "#eceff4"
fg_dim = "#616e88"
accent_primary = "#81a1c1"
accent_secondary = "#88c0d0"
accent_tertiary = "#8fbcbb"
color_error = "#bf616a"
color_warning = "#d08770"
color_success = "#a3be8c"
color_info = "#ebcb8b"
color_special = "#b48ead"
`
	path := filepath.Join(dir, "minimal.toml")
	os.WriteFile(path, []byte(content), 0644)
	th, err := Load(path)
	if err != nil {
		t.Fatalf("loading theme: %v", err)
	}

	style := th.GlamourStyle()

	// Should not panic; heading fields should be nil (Glamour defaults).
	if style.Document.Margin == nil || *style.Document.Margin != 0 {
		t.Error("document margin should be 0 even with minimal theme")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/theme/ -run TestGlamourStyle -v
```

Expected: FAIL — `GlamourStyle` method not defined.

- [ ] **Step 3: Write the implementation**

```go
// internal/theme/glamour.go
package theme

import (
	"github.com/charmbracelet/glamour/ansi"
)

func ptr[T any](v T) *T { return &v }

// tokenDef retrieves the raw token definition for a given token name.
// Returns the color hex and style booleans. If the token is not defined,
// all values are zero/empty.
func (t *Theme) tokenDef(name string) (hex string, bold, italic, underline bool) {
	// We need to look up the token definition to get the color slot name,
	// then resolve the slot to a hex color. The Theme struct only stores
	// the resolved SGR string, but for Glamour we need the original hex.
	// We need access to the original tokenDefinition and colors map.
	// Since Theme stores colors, we can reverse-engineer from the SGR string,
	// but it's cleaner to store what we need.
	//
	// This requires a small change: store the original themeFile data.
	// See the tokenDefs and colors fields added below.
	return "", false, false, false
}

// GlamourStyle builds a Glamour ansi.StyleConfig from the theme's tokens
// and color slots. Tokens map to Glamour style elements:
//
//   heading   → H1-H6
//   bold      → Strong
//   italic    → Emph
//   link_text → LinkText
//   rule      → HorizontalRule
func (t *Theme) GlamourStyle() ansi.StyleConfig {
	style := ansi.StyleConfig{
		Document: ansi.StyleBlock{
			Margin: ptr(uint(0)),
			Indent: ptr(uint(0)),
		},
	}

	// Heading: apply to all heading levels.
	if hdr := t.glamourPrimitive("heading"); hdr != nil {
		block := ansi.StyleBlock{StylePrimitive: *hdr}
		style.H1 = block
		style.H2 = block
		style.H3 = block
		style.H4 = block
		style.H5 = block
		style.H6 = block
	}

	// Bold → Strong
	if s := t.glamourPrimitive("bold"); s != nil {
		style.Strong = *s
	}

	// Italic → Emph
	if s := t.glamourPrimitive("italic"); s != nil {
		style.Emph = *s
	}

	// Link text
	if s := t.glamourPrimitive("link_text"); s != nil {
		style.LinkText = *s
	}

	// Horizontal rule
	if s := t.glamourPrimitive("rule"); s != nil {
		style.HorizontalRule = *s
	}

	return style
}

// glamourPrimitive converts a theme token to a Glamour StylePrimitive.
// Returns nil if the token is not defined.
func (t *Theme) glamourPrimitive(tokenName string) *ansi.StylePrimitive {
	def, ok := t.tokenDefs[tokenName]
	if !ok {
		return nil
	}

	p := &ansi.StylePrimitive{}

	if def.Color != "" {
		if hex, ok := t.colors[def.Color]; ok {
			p.Color = ptr(hex)
		}
	}
	if def.Bold {
		p.Bold = ptr(true)
	}
	if def.Italic {
		p.Italic = ptr(true)
	}
	if def.Underline {
		p.Underline = ptr(true)
	}

	return p
}
```

This requires `tokenDefs` on the `Theme` struct. Modify `internal/theme/theme.go`:

Add `tokenDefs` field to the `Theme` struct:

```go
// Theme holds parsed color slots and resolved ANSI tokens.
type Theme struct {
	Name      string
	colors    map[string]string            // "accent_primary" → "#81a1c1"
	tokens    map[string]string            // "hdr_key" → "38;2;129;161;193;1"
	tokenDefs map[string]tokenDefinition   // original token definitions for Glamour
}
```

In the `Load` function, after resolving tokens, store the definitions:

```go
	return &Theme{
		Name:      f.Name,
		colors:    f.Colors,
		tokens:    resolved,
		tokenDefs: f.Tokens,
	}, nil
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/theme/ -run TestGlamourStyle -v
```

Expected: PASS for both `TestGlamourStyle` and `TestGlamourStyleMissingToken`.

- [ ] **Step 5: Run full test suite**

```bash
go vet ./internal/theme/ && go test ./internal/theme/
```

Expected: all pass (existing tests should not break).

- [ ] **Step 6: Commit**

```bash
git add internal/theme/glamour.go internal/theme/glamour_test.go internal/theme/theme.go
git commit -m "Add theme-to-Glamour style mapping

New GlamourStyle() method on Theme converts TOML token definitions
to a glamour ansi.StyleConfig. Heading, bold, italic, link_text,
and rule tokens map to their Glamour equivalents. Document margins
set to zero for aerc's viewer.

Stores original tokenDefinition structs on Theme for hex color
access needed by Glamour (which takes hex strings, not SGR codes).

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 3: HTML-to-Markdown Converter with Layout Table Plugin

**Files:**
- Create: `internal/filter/convert.go`
- Create: `internal/filter/convert_test.go`

- [ ] **Step 1: Write the failing tests**

```go
// internal/filter/convert_test.go
package filter

import (
	"strings"
	"testing"
)

func TestConvertHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string // substring that must appear
	}{
		{
			name: "simple paragraph",
			html: "<p>Hello world</p>",
			want: "Hello world",
		},
		{
			name: "bold text",
			html: "<p><strong>Important</strong></p>",
			want: "**Important**",
		},
		{
			name: "link preserved",
			html: `<p><a href="https://example.com">Click here</a></p>`,
			want: "https://example.com",
		},
		{
			name: "heading",
			html: "<h2>Section Title</h2>",
			want: "## Section Title",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertHTML(tt.html)
			if err != nil {
				t.Fatalf("convertHTML: %v", err)
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("output missing %q\ngot: %s", tt.want, got)
			}
		})
	}
}

func TestConvertHTMLLayoutTable(t *testing.T) {
	// Layout table: no <th> → should be flattened.
	html := `<table><tr><td>Cell 1</td><td>Cell 2</td></tr></table>`
	got, err := convertHTML(html)
	if err != nil {
		t.Fatalf("convertHTML: %v", err)
	}
	// Should NOT contain pipe table syntax.
	if strings.Contains(got, "|") {
		t.Errorf("layout table should be flattened, got pipe table:\n%s", got)
	}
	if !strings.Contains(got, "Cell 1") || !strings.Contains(got, "Cell 2") {
		t.Errorf("cell content should be preserved:\n%s", got)
	}
}

func TestConvertHTMLDataTable(t *testing.T) {
	// Data table: has <th> → should render as pipe table.
	html := `<table>
		<thead><tr><th>Name</th><th>Age</th></tr></thead>
		<tbody><tr><td>Alice</td><td>30</td></tr></tbody>
	</table>`
	got, err := convertHTML(html)
	if err != nil {
		t.Fatalf("convertHTML: %v", err)
	}
	// Should contain pipe table syntax.
	if !strings.Contains(got, "|") {
		t.Errorf("data table should be a pipe table:\n%s", got)
	}
	if !strings.Contains(got, "Name") || !strings.Contains(got, "Alice") {
		t.Errorf("table content should be preserved:\n%s", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/filter/ -run TestConvertHTML -v
```

Expected: FAIL — `convertHTML` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/filter/convert.go
package filter

import (
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
	"golang.org/x/net/html"
)

// layoutTablePlugin flattens HTML tables that lack <th> elements (layout
// tables) into sequential paragraphs. Tables with <th> elements (data
// tables) pass through to the table plugin for pipe table rendering.
type layoutTablePlugin struct{}

func (p *layoutTablePlugin) Name() string { return "layout-table" }

func (p *layoutTablePlugin) Init(conv *converter.Converter) error {
	conv.Register.RendererFor("table", converter.TagTypeBlock,
		p.renderTable, converter.PriorityEarly)
	return nil
}

// hasTableHeader walks the table node's children looking for any <th> element.
func hasTableHeader(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "th" {
		return true
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if hasTableHeader(c) {
			return true
		}
	}
	return false
}

func (p *layoutTablePlugin) renderTable(ctx converter.Context, w converter.Writer, n *html.Node) converter.RenderStatus {
	if hasTableHeader(n) {
		// Data table — let the table plugin handle it.
		return converter.RenderTryNext
	}

	// Layout table — render each cell's content as a paragraph.
	renderCells(ctx, w, n)
	return converter.RenderSuccess
}

// renderCells recursively finds <td> elements and renders their children
// separated by blank lines.
func renderCells(ctx converter.Context, w converter.Writer, n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "td" {
		w.WriteString("\n\n")
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ctx.RenderChildNodes(ctx, w, c)
		}
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		renderCells(ctx, w, c)
	}
}

// newConverter creates an html-to-markdown converter with layout table
// detection and GFM table support for data tables.
func newConverter() *converter.Converter {
	return converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
			&layoutTablePlugin{},
		),
	)
}

// convertHTML converts an HTML string to markdown. Layout tables are
// flattened; data tables become pipe tables.
func convertHTML(input string) (string, error) {
	conv := newConverter()
	md, err := conv.ConvertString(input)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(md), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/filter/ -run TestConvertHTML -v
```

Expected: PASS for all three tests.

- [ ] **Step 5: Commit**

```bash
git add internal/filter/convert.go internal/filter/convert_test.go
git commit -m "Add html-to-markdown converter with layout table detection

Tables with <th> elements are preserved as pipe tables. Tables
without <th> (layout tables from marketing emails) are flattened
into sequential paragraphs. Uses html-to-markdown v2 with a custom
plugin registered at PriorityEarly.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 4: Rewrite HTML Filter Pipeline

**Files:**
- Rewrite: `internal/filter/html.go`
- Rewrite: `internal/filter/html_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/filter/html_test.go
package filter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/glw907/beautiful-aerc/internal/theme"
)

func testTheme(t *testing.T) *theme.Theme {
	t.Helper()
	content := `name = "test"

[colors]
bg_base = "#2e3440"
bg_elevated = "#3b4252"
bg_selection = "#394353"
bg_border = "#49576b"
fg_base = "#d8dee9"
fg_bright = "#e5e9f0"
fg_brightest = "#eceff4"
fg_dim = "#616e88"
accent_primary = "#81a1c1"
accent_secondary = "#88c0d0"
accent_tertiary = "#8fbcbb"
color_error = "#bf616a"
color_warning = "#d08770"
color_success = "#a3be8c"
color_info = "#ebcb8b"
color_special = "#b48ead"

[tokens]
heading = { color = "color_success", bold = true }
bold = { bold = true }
italic = { italic = true }
link_text = { color = "accent_secondary" }
rule = { color = "fg_dim" }
hdr_key = { color = "accent_primary", bold = true }
hdr_value = { color = "fg_base" }
hdr_dim = { color = "fg_dim" }
`
	dir := t.TempDir()
	path := dir + "/test.toml"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	th, err := theme.Load(path)
	if err != nil {
		t.Fatalf("loading theme: %v", err)
	}
	return th
}

func TestHTML(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string // substring in ANSI-stripped output
	}{
		{
			name: "simple paragraph",
			html: "<p>Hello world</p>",
			want: "Hello world",
		},
		{
			name: "heading rendered",
			html: "<h2>Title</h2><p>Body</p>",
			want: "Title",
		},
		{
			name: "link text preserved",
			html: `<p><a href="https://example.com">Click</a></p>`,
			want: "Click",
		},
		{
			name: "data table rendered",
			html: `<table><thead><tr><th>A</th><th>B</th></tr></thead>
			        <tbody><tr><td>1</td><td>2</td></tr></tbody></table>`,
			want: "A",
		},
		{
			name: "tracking image stripped",
			html: `<p>Text</p><img width="0" height="0" src="https://track.example.com/pixel.gif">`,
			want: "Text",
		},
		{
			name: "display none stripped",
			html: `<p>Visible</p><div style="display:none">Hidden</div>`,
			want: "Visible",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := testTheme(t)
			var buf bytes.Buffer
			err := HTML(strings.NewReader(tt.html), &buf, th, 80)
			if err != nil {
				t.Fatalf("HTML: %v", err)
			}
			plain := stripANSI(buf.String())
			if !strings.Contains(plain, tt.want) {
				t.Errorf("output missing %q\ngot: %s", tt.want, plain)
			}
		})
	}
}

func TestHTMLDisplayNoneNotInOutput(t *testing.T) {
	th := testTheme(t)
	html := `<p>Show</p><div style="display:none"><p>Secret</p></div>`
	var buf bytes.Buffer
	if err := HTML(strings.NewReader(html), &buf, th, 80); err != nil {
		t.Fatal(err)
	}
	plain := stripANSI(buf.String())
	if strings.Contains(plain, "Secret") {
		t.Error("display:none content should be stripped")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/filter/ -run TestHTML -v
```

Expected: FAIL — the current `HTML` function has a different implementation. The test may compile but fail due to changed behavior, or the new test helpers may conflict. We'll resolve in the next step.

- [ ] **Step 3: Rewrite html.go**

Replace the entire contents of `internal/filter/html.go` with:

```go
package filter

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

// Package-level compiled regexes for HTML pre-cleaning.
var (
	reMozClass    = regexp.MustCompile(` class="moz-[^"]*"`)
	reMozDataAttr = regexp.MustCompile(` data-moz-do-not-send="[^"]*"`)
	reMozAttr     = regexp.MustCompile(` moz-do-not-send="[^"]*"`)
	reHiddenDivOpen = regexp.MustCompile(`(?i)<div[^>]+style="[^"]*display:\s*none[^"]*"[^>]*>`)
	reZeroImg     = regexp.MustCompile(`(?i)<img[^>]*(?:width:\s*0|height:\s*0|width="0"|height="0")[^>]*/?>`)
	reANSI        = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

// prepareHTML cleans raw HTML before conversion: strips Mozilla-specific
// attributes, hidden elements, and zero-size tracking images.
func prepareHTML(body string) string {
	body = reMozClass.ReplaceAllString(body, "")
	body = reMozDataAttr.ReplaceAllString(body, "")
	body = reMozAttr.ReplaceAllString(body, "")
	body = stripHiddenElements(body)
	body = reZeroImg.ReplaceAllString(body, "")
	return body
}

// stripHiddenElements removes <div> elements whose inline style contains
// display:none. Uses nesting-aware depth tracking.
func stripHiddenElements(body string) string {
	for {
		loc := reHiddenDivOpen.FindStringIndex(body)
		if loc == nil {
			break
		}
		start := loc[0]
		rest := body[loc[1]:]
		depth := 1
		pos := 0
		for depth > 0 && pos < len(rest) {
			nextOpen := strings.Index(rest[pos:], "<div")
			nextClose := strings.Index(rest[pos:], "</div>")
			if nextClose < 0 {
				pos = len(rest)
				break
			}
			if nextOpen >= 0 && nextOpen < nextClose {
				depth++
				pos += nextOpen + len("<div")
			} else {
				depth--
				pos += nextClose + len("</div>")
			}
		}
		end := loc[1] + pos
		if end > len(body) {
			end = len(body)
		}
		body = body[:start] + body[end:]
	}
	return body
}

// stripANSI removes ANSI escape sequences from s.
func stripANSI(s string) string {
	return reANSI.ReplaceAllString(s, "")
}

// Markdown converts HTML email to clean markdown without ANSI styling.
// Used by the markdown subcommand for reply templates.
func Markdown(r io.Reader, w io.Writer, cols int) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	cleaned := prepareHTML(string(raw))
	md, err := convertHTML(cleaned)
	if err != nil {
		return fmt.Errorf("converting html: %w", err)
	}

	if _, err := fmt.Fprint(w, md+"\n"); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// HTML reads raw HTML email from r, converts it to markdown, and renders
// it to styled ANSI output via Glamour using theme t. cols sets the
// terminal width for wrapping.
func HTML(r io.Reader, w io.Writer, t *theme.Theme, cols int) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	cleaned := prepareHTML(string(raw))
	md, err := convertHTML(cleaned)
	if err != nil {
		return fmt.Errorf("converting html: %w", err)
	}

	style := t.GlamourStyle()
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(cols),
	)
	if err != nil {
		return fmt.Errorf("creating renderer: %w", err)
	}

	styled, err := renderer.Render(md)
	if err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}

	if _, err := fmt.Fprint(w, styled); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Remove old test file and add `os` import to new test**

The old `internal/filter/html_test.go` (if it exists) should be replaced entirely with the new test content from Step 1. Add the `"os"` import to the test file.

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/filter/ -run "TestHTML" -v
```

Expected: PASS for all tests.

- [ ] **Step 6: Commit**

```bash
git add internal/filter/html.go internal/filter/html_test.go
git commit -m "Rewrite HTML filter with html-to-markdown + Glamour

New three-stage pipeline: prepareHTML (email cleanup) →
html-to-markdown (Go-native conversion) → Glamour (ANSI rendering
with OSC 8 links). Replaces pandoc subprocess, custom markdown
highlighting, and footnote system.

Adds Markdown() export for clean markdown output (reply templates).

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 5: Delete Footnotes and Pick-Link

**Files:**
- Delete: `internal/filter/footnotes.go`
- Delete: `internal/filter/footnotes_test.go`
- Delete: `cmd/pick-link/root.go`
- Delete: `cmd/pick-link/main.go`
- Delete: `internal/picker/picker.go`

- [ ] **Step 1: Delete the files**

```bash
cd /home/glw907/Projects/beautiful-aerc
rm internal/filter/footnotes.go internal/filter/footnotes_test.go
rm -r cmd/pick-link/
rm -r internal/picker/
```

- [ ] **Step 2: Verify the build compiles**

```bash
go build ./...
```

Expected: clean build. If there are remaining references to deleted exports (`HTMLLinks`, `FootnoteLink`, `ExtractFootnoteLinks`), they should have been removed in Task 4 when html.go was rewritten. If any remain, fix them.

- [ ] **Step 3: Run full test suite**

```bash
go vet ./... && go test ./...
```

Expected: all pass. E2E tests may fail due to changed output — that's expected and handled in Task 8.

- [ ] **Step 4: Commit**

```bash
git add -u
git commit -m "Delete footnotes, pick-link binary, and picker package

Footnotes replaced by Glamour's OSC 8 hyperlinks. Pick-link
replaced by terminal Ctrl+click on OSC 8 links. Removes ~600
lines of custom code.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 6: Add `markdown` Subcommand

**Files:**
- Create: `cmd/mailrender/markdown.go`
- Modify: `cmd/mailrender/root.go`

- [ ] **Step 1: Write the subcommand**

```go
// cmd/mailrender/markdown.go
package main

import (
	"os"

	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/spf13/cobra"
)

func newMarkdownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "markdown",
		Short: "Convert HTML email to clean markdown (for reply templates)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cols := termCols()
			return filter.Markdown(os.Stdin, os.Stdout, cols)
		},
	}
	return cmd
}
```

- [ ] **Step 2: Register in root.go**

Add to `newRootCmd()` in `cmd/mailrender/root.go`:

```go
	cmd.AddCommand(newMarkdownCmd())
```

- [ ] **Step 3: Build and smoke test**

```bash
go build -o mailrender ./cmd/mailrender
echo '<p>Hello <strong>world</strong></p>' | ./mailrender markdown
```

Expected: output contains `Hello **world**` (clean markdown, no ANSI).

- [ ] **Step 4: Commit**

```bash
git add cmd/mailrender/markdown.go cmd/mailrender/root.go
git commit -m "Add markdown subcommand for reply templates

Converts HTML email to clean markdown without ANSI styling.
Intended for aerc reply templates so quoted text matches what
the user read in the viewer.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 7: Add `to-html` Subcommand

**Files:**
- Create: `cmd/mailrender/tohtml.go`
- Create: `internal/filter/tohtml.go`
- Create: `internal/filter/tohtml_test.go`
- Modify: `cmd/mailrender/root.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/filter/tohtml_test.go
package filter

import (
	"bytes"
	"strings"
	"testing"
)

func TestToHTML(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		want     string // substring in output
	}{
		{
			name:     "paragraph",
			markdown: "Hello world",
			want:     "<p>Hello world</p>",
		},
		{
			name:     "bold",
			markdown: "**Important**",
			want:     "<strong>Important</strong>",
		},
		{
			name:     "heading",
			markdown: "## Title",
			want:     "<h2>Title</h2>",
		},
		{
			name:     "link",
			markdown: "[Click](https://example.com)",
			want:     `<a href="https://example.com">Click</a>`,
		},
		{
			name:     "html document wrapper",
			markdown: "Hello",
			want:     "<!DOCTYPE html>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := ToHTML(strings.NewReader(tt.markdown), &buf)
			if err != nil {
				t.Fatalf("ToHTML: %v", err)
			}
			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("output missing %q\ngot: %s", tt.want, buf.String())
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/filter/ -run TestToHTML -v
```

Expected: FAIL — `ToHTML` undefined.

- [ ] **Step 3: Write the implementation**

```go
// internal/filter/tohtml.go
package filter

import (
	"bytes"
	"fmt"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// ToHTML converts markdown from r to a standalone HTML document written to w.
// Replaces pandoc in aerc's multipart-converters.
func ToHTML(r io.Reader, w io.Writer) error {
	src, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
	)

	var body bytes.Buffer
	if err := md.Convert(src, &body); err != nil {
		return fmt.Errorf("converting markdown: %w", err)
	}

	const head = `<!DOCTYPE html>
<html>
<head><meta charset="utf-8"></head>
<body>
`
	const tail = `</body>
</html>
`
	if _, err := fmt.Fprint(w, head+body.String()+tail); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}
```

- [ ] **Step 4: Write the subcommand**

```go
// cmd/mailrender/tohtml.go
package main

import (
	"os"

	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/spf13/cobra"
)

func newToHTMLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "to-html",
		Short: "Convert markdown to HTML (for multipart-converters)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return filter.ToHTML(os.Stdin, os.Stdout)
		},
	}
	return cmd
}
```

- [ ] **Step 5: Register in root.go**

Add to `newRootCmd()` in `cmd/mailrender/root.go`:

```go
	cmd.AddCommand(newToHTMLCmd())
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/filter/ -run TestToHTML -v
```

Expected: PASS for all tests.

- [ ] **Step 7: Smoke test**

```bash
go build -o mailrender ./cmd/mailrender
echo '## Hello **world**' | ./mailrender to-html
```

Expected: HTML document with `<h2>Hello <strong>world</strong></h2>`.

- [ ] **Step 8: Commit**

```bash
git add internal/filter/tohtml.go internal/filter/tohtml_test.go cmd/mailrender/tohtml.go cmd/mailrender/root.go
git commit -m "Add to-html subcommand for multipart-converters

Converts markdown to standalone HTML using goldmark. Replaces
pandoc in aerc's multipart-converters config. Supports tables
via goldmark's table extension.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 8: Update Makefile and Config

**Files:**
- Modify: `Makefile`
- Modify: `.config/aerc/aerc.conf`

- [ ] **Step 1: Update Makefile**

Remove `pick-link` from `build`, `install`, and `clean` targets. The Makefile should become:

```makefile
build:
	go build -o mailrender ./cmd/mailrender
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
	GOBIN=$(HOME)/.local/bin go install ./cmd/fastmail-cli
	GOBIN=$(HOME)/.local/bin go install ./cmd/tidytext
	GOBIN=$(HOME)/.local/bin go install ./cmd/compose-prep

check: vet test

clean:
	rm -f mailrender fastmail-cli tidytext compose-prep

.PHONY: build test vet lint install check clean
```

- [ ] **Step 2: Update aerc.conf multipart-converters**

In `.config/aerc/aerc.conf`, change the multipart-converters line from:

```ini
text/html = pandoc -f markdown -t html --standalone
```

to:

```ini
text/html = mailrender to-html
```

- [ ] **Step 3: Verify build**

```bash
make build
```

Expected: builds four binaries, no pick-link.

- [ ] **Step 4: Commit**

```bash
git add Makefile .config/aerc/aerc.conf
git commit -m "Remove pick-link from build, use mailrender to-html

Update Makefile to build four binaries (mailrender, fastmail-cli,
tidytext, compose-prep). Update aerc.conf multipart-converters to
use mailrender to-html instead of pandoc.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 9: Update E2E Tests and Golden Files

**Files:**
- Modify: `e2e/e2e_test.go`
- Regenerate: `e2e/testdata/golden/*.txt`

- [ ] **Step 1: Run E2E tests to see current failures**

```bash
cd /home/glw907/Projects/beautiful-aerc
go test ./e2e/ -v
```

Expected: failures because output format changed (Glamour rendering vs old footnote+highlight format).

- [ ] **Step 2: Regenerate golden files**

```bash
go test ./e2e/ -update-golden -v
```

Expected: golden files rewritten with new Glamour output.

- [ ] **Step 3: Verify E2E tests pass with new golden files**

```bash
go test ./e2e/ -v
```

Expected: all pass.

- [ ] **Step 4: Manually review golden files**

Read through each golden file to verify the output looks correct:

```bash
cat e2e/testdata/golden/marketing.txt
cat e2e/testdata/golden/transactional.txt
cat e2e/testdata/golden/developer.txt
cat e2e/testdata/golden/simple.txt
cat e2e/testdata/golden/edge-links.txt
```

Check: headings styled, bold/italic rendered, data tables have borders, layout tables are flattened, links are present as text, no raw HTML fragments.

- [ ] **Step 5: Run full test suite**

```bash
make check
```

Expected: all tests pass, vet clean.

- [ ] **Step 6: Commit**

```bash
git add e2e/
git commit -m "Regenerate E2E golden files for Glamour pipeline

Output format changed: Glamour rendering replaces custom markdown
highlighting and footnote references. Links now use OSC 8 instead
of [^N] footnotes. Data tables render with box-drawing borders.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 10: Per-Line ANSI Verification and Fix

**Files:**
- Possibly modify: `internal/filter/html.go`

- [ ] **Step 1: Test with aerc**

```bash
make install
```

Open aerc, view an email with bold text spanning multiple lines or a colored heading. Check whether ANSI styling carries across lines correctly.

- [ ] **Step 2: If ANSI state is lost across lines, add post-processing**

If aerc strips carry-over ANSI state, add a `reemitANSI` function to `html.go` that:
1. Tracks open ANSI state (bold, italic, color) as it scans each line
2. At the start of each line, re-emits any open ANSI codes
3. Runs as a post-processing step after Glamour

This is a conditional step — only implement if testing reveals the problem.

- [ ] **Step 3: If needed, test and commit the fix**

```bash
go test ./internal/filter/ -v
make install
```

Verify in aerc, then:

```bash
git add internal/filter/html.go
git commit -m "Add per-line ANSI re-emission for aerc viewer

aerc's viewer renders each line independently, losing carry-over
ANSI state. Post-process Glamour output to re-emit open ANSI
codes at the start of each line.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 11: Update Personal Dotfiles

**Files:**
- Modify: `~/.dotfiles/beautiful-aerc/.config/aerc/aerc.conf`
- Modify: `~/.dotfiles/beautiful-aerc/.config/aerc/binds.conf`

Per CLAUDE.md, changes to `.config/aerc/` must be applied to both project repo and personal dotfiles.

- [ ] **Step 1: Update dotfiles aerc.conf**

Apply the same `multipart-converters` change from Task 8 to the personal dotfiles copy.

- [ ] **Step 2: Remove pick-link keybinding from dotfiles binds.conf**

Remove the `Tab` binding for pick-link in the `[view]` section of `~/.dotfiles/beautiful-aerc/.config/aerc/binds.conf`.

- [ ] **Step 3: Remove pick-link binary**

```bash
rm ~/.local/bin/pick-link
```

- [ ] **Step 4: Commit dotfiles changes**

```bash
cd ~/.dotfiles
git add beautiful-aerc/.config/aerc/aerc.conf beautiful-aerc/.config/aerc/binds.conf
git commit -m "Update aerc config for Glamour pipeline migration

Use mailrender to-html instead of pandoc for multipart-converters.
Remove pick-link keybinding (replaced by OSC 8 links).

Co-Authored-By: Claude <noreply@anthropic.com>"
```

---

## Task 12: Clean Up go.mod and Final Verification

**Files:**
- Modify: `go.mod` / `go.sum`

- [ ] **Step 1: Tidy dependencies**

```bash
cd /home/glw907/Projects/beautiful-aerc
go mod tidy
```

This should remove `golang.org/x/sys` (only used by the deleted picker) if no other code imports it.

- [ ] **Step 2: Full verification**

```bash
make check
make build
make install
```

Expected: all pass, all binaries build and install.

- [ ] **Step 3: Live smoke test**

Open aerc and verify:
- HTML email renders with Glamour styling (headings, bold, tables)
- Links are clickable via Ctrl+click in kitty
- Reply quotes the same markdown you read
- Sending via `y` in compose review produces clean HTML

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum
git commit -m "Tidy go.mod after Glamour pipeline migration

Remove unused dependencies from pick-link and pandoc integration.

Co-Authored-By: Claude <noreply@anthropic.com>"
```

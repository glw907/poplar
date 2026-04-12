# Mailrender Lipgloss Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use
> superpowers:subagent-driven-development (recommended) or
> superpowers:executing-plans to implement this plan task-by-task.
> Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace glamour rendering and TOML theme system with a
lipgloss-native block model that serves both aerc CLI filters and
the poplar viewer.

**Architecture:** Three-layer split: content pipeline
(`internal/filter/`) produces clean markdown, block model
(`internal/content/`) parses it into email-aware blocks and renders
with lipgloss, compiled themes (`internal/theme/`) provide lipgloss
styles. The CLI filter and poplar viewer call the same rendering
functions.

**Tech Stack:** Go, lipgloss, html-to-markdown, goldmark, cobra

**Spec:** `docs/superpowers/specs/2026-04-10-mailrender-lipgloss-design.md`

---

## File Structure

### New files

| File | Responsibility |
|---|---|
| `internal/content/blocks.go` | Block and Span type definitions |
| `internal/content/parse.go` | Markdown to `[]Block` parser |
| `internal/content/parse_test.go` | Block parser tests |
| `internal/content/headers.go` | `ParsedHeaders` type and `ParseHeaders` |
| `internal/content/headers_test.go` | Header parser tests |
| `internal/content/render.go` | `RenderBody` and `RenderHeaders` with lipgloss |
| `internal/content/render_test.go` | Renderer tests |
| `internal/theme/palette.go` | `Palette` type and `NewTheme` constructor |
| `internal/theme/themes.go` | `Nord`, `SolarizedDark`, `GruvboxDark` compiled values |
| `internal/theme/palette_test.go` | Compiled theme tests |

### Modified files

| File | Change |
|---|---|
| `internal/filter/html.go` | Extract `CleanHTML`, delete rendering stages |
| `internal/filter/plain.go` | Extract `CleanPlain`, delete subprocess calls |
| `internal/filter/html_test.go` | Update tests for new API |
| `internal/filter/plain_test.go` | Update tests for new API |
| `internal/theme/theme.go` | Replace TOML loader with compiled palette |
| `internal/theme/glamour.go` | Delete entirely |
| `internal/theme/glamour_test.go` | Delete entirely |
| `internal/theme/theme_test.go` | Rewrite for compiled themes |
| `internal/theme/styleset.go` | Update to use new `Theme` struct |
| `internal/theme/styleset_test.go` | Update for new theme API |
| `cmd/mailrender/root.go` | Add `preview` subcommand |
| `cmd/mailrender/headers.go` | Replace `loadTheme`/`colorsFromTheme` with compiled theme, use `content.RenderHeaders` |
| `cmd/mailrender/html.go` | Use `filter.CleanHTML` + `content.ParseBlocks` + `content.RenderBody` |
| `cmd/mailrender/plain.go` | Use `filter.CleanPlain` + same render path |
| `cmd/mailrender/markdown.go` | Use `filter.CleanHTML` + `content.ParseBlocks` + plain text output |
| `cmd/mailrender/tohtml.go` | Move `ToHTML` to `internal/content/` |
| `cmd/mailrender/themes.go` | Use compiled themes instead of loading TOML |
| `e2e/e2e_test.go` | Update golden test setup (no TOML theme) |
| `go.mod` | Remove glamour, promote lipgloss to direct |

### Deleted files

| File | Reason |
|---|---|
| `internal/theme/glamour.go` | Glamour removed |
| `internal/theme/glamour_test.go` | Glamour removed |

---

## Phase 1: Block Model and Parser

### Task 1: Block and Span types

**Files:**
- Create: `internal/content/blocks.go`

- [ ] **Step 1: Create the block and span type definitions**

```go
// internal/content/blocks.go
package content

// Block represents a semantic unit of email content.
type Block interface {
	blockType() blockKind
}

type blockKind int

const (
	kindParagraph blockKind = iota
	kindHeading
	kindBlockquote
	kindQuoteAttribution
	kindSignature
	kindRule
	kindCodeBlock
	kindTable
	kindListItem
)

// Span represents an inline styled segment within a block.
type Span interface {
	spanType() spanKind
}

type spanKind int

const (
	kindText spanKind = iota
	kindBold
	kindItalic
	kindCode
	kindLink
)

type Paragraph struct{ Spans []Span }
type Heading struct {
	Spans []Span
	Level int
}
type Blockquote struct {
	Blocks []Block
	Level  int
}
type QuoteAttribution struct{ Spans []Span }
type Signature struct{ Lines [][]Span }
type Rule struct{}
type CodeBlock struct {
	Text string
	Lang string
}
type Table struct {
	Headers [][]Span
	Rows    [][][]Span
}
type ListItem struct {
	Spans   []Span
	Ordered bool
	Index   int
}

func (Paragraph) blockType() blockKind        { return kindParagraph }
func (Heading) blockType() blockKind          { return kindHeading }
func (Blockquote) blockType() blockKind       { return kindBlockquote }
func (QuoteAttribution) blockType() blockKind { return kindQuoteAttribution }
func (Signature) blockType() blockKind        { return kindSignature }
func (Rule) blockType() blockKind             { return kindRule }
func (CodeBlock) blockType() blockKind        { return kindCodeBlock }
func (Table) blockType() blockKind            { return kindTable }
func (ListItem) blockType() blockKind         { return kindListItem }

type Text struct{ Content string }
type Bold struct{ Content string }
type Italic struct{ Content string }
type Code struct{ Content string }
type Link struct {
	Text string
	URL  string
}

func (Text) spanType() spanKind   { return kindText }
func (Bold) spanType() spanKind   { return kindBold }
func (Italic) spanType() spanKind { return kindItalic }
func (Code) spanType() spanKind   { return kindCode }
func (Link) spanType() spanKind   { return kindLink }

// Address is a parsed email address with optional display name.
type Address struct {
	Name  string
	Email string
}

// ParsedHeaders holds the structured header fields from an email.
type ParsedHeaders struct {
	From    []Address
	To      []Address
	Cc      []Address
	Bcc     []Address
	Date    string
	Subject string
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./internal/content/`
Expected: success, no errors

- [ ] **Step 3: Commit**

```bash
git add internal/content/blocks.go
git commit -m "Add block and span type definitions for content model"
```

---

### Task 2: Inline span parser

**Files:**
- Create: `internal/content/parse.go`
- Create: `internal/content/parse_test.go`

- [ ] **Step 1: Write failing tests for inline span parsing**

```go
// internal/content/parse_test.go
package content

import (
	"testing"
)

func spansEqual(t *testing.T, got, want []Span) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("span count: got %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("span[%d]: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestParseSpans(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Span
	}{
		{
			name:  "plain text",
			input: "hello world",
			want:  []Span{Text{Content: "hello world"}},
		},
		{
			name:  "bold",
			input: "hello **world**",
			want:  []Span{Text{Content: "hello "}, Bold{Content: "world"}},
		},
		{
			name:  "italic",
			input: "hello *world*",
			want:  []Span{Text{Content: "hello "}, Italic{Content: "world"}},
		},
		{
			name:  "inline code",
			input: "use `fmt.Println`",
			want:  []Span{Text{Content: "use "}, Code{Content: "fmt.Println"}},
		},
		{
			name:  "link",
			input: "visit [example](https://example.com) today",
			want: []Span{
				Text{Content: "visit "},
				Link{Text: "example", URL: "https://example.com"},
				Text{Content: " today"},
			},
		},
		{
			name:  "mixed",
			input: "**bold** and *italic* and `code`",
			want: []Span{
				Bold{Content: "bold"},
				Text{Content: " and "},
				Italic{Content: "italic"},
				Text{Content: " and "},
				Code{Content: "code"},
			},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSpans(tt.input)
			spansEqual(t, got, tt.want)
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/content/ -run TestParseSpans -v`
Expected: FAIL — `parseSpans` undefined

- [ ] **Step 3: Implement the span parser**

```go
// internal/content/parse.go
package content

import (
	"strings"
)

// parseSpans parses inline markdown formatting into a slice of Span values.
// Handles: **bold**, *italic*, `code`, [text](url).
func parseSpans(input string) []Span {
	if input == "" {
		return nil
	}
	var spans []Span
	remaining := input

	for len(remaining) > 0 {
		// Find the earliest inline marker
		boldIdx := strings.Index(remaining, "**")
		italicIdx := -1
		// Only match single * that isn't part of **
		for i := 0; i < len(remaining); i++ {
			if remaining[i] == '*' {
				if i+1 < len(remaining) && remaining[i+1] == '*' {
					i++ // skip **
					continue
				}
				italicIdx = i
				break
			}
		}
		codeIdx := strings.Index(remaining, "`")
		linkIdx := strings.Index(remaining, "[")

		// Find the earliest marker
		best := len(remaining)
		bestKind := -1
		for _, candidate := range []struct {
			idx  int
			kind int
		}{
			{boldIdx, 0},
			{italicIdx, 1},
			{codeIdx, 2},
			{linkIdx, 3},
		} {
			if candidate.idx >= 0 && candidate.idx < best {
				best = candidate.idx
				bestKind = candidate.kind
			}
		}

		if bestKind == -1 {
			// No more markers — rest is plain text
			spans = append(spans, Text{Content: remaining})
			break
		}

		// Add any text before the marker
		if best > 0 {
			spans = append(spans, Text{Content: remaining[:best]})
		}

		switch bestKind {
		case 0: // **bold**
			end := strings.Index(remaining[best+2:], "**")
			if end < 0 {
				spans = append(spans, Text{Content: remaining[best:]})
				remaining = ""
				continue
			}
			spans = append(spans, Bold{Content: remaining[best+2 : best+2+end]})
			remaining = remaining[best+2+end+2:]

		case 1: // *italic*
			end := strings.Index(remaining[best+1:], "*")
			if end < 0 {
				spans = append(spans, Text{Content: remaining[best:]})
				remaining = ""
				continue
			}
			spans = append(spans, Italic{Content: remaining[best+1 : best+1+end]})
			remaining = remaining[best+1+end+1:]

		case 2: // `code`
			end := strings.Index(remaining[best+1:], "`")
			if end < 0 {
				spans = append(spans, Text{Content: remaining[best:]})
				remaining = ""
				continue
			}
			spans = append(spans, Code{Content: remaining[best+1 : best+1+end]})
			remaining = remaining[best+1+end+1:]

		case 3: // [text](url)
			closeBracket := strings.Index(remaining[best:], "](")
			if closeBracket < 0 {
				spans = append(spans, Text{Content: remaining[best:]})
				remaining = ""
				continue
			}
			closeParen := strings.Index(remaining[best+closeBracket+2:], ")")
			if closeParen < 0 {
				spans = append(spans, Text{Content: remaining[best:]})
				remaining = ""
				continue
			}
			linkText := remaining[best+1 : best+closeBracket]
			linkURL := remaining[best+closeBracket+2 : best+closeBracket+2+closeParen]
			spans = append(spans, Link{Text: linkText, URL: linkURL})
			remaining = remaining[best+closeBracket+2+closeParen+1:]
		}
	}

	return spans
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/content/ -run TestParseSpans -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/content/parse.go internal/content/parse_test.go
git commit -m "Add inline span parser for markdown formatting"
```

---

### Task 3: Block parser

**Files:**
- Modify: `internal/content/parse.go`
- Modify: `internal/content/parse_test.go`

- [ ] **Step 1: Write failing tests for block parsing**

```go
// append to internal/content/parse_test.go

func TestParseBlocks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		types []blockKind
	}{
		{
			name:  "single paragraph",
			input: "Hello world",
			types: []blockKind{kindParagraph},
		},
		{
			name:  "two paragraphs",
			input: "First paragraph.\n\nSecond paragraph.",
			types: []blockKind{kindParagraph, kindParagraph},
		},
		{
			name:  "heading",
			input: "# Title\n\nBody text.",
			types: []blockKind{kindHeading, kindParagraph},
		},
		{
			name:  "heading levels",
			input: "## Level 2\n\n### Level 3",
			types: []blockKind{kindHeading, kindHeading},
		},
		{
			name:  "blockquote",
			input: "> quoted text",
			types: []blockKind{kindBlockquote},
		},
		{
			name:  "nested blockquote",
			input: "> > deeply quoted",
			types: []blockKind{kindBlockquote},
		},
		{
			name:  "horizontal rule",
			input: "Above\n\n---\n\nBelow",
			types: []blockKind{kindParagraph, kindRule, kindParagraph},
		},
		{
			name:  "code block",
			input: "```go\nfmt.Println()\n```",
			types: []blockKind{kindCodeBlock},
		},
		{
			name:  "unordered list",
			input: "- item one\n- item two",
			types: []blockKind{kindListItem, kindListItem},
		},
		{
			name:  "ordered list",
			input: "1. first\n2. second",
			types: []blockKind{kindListItem, kindListItem},
		},
		{
			name:  "signature",
			input: "Body text.\n\n-- \nGeoff Wright\ngeoff@907.life",
			types: []blockKind{kindParagraph, kindSignature},
		},
		{
			name:  "quote attribution",
			input: "On Mon, Jan 5, Alice wrote:\n\n> quoted reply",
			types: []blockKind{kindQuoteAttribution, kindBlockquote},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := ParseBlocks(tt.input)
			if len(blocks) != len(tt.types) {
				t.Fatalf("block count: got %d, want %d\nblocks: %v", len(blocks), len(tt.types), blocks)
			}
			for i, b := range blocks {
				if b.blockType() != tt.types[i] {
					t.Errorf("block[%d]: got kind %d, want %d", i, b.blockType(), tt.types[i])
				}
			}
		})
	}
}

func TestParseBlocksHeadingLevel(t *testing.T) {
	blocks := ParseBlocks("## Level 2")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	h, ok := blocks[0].(Heading)
	if !ok {
		t.Fatalf("expected Heading, got %T", blocks[0])
	}
	if h.Level != 2 {
		t.Errorf("level: got %d, want 2", h.Level)
	}
}

func TestParseBlocksCodeBlockLang(t *testing.T) {
	blocks := ParseBlocks("```python\nprint('hi')\n```")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	cb, ok := blocks[0].(CodeBlock)
	if !ok {
		t.Fatalf("expected CodeBlock, got %T", blocks[0])
	}
	if cb.Lang != "python" {
		t.Errorf("lang: got %q, want %q", cb.Lang, "python")
	}
	if cb.Text != "print('hi')" {
		t.Errorf("text: got %q, want %q", cb.Text, "print('hi')")
	}
}

func TestParseBlocksSignatureContent(t *testing.T) {
	blocks := ParseBlocks("Body.\n\n-- \nLine 1\nLine 2")
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	sig, ok := blocks[1].(Signature)
	if !ok {
		t.Fatalf("expected Signature, got %T", blocks[1])
	}
	if len(sig.Lines) != 2 {
		t.Errorf("signature lines: got %d, want 2", len(sig.Lines))
	}
}

func TestParseBlocksBlockquoteLevel(t *testing.T) {
	blocks := ParseBlocks("> > nested quote")
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	bq, ok := blocks[0].(Blockquote)
	if !ok {
		t.Fatalf("expected Blockquote, got %T", blocks[0])
	}
	if bq.Level != 1 {
		t.Errorf("outer level: got %d, want 1", bq.Level)
	}
	if len(bq.Blocks) != 1 {
		t.Fatalf("inner blocks: got %d, want 1", len(bq.Blocks))
	}
	inner, ok := bq.Blocks[0].(Blockquote)
	if !ok {
		t.Fatalf("expected inner Blockquote, got %T", bq.Blocks[0])
	}
	if inner.Level != 2 {
		t.Errorf("inner level: got %d, want 2", inner.Level)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/content/ -run TestParseBlocks -v`
Expected: FAIL — `ParseBlocks` undefined

- [ ] **Step 3: Implement the block parser**

Add to `internal/content/parse.go`:

```go
import (
	"regexp"
	"strings"
)

var (
	reHeading     = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	reCodeFence   = regexp.MustCompile("^```(\\w*)$")
	reRule        = regexp.MustCompile(`^(-{3,}|_{3,}|\*{3,})$`)
	reUnordered   = regexp.MustCompile(`^[-*+]\s+(.+)$`)
	reOrdered     = regexp.MustCompile(`^(\d+)\.\s+(.+)$`)
	reQuotePrefix = regexp.MustCompile(`^(>+)\s?(.*)$`)
	reAttribution = regexp.MustCompile(`(?i)^on\s.+wrote:\s*$`)
	reSignature   = regexp.MustCompile(`^-- $`)
)

// ParseBlocks parses normalized markdown into email-aware block types.
func ParseBlocks(markdown string) []Block {
	lines := strings.Split(markdown, "\n")
	var blocks []Block
	i := 0

	for i < len(lines) {
		line := lines[i]

		// Skip blank lines between blocks
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// Signature: "-- " marker, everything after is signature
		if reSignature.MatchString(line) {
			var sigLines [][]Span
			for i++; i < len(lines); i++ {
				sigLines = append(sigLines, parseSpans(lines[i]))
			}
			if len(sigLines) > 0 {
				blocks = append(blocks, Signature{Lines: sigLines})
			}
			break
		}

		// Code fence
		if m := reCodeFence.FindStringSubmatch(line); m != nil {
			lang := m[1]
			var codeLines []string
			i++
			for i < len(lines) && !strings.HasPrefix(lines[i], "```") {
				codeLines = append(codeLines, lines[i])
				i++
			}
			if i < len(lines) {
				i++ // skip closing fence
			}
			blocks = append(blocks, CodeBlock{
				Text: strings.Join(codeLines, "\n"),
				Lang: lang,
			})
			continue
		}

		// Heading
		if m := reHeading.FindStringSubmatch(line); m != nil {
			blocks = append(blocks, Heading{
				Level: len(m[1]),
				Spans: parseSpans(m[2]),
			})
			i++
			continue
		}

		// Horizontal rule
		if reRule.MatchString(strings.TrimSpace(line)) {
			blocks = append(blocks, Rule{})
			i++
			continue
		}

		// Quote attribution (must check before blockquote)
		if reAttribution.MatchString(strings.TrimSpace(line)) {
			blocks = append(blocks, QuoteAttribution{Spans: parseSpans(strings.TrimSpace(line))})
			i++
			continue
		}

		// Blockquote
		if reQuotePrefix.MatchString(line) {
			var quoteLines []string
			for i < len(lines) && reQuotePrefix.MatchString(lines[i]) {
				quoteLines = append(quoteLines, lines[i])
				i++
			}
			blocks = append(blocks, parseBlockquote(quoteLines, 1))
			continue
		}

		// List items (unordered)
		if m := reUnordered.FindStringSubmatch(line); m != nil {
			blocks = append(blocks, ListItem{
				Spans:   parseSpans(m[1]),
				Ordered: false,
			})
			i++
			continue
		}

		// List items (ordered)
		if m := reOrdered.FindStringSubmatch(line); m != nil {
			idx := 0
			for _, c := range m[1] {
				idx = idx*10 + int(c-'0')
			}
			blocks = append(blocks, ListItem{
				Spans:   parseSpans(m[2]),
				Ordered: true,
				Index:   idx,
			})
			i++
			continue
		}

		// Paragraph: collect consecutive non-blank, non-special lines
		var paraLines []string
		for i < len(lines) {
			l := lines[i]
			if strings.TrimSpace(l) == "" {
				break
			}
			if reHeading.MatchString(l) || reCodeFence.MatchString(l) ||
				reRule.MatchString(strings.TrimSpace(l)) || reQuotePrefix.MatchString(l) ||
				reUnordered.MatchString(l) || reOrdered.MatchString(l) ||
				reSignature.MatchString(l) {
				break
			}
			paraLines = append(paraLines, l)
			i++
		}
		if len(paraLines) > 0 {
			blocks = append(blocks, Paragraph{
				Spans: parseSpans(strings.Join(paraLines, " ")),
			})
		}
	}

	return blocks
}

// parseBlockquote parses collected quote-prefixed lines into a
// Blockquote with recursive nesting.
func parseBlockquote(lines []string, level int) Blockquote {
	// Strip one level of "> " prefix
	var stripped []string
	for _, line := range lines {
		m := reQuotePrefix.FindStringSubmatch(line)
		if m != nil {
			prefixLen := len(m[1])
			if prefixLen > 1 {
				// Still has deeper quote levels
				stripped = append(stripped, strings.Repeat(">", prefixLen-1)+" "+m[2])
			} else {
				stripped = append(stripped, m[2])
			}
		} else {
			stripped = append(stripped, line)
		}
	}

	inner := strings.Join(stripped, "\n")
	return Blockquote{
		Blocks: ParseBlocks(inner),
		Level:  level,
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/content/ -run TestParseBlocks -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/content/parse.go internal/content/parse_test.go
git commit -m "Add block parser for markdown to email-aware blocks"
```

---

### Task 4: Header parser

**Files:**
- Create: `internal/content/headers.go`
- Create: `internal/content/headers_test.go`

- [ ] **Step 1: Write failing tests for header parsing**

```go
// internal/content/headers_test.go
package content

import "testing"

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		from    []Address
		to      []Address
		subject string
	}{
		{
			name:    "simple",
			input:   "From: Alice <alice@example.com>\r\nTo: Bob <bob@example.com>\r\nDate: Mon, 5 Jan 2026\r\nSubject: Hello\r\n\r\n",
			from:    []Address{{Name: "Alice", Email: "alice@example.com"}},
			to:      []Address{{Name: "Bob", Email: "bob@example.com"}},
			subject: "Hello",
		},
		{
			name:    "bare email",
			input:   "From: alice@example.com\r\nSubject: Test\r\n\r\n",
			from:    []Address{{Email: "alice@example.com"}},
			subject: "Test",
		},
		{
			name:  "multiple recipients",
			input: "From: Alice <alice@example.com>\r\nTo: Bob <bob@example.com>, Carol <carol@example.com>\r\nSubject: Group\r\n\r\n",
			from:  []Address{{Name: "Alice", Email: "alice@example.com"}},
			to: []Address{
				{Name: "Bob", Email: "bob@example.com"},
				{Name: "Carol", Email: "carol@example.com"},
			},
			subject: "Group",
		},
		{
			name:    "folded header",
			input:   "From: Alice <alice@example.com>\r\nSubject: This is a very\r\n long subject line\r\n\r\n",
			from:    []Address{{Name: "Alice", Email: "alice@example.com"}},
			subject: "This is a very long subject line",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := ParseHeaders(tt.input)
			if len(h.From) != len(tt.from) {
				t.Fatalf("From count: got %d, want %d", len(h.From), len(tt.from))
			}
			for i, a := range h.From {
				if a != tt.from[i] {
					t.Errorf("From[%d]: got %v, want %v", i, a, tt.from[i])
				}
			}
			if len(h.To) != len(tt.to) {
				t.Fatalf("To count: got %d, want %d", len(h.To), len(tt.to))
			}
			for i, a := range h.To {
				if a != tt.to[i] {
					t.Errorf("To[%d]: got %v, want %v", i, a, tt.to[i])
				}
			}
			if h.Subject != tt.subject {
				t.Errorf("Subject: got %q, want %q", h.Subject, tt.subject)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/content/ -run TestParseHeaders -v`
Expected: FAIL — `ParseHeaders` undefined

- [ ] **Step 3: Implement the header parser**

```go
// internal/content/headers.go
package content

import (
	"net/mail"
	"strings"
)

// ParseHeaders parses raw RFC 2822 headers into structured fields.
// Handles continuation lines, CRLF, and bare email addresses.
func ParseHeaders(raw string) ParsedHeaders {
	var h ParsedHeaders

	// Unfold continuation lines and normalize CRLF
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	lines := strings.Split(raw, "\n")

	var unfolded []string
	for _, line := range lines {
		if line == "" {
			break // end of headers
		}
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			// Continuation line
			if len(unfolded) > 0 {
				unfolded[len(unfolded)-1] += " " + strings.TrimSpace(line)
			}
		} else {
			unfolded = append(unfolded, line)
		}
	}

	for _, line := range unfolded {
		colonIdx := strings.IndexByte(line, ':')
		if colonIdx < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:colonIdx]))
		val := strings.TrimSpace(line[colonIdx+1:])

		switch key {
		case "from":
			h.From = parseAddressList(val)
		case "to":
			h.To = parseAddressList(val)
		case "cc":
			h.Cc = parseAddressList(val)
		case "bcc":
			h.Bcc = parseAddressList(val)
		case "date":
			h.Date = val
		case "subject":
			h.Subject = val
		}
	}

	return h
}

// parseAddressList parses a comma-separated list of RFC 5322 addresses.
func parseAddressList(val string) []Address {
	addrs, err := mail.ParseAddressList(val)
	if err != nil {
		// Fallback: treat the whole value as a single bare email
		email := strings.TrimSpace(val)
		email = strings.TrimPrefix(email, "<")
		email = strings.TrimSuffix(email, ">")
		if email != "" {
			return []Address{{Email: email}}
		}
		return nil
	}

	result := make([]Address, len(addrs))
	for i, a := range addrs {
		result[i] = Address{Name: a.Name, Email: a.Address}
	}
	return result
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/content/ -run TestParseHeaders -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/content/headers.go internal/content/headers_test.go
git commit -m "Add RFC 2822 header parser for content model"
```

---

### Task 5: Test block parser against corpus

**Files:**
- Modify: `internal/content/parse_test.go`

- [ ] **Step 1: Write integration test that parses corpus emails through the full pipeline**

```go
// append to internal/content/parse_test.go

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseBlocksCorpus(t *testing.T) {
	// Use e2e fixtures as stand-in for corpus
	fixtures, err := filepath.Glob("../../e2e/testdata/*.html")
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) == 0 {
		t.Skip("no e2e fixtures found")
	}
	for _, fix := range fixtures {
		t.Run(filepath.Base(fix), func(t *testing.T) {
			raw, err := os.ReadFile(fix)
			if err != nil {
				t.Fatal(err)
			}
			// We can't call filter.CleanHTML yet (circular dep risk),
			// but we can verify ParseBlocks doesn't panic on HTML input.
			// Real integration will be tested in e2e.
			blocks := ParseBlocks(string(raw))
			if len(blocks) == 0 {
				t.Error("expected at least one block from HTML input")
			}
		})
	}
}
```

- [ ] **Step 2: Run the test**

Run: `go test ./internal/content/ -run TestParseBlocksCorpus -v`
Expected: PASS (blocks may not be ideal since input is raw HTML not
cleaned markdown, but parser should not panic)

- [ ] **Step 3: Commit**

```bash
git add internal/content/parse_test.go
git commit -m "Add corpus smoke test for block parser"
```

---

## Phase 2: Theme Refactor

### Task 6: Compiled palette and theme constructor

**Files:**
- Create: `internal/theme/palette.go`
- Create: `internal/theme/palette_test.go`

- [ ] **Step 1: Write failing tests for the new theme system**

```go
// internal/theme/palette_test.go
package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestNewTheme(t *testing.T) {
	th := NewTheme(nordPalette)

	// Palette colors are accessible
	if th.BgBase != lipgloss.Color("#2e3440") {
		t.Errorf("BgBase: got %v, want #2e3440", th.BgBase)
	}
	if th.AccentPrimary != lipgloss.Color("#81a1c1") {
		t.Errorf("AccentPrimary: got %v, want #81a1c1", th.AccentPrimary)
	}
}

func TestNewThemeStyles(t *testing.T) {
	th := NewTheme(nordPalette)

	// HeaderKey should render non-empty (has foreground + bold)
	rendered := th.HeaderKey.Render("From:")
	if rendered == "" {
		t.Error("HeaderKey.Render produced empty string")
	}
	if rendered == "From:" {
		t.Error("HeaderKey.Render produced unstyled string")
	}
}

func TestAllThemesBuild(t *testing.T) {
	themes := map[string]*Theme{
		"Nord":          Nord,
		"SolarizedDark": SolarizedDark,
		"GruvboxDark":   GruvboxDark,
	}
	for name, th := range themes {
		t.Run(name, func(t *testing.T) {
			if th == nil {
				t.Fatal("theme is nil")
			}
			// Every theme should produce styled output
			rendered := th.Heading.Render("Test")
			if rendered == "Test" {
				t.Error("Heading style is unstyled")
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/theme/ -run TestNewTheme -v`
Expected: FAIL — `NewTheme`, `nordPalette` undefined

- [ ] **Step 3: Implement the palette and theme constructor**

```go
// internal/theme/palette.go
package theme

import "github.com/charmbracelet/lipgloss"

// Palette holds the 16 semantic hex color values for a theme.
type Palette struct {
	BgBase      string
	BgElevated  string
	BgSelection string
	BgBorder    string

	FgBase      string
	FgBright    string
	FgBrightest string
	FgDim       string

	AccentPrimary   string
	AccentSecondary string
	AccentTertiary  string

	ColorError   string
	ColorWarning string
	ColorSuccess string
	ColorInfo    string
	ColorSpecial string
}

// Theme holds lipgloss colors and composed styles for rendering.
type Theme struct {
	// Name is the display name of the theme.
	Name string

	// Palette colors as lipgloss values
	BgBase, BgElevated, BgSelection, BgBorder         lipgloss.Color
	FgBase, FgBright, FgBrightest, FgDim               lipgloss.Color
	AccentPrimary, AccentSecondary, AccentTertiary      lipgloss.Color
	ColorError, ColorWarning, ColorSuccess              lipgloss.Color
	ColorInfo, ColorSpecial                             lipgloss.Color

	// Composed styles for content rendering
	HeaderKey     lipgloss.Style
	HeaderValue   lipgloss.Style
	HeaderDim     lipgloss.Style
	Paragraph     lipgloss.Style
	Heading       lipgloss.Style
	Quote         lipgloss.Style
	DeepQuote     lipgloss.Style
	Attribution   lipgloss.Style
	Signature     lipgloss.Style
	Bold          lipgloss.Style
	Italic        lipgloss.Style
	Link          lipgloss.Style
	CodeInline    lipgloss.Style
	CodeBlock     lipgloss.Style
	HorizontalRule lipgloss.Style
}

// NewTheme creates a Theme from a Palette, building all composed styles.
func NewTheme(name string, p Palette) *Theme {
	t := &Theme{
		Name: name,

		BgBase:          lipgloss.Color(p.BgBase),
		BgElevated:      lipgloss.Color(p.BgElevated),
		BgSelection:     lipgloss.Color(p.BgSelection),
		BgBorder:        lipgloss.Color(p.BgBorder),
		FgBase:          lipgloss.Color(p.FgBase),
		FgBright:        lipgloss.Color(p.FgBright),
		FgBrightest:     lipgloss.Color(p.FgBrightest),
		FgDim:           lipgloss.Color(p.FgDim),
		AccentPrimary:   lipgloss.Color(p.AccentPrimary),
		AccentSecondary: lipgloss.Color(p.AccentSecondary),
		AccentTertiary:  lipgloss.Color(p.AccentTertiary),
		ColorError:      lipgloss.Color(p.ColorError),
		ColorWarning:    lipgloss.Color(p.ColorWarning),
		ColorSuccess:    lipgloss.Color(p.ColorSuccess),
		ColorInfo:       lipgloss.Color(p.ColorInfo),
		ColorSpecial:    lipgloss.Color(p.ColorSpecial),
	}

	t.HeaderKey = lipgloss.NewStyle().
		Foreground(t.AccentPrimary).Bold(true)
	t.HeaderValue = lipgloss.NewStyle().
		Foreground(t.FgBase)
	t.HeaderDim = lipgloss.NewStyle().
		Foreground(t.FgDim)
	t.Paragraph = lipgloss.NewStyle().
		Foreground(t.FgBase)
	t.Heading = lipgloss.NewStyle().
		Foreground(t.ColorSuccess).Bold(true)
	t.Quote = lipgloss.NewStyle().
		Foreground(t.AccentTertiary)
	t.DeepQuote = lipgloss.NewStyle().
		Foreground(t.FgDim)
	t.Attribution = lipgloss.NewStyle().
		Foreground(t.FgDim).Italic(true)
	t.Signature = lipgloss.NewStyle().
		Foreground(t.FgDim)
	t.Bold = lipgloss.NewStyle().Bold(true)
	t.Italic = lipgloss.NewStyle().Italic(true)
	t.Link = lipgloss.NewStyle().
		Foreground(t.AccentPrimary).Underline(true)
	t.CodeInline = lipgloss.NewStyle().
		Foreground(t.FgBright)
	t.CodeBlock = lipgloss.NewStyle().
		Foreground(t.FgBright)
	t.HorizontalRule = lipgloss.NewStyle().
		Foreground(t.FgDim)

	return t
}

// PaletteHex returns the raw hex value for a color slot by name.
// Used by the styleset generator.
func (t *Theme) PaletteHex(name string) string {
	switch name {
	case "bg_base":
		return string(t.BgBase)
	case "bg_elevated":
		return string(t.BgElevated)
	case "bg_selection":
		return string(t.BgSelection)
	case "bg_border":
		return string(t.BgBorder)
	case "fg_base":
		return string(t.FgBase)
	case "fg_bright":
		return string(t.FgBright)
	case "fg_brightest":
		return string(t.FgBrightest)
	case "fg_dim":
		return string(t.FgDim)
	case "accent_primary":
		return string(t.AccentPrimary)
	case "accent_secondary":
		return string(t.AccentSecondary)
	case "accent_tertiary":
		return string(t.AccentTertiary)
	case "color_error":
		return string(t.ColorError)
	case "color_warning":
		return string(t.ColorWarning)
	case "color_success":
		return string(t.ColorSuccess)
	case "color_info":
		return string(t.ColorInfo)
	case "color_special":
		return string(t.ColorSpecial)
	default:
		return ""
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/theme/ -run TestNewTheme -v`
Expected: FAIL — `nordPalette` and `Nord` not yet defined (themes.go)

- [ ] **Step 5: Create compiled theme values**

```go
// internal/theme/themes.go
package theme

var nordPalette = Palette{
	BgBase:          "#2e3440",
	BgElevated:      "#3b4252",
	BgSelection:     "#394353",
	BgBorder:        "#49576b",
	FgBase:          "#d8dee9",
	FgBright:        "#e5e9f0",
	FgBrightest:     "#eceff4",
	FgDim:           "#616e88",
	AccentPrimary:   "#81a1c1",
	AccentSecondary: "#88c0d0",
	AccentTertiary:  "#8fbcbb",
	ColorError:      "#bf616a",
	ColorWarning:    "#d08770",
	ColorSuccess:    "#a3be8c",
	ColorInfo:       "#ebcb8b",
	ColorSpecial:    "#b48ead",
}

var solarizedDarkPalette = Palette{
	BgBase:          "#002b36",
	BgElevated:      "#073642",
	BgSelection:     "#073642",
	BgBorder:        "#586e75",
	FgBase:          "#839496",
	FgBright:        "#93a1a1",
	FgBrightest:     "#eee8d5",
	FgDim:           "#657b83",
	AccentPrimary:   "#268bd2",
	AccentSecondary: "#2aa198",
	AccentTertiary:  "#2aa198",
	ColorError:      "#dc322f",
	ColorWarning:    "#cb4b16",
	ColorSuccess:    "#859900",
	ColorInfo:       "#b58900",
	ColorSpecial:    "#6c71c4",
}

var gruvboxDarkPalette = Palette{
	BgBase:          "#282828",
	BgElevated:      "#3c3836",
	BgSelection:     "#3c3836",
	BgBorder:        "#665c54",
	FgBase:          "#ebdbb2",
	FgBright:        "#fbf1c7",
	FgBrightest:     "#fbf1c7",
	FgDim:           "#928374",
	AccentPrimary:   "#83a598",
	AccentSecondary: "#8ec07c",
	AccentTertiary:  "#8ec07c",
	ColorError:      "#fb4934",
	ColorWarning:    "#fe8019",
	ColorSuccess:    "#b8bb26",
	ColorInfo:       "#fabd2f",
	ColorSpecial:    "#d3869b",
}

// Nord is the compiled Nord theme.
var Nord = NewTheme("Nord", nordPalette)

// SolarizedDark is the compiled Solarized Dark theme.
var SolarizedDark = NewTheme("Solarized Dark", solarizedDarkPalette)

// GruvboxDark is the compiled Gruvbox Dark theme.
var GruvboxDark = NewTheme("Gruvbox Dark", gruvboxDarkPalette)
```

- [ ] **Step 6: Run all new theme tests**

Run: `go test ./internal/theme/ -run "TestNewTheme|TestAllThemes" -v`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add internal/theme/palette.go internal/theme/themes.go internal/theme/palette_test.go
git commit -m "Add compiled lipgloss themes with palette constructor"
```

---

### Task 7: Update styleset generator for new Theme type

**Files:**
- Modify: `internal/theme/styleset.go`
- Modify: `internal/theme/styleset_test.go`

- [ ] **Step 1: Update styleset.go to use PaletteHex instead of old Color method**

The existing `stylesetData` wrapper has a `C(name string) string`
method that calls `t.theme.Color(name)`. Update it to call
`t.theme.PaletteHex(name)` instead. The template itself does not
change — it already uses `{{.C "slot_name"}}`.

In `internal/theme/styleset.go`, change the `stylesetData` type
and its method:

```go
type stylesetData struct {
	Name  string
	theme *Theme
}

func (d stylesetData) C(name string) string {
	return d.theme.PaletteHex(name)
}
```

Update `GenerateStyleset` to accept `*Theme` (the new type — same
pointer receiver, but the struct has changed):

```go
func GenerateStyleset(t *Theme) (string, error) {
	var buf bytes.Buffer
	data := stylesetData{Name: t.Name, theme: t}
	if err := stylesetTmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute styleset template: %w", err)
	}
	return buf.String(), nil
}

func WriteStyleset(t *Theme, path string) error {
	content, err := GenerateStyleset(t)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}
```

- [ ] **Step 2: Update styleset_test.go to use compiled themes**

```go
// internal/theme/styleset_test.go
package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateStyleset(t *testing.T) {
	content, err := GenerateStyleset(Nord)
	if err != nil {
		t.Fatal(err)
	}
	checks := []string{
		"title.bg=#81a1c1",
		"title.fg=#2e3440",
		"error.fg=#bf616a",
		"warning.fg=#d08770",
		"success.fg=#a3be8c",
		"msglist_unread.fg=#8fbcbb",
		"tab.selected.bg=#88c0d0",
		"*.selected.bg=#394353",
		"quote_1.fg=#8fbcbb",
		"diff_add.fg=#a3be8c",
		"diff_del.fg=#bf616a",
		"border.fg=#49576b",
		"[viewer]",
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("missing %q in generated styleset", check)
		}
	}
}

func TestGenerateStylesetWriteFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "TestTheme")
	if err := WriteStyleset(Nord, path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "title.bg=#81a1c1") {
		t.Error("written file missing expected content")
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/theme/ -run TestGenerateStyleset -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/theme/styleset.go internal/theme/styleset_test.go
git commit -m "Update styleset generator to use compiled theme palettes"
```

---

### Task 8: Delete old theme system

**Files:**
- Delete: `internal/theme/glamour.go`
- Delete: `internal/theme/glamour_test.go`
- Modify: `internal/theme/theme.go` — delete old `Theme` struct, `Load`, `FindPath`, `FindConfigDir`, `ANSI`, `Color`, `Raw`, `Reset`, `GlamourStyle` and all TOML parsing. Keep only the package declaration.
- Modify: `internal/theme/theme_test.go` — delete all old tests

- [ ] **Step 1: Delete glamour files**

```bash
rm internal/theme/glamour.go internal/theme/glamour_test.go
```

- [ ] **Step 2: Gut theme.go**

Replace `internal/theme/theme.go` with just a package declaration
and a comment pointing to the new files:

```go
// Package theme provides compiled color themes for mailrender and poplar.
//
// Themes are defined as Go values in themes.go. Each theme is built
// from a Palette (16 hex colors) via NewTheme, which constructs all
// lipgloss styles. There is no runtime file loading.
//
// The styleset generator (styleset.go) writes aerc stylesets from
// the same palette values.
package theme
```

- [ ] **Step 3: Replace theme_test.go with tests for compiled themes**

```go
// internal/theme/theme_test.go
package theme

import "testing"

func TestPaletteHex(t *testing.T) {
	tests := []struct {
		name string
		slot string
		want string
	}{
		{"bg_base", "bg_base", "#2e3440"},
		{"accent_primary", "accent_primary", "#81a1c1"},
		{"unknown", "nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Nord.PaletteHex(tt.slot)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAllThemesHaveDistinctColors(t *testing.T) {
	// Sanity check that themes have different palettes
	if Nord.PaletteHex("bg_base") == SolarizedDark.PaletteHex("bg_base") {
		t.Error("Nord and SolarizedDark have same bg_base")
	}
	if Nord.PaletteHex("bg_base") == GruvboxDark.PaletteHex("bg_base") {
		t.Error("Nord and GruvboxDark have same bg_base")
	}
}
```

- [ ] **Step 4: Run all theme tests**

Run: `go test ./internal/theme/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add -u internal/theme/
git commit -m "Remove TOML theme loader and glamour bridge"
```

---

## Phase 3: Lipgloss Renderer

### Task 9: Body renderer

**Files:**
- Create: `internal/content/render.go`
- Create: `internal/content/render_test.go`

- [ ] **Step 1: Write failing tests for body rendering**

```go
// internal/content/render_test.go
package content

import (
	"strings"
	"testing"

	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestRenderBodyParagraph(t *testing.T) {
	blocks := []Block{
		Paragraph{Spans: []Span{Text{Content: "Hello world"}}},
	}
	result := RenderBody(blocks, theme.Nord, 80)
	if !strings.Contains(result, "Hello world") {
		t.Errorf("expected 'Hello world' in output, got %q", result)
	}
}

func TestRenderBodyHeading(t *testing.T) {
	blocks := []Block{
		Heading{Spans: []Span{Text{Content: "Title"}}, Level: 1},
	}
	result := RenderBody(blocks, theme.Nord, 80)
	if !strings.Contains(result, "Title") {
		t.Errorf("expected 'Title' in output, got %q", result)
	}
	// Heading should be styled (not plain text)
	if result == "Title\n" {
		t.Error("heading appears unstyled")
	}
}

func TestRenderBodySignature(t *testing.T) {
	blocks := []Block{
		Paragraph{Spans: []Span{Text{Content: "Body text"}}},
		Signature{Lines: [][]Span{
			{Text{Content: "-- "}},
			{Text{Content: "Geoff"}},
		}},
	}
	result := RenderBody(blocks, theme.Nord, 80)
	if !strings.Contains(result, "Body text") {
		t.Error("missing body text")
	}
	if !strings.Contains(result, "Geoff") {
		t.Error("missing signature")
	}
}

func TestRenderBodyWidthCap(t *testing.T) {
	long := strings.Repeat("word ", 30) // ~150 chars
	blocks := []Block{
		Paragraph{Spans: []Span{Text{Content: long}}},
	}
	result := RenderBody(blocks, theme.Nord, 120)
	// Body should cap at 78 even though terminal is 120
	for _, line := range strings.Split(result, "\n") {
		// Strip ANSI for length check
		visible := stripANSITest(line)
		if len(visible) > 80 { // small margin for style edge cases
			t.Errorf("line exceeds 78 visible chars: %d chars: %q", len(visible), visible)
		}
	}
}

func TestRenderBodyNarrowTerminal(t *testing.T) {
	long := strings.Repeat("word ", 20)
	blocks := []Block{
		Paragraph{Spans: []Span{Text{Content: long}}},
	}
	result := RenderBody(blocks, theme.Nord, 40)
	for _, line := range strings.Split(result, "\n") {
		visible := stripANSITest(line)
		if len(visible) > 42 {
			t.Errorf("line exceeds terminal width: %d chars", len(visible))
		}
	}
}

func TestRenderBodyBoldSpan(t *testing.T) {
	blocks := []Block{
		Paragraph{Spans: []Span{
			Text{Content: "hello "},
			Bold{Content: "bold"},
			Text{Content: " world"},
		}},
	}
	result := RenderBody(blocks, theme.Nord, 80)
	if !strings.Contains(result, "bold") {
		t.Error("missing bold text")
	}
}

func TestRenderBodyLink(t *testing.T) {
	blocks := []Block{
		Paragraph{Spans: []Span{
			Link{Text: "click", URL: "https://example.com"},
		}},
	}
	result := RenderBody(blocks, theme.Nord, 80)
	if !strings.Contains(result, "click") {
		t.Error("missing link text")
	}
}

func TestRenderBodyRule(t *testing.T) {
	blocks := []Block{
		Paragraph{Spans: []Span{Text{Content: "above"}}},
		Rule{},
		Paragraph{Spans: []Span{Text{Content: "below"}}},
	}
	result := RenderBody(blocks, theme.Nord, 80)
	if !strings.Contains(result, "above") || !strings.Contains(result, "below") {
		t.Error("missing content around rule")
	}
}

// stripANSITest removes ANSI escape sequences for visible length checks.
func stripANSITest(s string) string {
	// Simple CSI stripper for test assertions
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// Skip CSI sequence
			i++
			if i < len(s) && s[i] == '[' {
				i++
				for i < len(s) && s[i] < 0x40 {
					i++
				}
				if i < len(s) {
					i++ // skip final byte
				}
			}
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/content/ -run TestRenderBody -v`
Expected: FAIL — `RenderBody` undefined

- [ ] **Step 3: Implement the body renderer**

```go
// internal/content/render.go
package content

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

const maxBodyWidth = 78

// RenderBody renders blocks into a styled string using lipgloss.
// Width is capped at maxBodyWidth for readability.
func RenderBody(blocks []Block, t *theme.Theme, width int) string {
	w := width
	if w > maxBodyWidth {
		w = maxBodyWidth
	}

	var sections []string
	for _, block := range blocks {
		sections = append(sections, renderBlock(block, t, w))
	}
	return strings.Join(sections, "\n")
}

func renderBlock(block Block, t *theme.Theme, width int) string {
	switch b := block.(type) {
	case Paragraph:
		text := renderSpans(b.Spans, t)
		return t.Paragraph.Width(width).Render(text)

	case Heading:
		text := renderSpans(b.Spans, t)
		prefix := strings.Repeat("#", b.Level) + " "
		return t.Heading.Render(prefix + text)

	case Blockquote:
		style := t.Quote
		if b.Level > 1 {
			style = t.DeepQuote
		}
		prefix := strings.Repeat("> ", b.Level)
		var inner []string
		for _, child := range b.Blocks {
			inner = append(inner, renderBlock(child, t, width-len(prefix)))
		}
		content := strings.Join(inner, "\n")
		// Prepend quote prefix to each line
		var lines []string
		for _, line := range strings.Split(content, "\n") {
			lines = append(lines, style.Render(prefix)+line)
		}
		return strings.Join(lines, "\n")

	case QuoteAttribution:
		text := renderSpans(b.Spans, t)
		return t.Attribution.Render(text)

	case Signature:
		var lines []string
		for _, spans := range b.Lines {
			text := renderSpans(spans, t)
			lines = append(lines, t.Signature.Render(text))
		}
		return strings.Join(lines, "\n")

	case Rule:
		line := strings.Repeat("─", width)
		return t.HorizontalRule.Render(line)

	case CodeBlock:
		return t.CodeBlock.Width(width).Render(b.Text)

	case Table:
		return renderTable(b, t, width)

	case ListItem:
		text := renderSpans(b.Spans, t)
		prefix := "- "
		if b.Ordered {
			prefix = strings.Repeat(" ", 2) // indent for alignment
			prefix = string(rune('0'+b.Index%10)) + ". "
		}
		return t.Paragraph.Render(prefix + text)

	default:
		return ""
	}
}

func renderSpans(spans []Span, t *theme.Theme) string {
	var parts []string
	for _, span := range spans {
		switch s := span.(type) {
		case Text:
			parts = append(parts, s.Content)
		case Bold:
			parts = append(parts, t.Bold.Render(s.Content))
		case Italic:
			parts = append(parts, t.Italic.Render(s.Content))
		case Code:
			parts = append(parts, t.CodeInline.Render(s.Content))
		case Link:
			parts = append(parts, t.Link.Render(s.Text))
		}
	}
	return strings.Join(parts, "")
}

func renderTable(table Table, t *theme.Theme, width int) string {
	// Simple pipe table rendering
	var rows []string

	if len(table.Headers) > 0 {
		var cells []string
		for _, hdr := range table.Headers {
			cells = append(cells, t.Bold.Render(renderSpans(hdr, t)))
		}
		rows = append(rows, strings.Join(cells, " | "))
		// Separator
		var seps []string
		for range table.Headers {
			seps = append(seps, "---")
		}
		rows = append(rows, strings.Join(seps, " | "))
	}

	for _, row := range table.Rows {
		var cells []string
		for _, cell := range row {
			cells = append(cells, renderSpans(cell, t))
		}
		rows = append(rows, strings.Join(cells, " | "))
	}

	return strings.Join(rows, "\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/content/ -run TestRenderBody -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/content/render.go internal/content/render_test.go
git commit -m "Add lipgloss body renderer for block model"
```

---

### Task 10: Header renderer

**Files:**
- Modify: `internal/content/render.go`
- Modify: `internal/content/render_test.go`

- [ ] **Step 1: Write failing tests for header rendering**

```go
// append to internal/content/render_test.go

func TestRenderHeaders(t *testing.T) {
	h := ParsedHeaders{
		From:    []Address{{Name: "Alice", Email: "alice@example.com"}},
		To:      []Address{{Name: "Bob", Email: "bob@example.com"}},
		Date:    "Mon, 5 Jan 2026",
		Subject: "Hello World",
	}
	result := RenderHeaders(h, theme.Nord, 80)
	if !strings.Contains(result, "From:") {
		t.Error("missing From header")
	}
	if !strings.Contains(result, "Alice") {
		t.Error("missing From name")
	}
	if !strings.Contains(result, "Subject:") {
		t.Error("missing Subject header")
	}
	if !strings.Contains(result, "Hello World") {
		t.Error("missing Subject value")
	}
}

func TestRenderHeadersOrder(t *testing.T) {
	h := ParsedHeaders{
		From:    []Address{{Name: "Alice", Email: "alice@example.com"}},
		To:      []Address{{Name: "Bob", Email: "bob@example.com"}},
		Date:    "Mon, 5 Jan 2026",
		Subject: "Test",
	}
	result := RenderHeaders(h, theme.Nord, 80)
	fromIdx := strings.Index(result, "From:")
	toIdx := strings.Index(result, "To:")
	dateIdx := strings.Index(result, "Date:")
	subjectIdx := strings.Index(result, "Subject:")

	if fromIdx > toIdx {
		t.Error("From should appear before To")
	}
	if toIdx > dateIdx {
		t.Error("To should appear before Date")
	}
	if dateIdx > subjectIdx {
		t.Error("Date should appear before Subject")
	}
}

func TestRenderHeadersSkipsEmpty(t *testing.T) {
	h := ParsedHeaders{
		From:    []Address{{Name: "Alice", Email: "alice@example.com"}},
		Subject: "Test",
	}
	result := RenderHeaders(h, theme.Nord, 80)
	if strings.Contains(result, "To:") {
		t.Error("should not render empty To header")
	}
	if strings.Contains(result, "Cc:") {
		t.Error("should not render empty Cc header")
	}
}

func TestRenderHeadersSeparator(t *testing.T) {
	h := ParsedHeaders{
		From:    []Address{{Email: "alice@example.com"}},
		Subject: "Test",
	}
	result := RenderHeaders(h, theme.Nord, 80)
	if !strings.Contains(result, "─") {
		t.Error("missing separator line")
	}
}

func TestRenderHeadersAddressWrap(t *testing.T) {
	var addrs []Address
	for i := 0; i < 10; i++ {
		addrs = append(addrs, Address{
			Name:  "Recipient Name",
			Email: "recipient@example.com",
		})
	}
	h := ParsedHeaders{
		From:    []Address{{Email: "sender@example.com"}},
		To:      addrs,
		Subject: "Test",
	}
	result := RenderHeaders(h, theme.Nord, 80)
	// Should wrap — multiple lines for To
	lines := strings.Split(result, "\n")
	toLines := 0
	for _, line := range lines {
		visible := stripANSITest(line)
		if strings.Contains(visible, "Recipient") {
			toLines++
		}
	}
	if toLines < 2 {
		t.Errorf("expected To: to wrap across multiple lines, got %d", toLines)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/content/ -run TestRenderHeaders -v`
Expected: FAIL — `RenderHeaders` undefined

- [ ] **Step 3: Implement the header renderer**

Add to `internal/content/render.go`:

```go
import "fmt"

// RenderHeaders renders parsed headers into a styled string.
// Headers use the full terminal width for address wrapping.
func RenderHeaders(h ParsedHeaders, t *theme.Theme, width int) string {
	var lines []string

	// Fixed order: From, To, Cc, Bcc, Date, Subject
	if len(h.From) > 0 {
		lines = append(lines, renderHeaderAddresses("From", h.From, t, width)...)
	}
	if len(h.To) > 0 {
		lines = append(lines, renderHeaderAddresses("To", h.To, t, width)...)
	}
	if len(h.Cc) > 0 {
		lines = append(lines, renderHeaderAddresses("Cc", h.Cc, t, width)...)
	}
	if len(h.Bcc) > 0 {
		lines = append(lines, renderHeaderAddresses("Bcc", h.Bcc, t, width)...)
	}
	if h.Date != "" {
		lines = append(lines, renderHeaderScalar("Date", h.Date, t))
	}
	if h.Subject != "" {
		lines = append(lines, renderHeaderScalar("Subject", h.Subject, t))
	}

	// Separator
	sep := t.HeaderDim.Render(strings.Repeat("─", width))
	lines = append(lines, sep)

	return strings.Join(lines, "\n")
}

func renderHeaderScalar(key, value string, t *theme.Theme) string {
	return t.HeaderKey.Render(key+":") + " " + t.HeaderValue.Render(value)
}

func renderHeaderAddresses(key string, addrs []Address, t *theme.Theme, width int) []string {
	keyStr := t.HeaderKey.Render(key + ":")
	indent := strings.Repeat(" ", len(key)+2) // "Key: " alignment

	var formatted []string
	for _, a := range addrs {
		if a.Name != "" {
			formatted = append(formatted, fmt.Sprintf("%s %s",
				t.HeaderValue.Render(a.Name),
				t.HeaderDim.Render("<"+a.Email+">")))
		} else {
			formatted = append(formatted, t.HeaderValue.Render(a.Email))
		}
	}

	// Build lines with wrapping
	var lines []string
	current := keyStr + " "
	currentVisible := len(key) + 2

	for i, addr := range formatted {
		addrVisible := len(addrs[i].Name) + len(addrs[i].Email) + 3 // "Name <email>"
		if addrs[i].Name == "" {
			addrVisible = len(addrs[i].Email)
		}

		sep := ""
		sepLen := 0
		if i > 0 {
			sep = ", "
			sepLen = 2
		}

		if currentVisible+sepLen+addrVisible > width && i > 0 {
			lines = append(lines, current)
			current = indent + addr
			currentVisible = len(indent) + addrVisible
		} else {
			current += sep + addr
			currentVisible += sepLen + addrVisible
		}
	}
	lines = append(lines, current)

	return lines
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/content/ -run TestRenderHeaders -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/content/render.go internal/content/render_test.go
git commit -m "Add lipgloss header renderer with address wrapping"
```

---

## Phase 4: Wire Up CLI Filters

### Task 11: Refactor filter package to export CleanHTML and CleanPlain

**Files:**
- Modify: `internal/filter/html.go`
- Modify: `internal/filter/plain.go`
- Modify: `internal/filter/html_test.go`
- Modify: `internal/filter/plain_test.go`

- [ ] **Step 1: Add CleanHTML export**

Add to `internal/filter/html.go`:

```go
// CleanHTML converts raw HTML email content to normalized markdown.
// This is the content pipeline only — no rendering or styling.
func CleanHTML(html string) string {
	html = prepareHTML(html)
	md, err := convertHTML(html)
	if err != nil {
		return html // fallback: return cleaned HTML as-is
	}
	md = normalizeWhitespace(md)
	md = deduplicateBlocks(md)
	md = stripEmptyLinks(md)
	md = unflattenQuotes(md)
	return md
}
```

Note: `collapseShortBlocks`, `compactLineRuns`, and
`reflowMarkdown` are deliberately excluded — those are rendering
decisions that the block renderer handles.

- [ ] **Step 2: Add CleanPlain export**

Add to `internal/filter/plain.go`:

```go
// CleanPlain normalizes plain text email content to markdown.
// Detects HTML-in-plain-text and routes through CleanHTML if found.
func CleanPlain(text string) string {
	if detectHTML(text) {
		return CleanHTML(text)
	}
	text = html.UnescapeString(text)
	return text
}
```

Add `"html"` to the imports if not already present.

- [ ] **Step 3: Write tests for the new exports**

```go
// append to internal/filter/html_test.go

func TestCleanHTML(t *testing.T) {
	input := "<div><p>Hello <strong>world</strong></p></div>"
	result := CleanHTML(input)
	if !strings.Contains(result, "Hello") {
		t.Error("expected 'Hello' in output")
	}
	if !strings.Contains(result, "**world**") {
		t.Error("expected bold markdown in output")
	}
	// Should NOT contain ANSI escapes
	if strings.Contains(result, "\033") {
		t.Error("CleanHTML should not produce ANSI output")
	}
}

func TestCleanHTMLStripsTracking(t *testing.T) {
	input := `<div><p>Content</p><img width="1" height="1" src="track.gif"></div>`
	result := CleanHTML(input)
	if strings.Contains(result, "track.gif") {
		t.Error("tracking image should be stripped")
	}
}
```

```go
// append to internal/filter/plain_test.go

func TestCleanPlain(t *testing.T) {
	result := CleanPlain("Hello world &amp; goodbye")
	if result != "Hello world & goodbye" {
		t.Errorf("got %q", result)
	}
}

func TestCleanPlainHTMLDetection(t *testing.T) {
	input := "<html><body><p>Hello</p></body></html>"
	result := CleanPlain(input)
	if strings.Contains(result, "<html>") {
		t.Error("HTML-in-plain should be cleaned")
	}
	if !strings.Contains(result, "Hello") {
		t.Error("content should be preserved")
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/filter/ -run "TestCleanHTML|TestCleanPlain" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/filter/html.go internal/filter/plain.go internal/filter/html_test.go internal/filter/plain_test.go
git commit -m "Export CleanHTML and CleanPlain content pipeline functions"
```

---

### Task 12: Wire cmd/mailrender to new pipeline

**Files:**
- Modify: `cmd/mailrender/html.go`
- Modify: `cmd/mailrender/plain.go`
- Modify: `cmd/mailrender/headers.go`
- Modify: `cmd/mailrender/markdown.go`
- Modify: `cmd/mailrender/tohtml.go`
- Modify: `cmd/mailrender/themes.go`
- Modify: `cmd/mailrender/root.go`

- [ ] **Step 1: Update html.go to use new pipeline**

```go
// cmd/mailrender/html.go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/glw907/beautiful-aerc/internal/content"
	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/glw907/beautiful-aerc/internal/theme"
	"github.com/spf13/cobra"
)

func newHTMLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "html",
		Short: "Render HTML email to styled terminal output",
		RunE: func(cmd *cobra.Command, args []string) error {
			cols := termCols()
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			md := filter.CleanHTML(string(raw))
			blocks := content.ParseBlocks(md)
			result := content.RenderBody(blocks, selectedTheme(), cols)
			fmt.Fprint(os.Stdout, "\n"+result)
			return nil
		},
	}
}
```

- [ ] **Step 2: Update plain.go**

```go
// cmd/mailrender/plain.go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/glw907/beautiful-aerc/internal/content"
	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/glw907/beautiful-aerc/internal/theme"
	"github.com/spf13/cobra"
)

func newPlainCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "plain",
		Short: "Render plain text email to styled terminal output",
		RunE: func(cmd *cobra.Command, args []string) error {
			cols := termCols()
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			md := filter.CleanPlain(string(raw))
			blocks := content.ParseBlocks(md)
			result := content.RenderBody(blocks, selectedTheme(), cols)
			fmt.Fprintln(os.Stdout)
			fmt.Fprint(os.Stdout, result)
			return nil
		},
	}
}
```

- [ ] **Step 3: Update headers.go — remove loadTheme/colorsFromTheme, add selectedTheme/termCols**

```go
// cmd/mailrender/headers.go
package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/glw907/beautiful-aerc/internal/content"
	"github.com/glw907/beautiful-aerc/internal/theme"
	"github.com/spf13/cobra"
)

func newHeadersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "headers",
		Short: "Render email headers with styling",
		RunE: func(cmd *cobra.Command, args []string) error {
			cols := termCols()
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			h := content.ParseHeaders(string(raw))
			result := content.RenderHeaders(h, selectedTheme(), cols)
			fmt.Fprint(os.Stdout, result)
			return nil
		},
	}
}

// termCols reads the terminal width from AERC_COLUMNS, defaulting to 80.
func termCols() int {
	if s := os.Getenv("AERC_COLUMNS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 80
}

// selectedTheme returns the active compiled theme.
// For now, always Nord. A --theme flag can be added later.
func selectedTheme() *theme.Theme {
	return theme.Nord
}
```

- [ ] **Step 4: Update markdown.go**

```go
// cmd/mailrender/markdown.go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/spf13/cobra"
)

func newMarkdownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "markdown",
		Short: "Convert HTML email to clean markdown",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read stdin: %w", err)
			}
			md := filter.CleanHTML(string(raw))
			fmt.Fprintln(os.Stdout, md)
			return nil
		},
	}
}
```

- [ ] **Step 5: Update themes.go to use compiled themes**

```go
// cmd/mailrender/themes.go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/beautiful-aerc/internal/theme"
	"github.com/spf13/cobra"
)

func newThemesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "themes",
		Short: "Theme management commands",
	}
	cmd.AddCommand(newThemesGenerateCmd())
	return cmd
}

func newThemesGenerateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate [theme-name]",
		Short: "Generate aerc styleset from a compiled theme",
		Long:  "Available themes: nord, solarized-dark, gruvbox-dark. Generates all if no name given.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, err := findConfigDir()
			if err != nil {
				return err
			}
			stylesetsDir := filepath.Join(configDir, "stylesets")
			if err := os.MkdirAll(stylesetsDir, 0755); err != nil {
				return fmt.Errorf("create stylesets dir: %w", err)
			}

			themes := map[string]*theme.Theme{
				"nord":           theme.Nord,
				"solarized-dark": theme.SolarizedDark,
				"gruvbox-dark":   theme.GruvboxDark,
			}

			if len(args) > 0 {
				t, ok := themes[args[0]]
				if !ok {
					return fmt.Errorf("unknown theme %q (available: nord, solarized-dark, gruvbox-dark)", args[0])
				}
				return generateOne(t, stylesetsDir)
			}

			for _, t := range []*theme.Theme{theme.Nord, theme.SolarizedDark, theme.GruvboxDark} {
				if err := generateOne(t, stylesetsDir); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func generateOne(t *theme.Theme, dir string) error {
	outPath := filepath.Join(dir, t.Name)
	if err := theme.WriteStyleset(t, outPath); err != nil {
		return fmt.Errorf("generate %s: %w", t.Name, err)
	}
	fmt.Fprintf(os.Stderr, "Theme: %s\nStyleset: %s\n", t.Name, outPath)
	return nil
}

func findConfigDir() (string, error) {
	if dir := os.Getenv("AERC_CONFIG"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home dir: %w", err)
	}
	return filepath.Join(home, ".config", "aerc"), nil
}
```

- [ ] **Step 6: Add preview subcommand to root.go**

```go
// cmd/mailrender/preview.go
package main

import (
	"fmt"
	"os"

	"github.com/glw907/beautiful-aerc/internal/content"
	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/spf13/cobra"
)

func newPreviewCmd() *cobra.Command {
	var themeName string
	var width int

	cmd := &cobra.Command{
		Use:   "preview <file>",
		Short: "Preview email rendering from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}
			t := resolveTheme(themeName)
			if width <= 0 {
				width = 78
			}
			md := filter.CleanHTML(string(raw))
			blocks := content.ParseBlocks(md)
			result := content.RenderBody(blocks, t, width)
			fmt.Print(result)
			return nil
		},
	}
	cmd.Flags().StringVar(&themeName, "theme", "nord", "theme name (nord, solarized-dark, gruvbox-dark)")
	cmd.Flags().IntVar(&width, "width", 78, "rendering width in columns")
	return cmd
}

func resolveTheme(name string) *theme.Theme {
	switch name {
	case "solarized-dark":
		return theme.SolarizedDark
	case "gruvbox-dark":
		return theme.GruvboxDark
	default:
		return theme.Nord
	}
}
```

Add `preview` import and register in `root.go`:

```go
// In newRootCmd(), add:
cmd.AddCommand(newPreviewCmd())
```

- [ ] **Step 7: Build and verify**

Run: `go build ./cmd/mailrender/`
Expected: success

- [ ] **Step 8: Run make check**

Run: `make check`
Expected: vet passes. Some existing tests may fail due to
theme API changes — fix in next task.

- [ ] **Step 9: Commit**

```bash
git add cmd/mailrender/
git commit -m "Wire CLI filters to content pipeline and lipgloss renderer"
```

---

### Task 13: Update e2e tests and golden files

**Files:**
- Modify: `e2e/e2e_test.go`

- [ ] **Step 1: Update TestMain setup**

The e2e tests currently write a TOML theme to a temp dir and set
`AERC_CONFIG`. With compiled themes, the binary no longer reads
TOML files. Remove the theme file setup — the binary uses the
compiled Nord theme by default.

Update `TestMain` in `e2e/e2e_test.go`:
- Remove the temp theme TOML file creation
- Remove `AERC_CONFIG` env var setup (unless still needed for
  `styleset-name`)
- Keep the binary build step

- [ ] **Step 2: Regenerate golden files**

Run: `go test ./e2e/ -update-golden`
Expected: golden files regenerated with lipgloss output instead
of glamour output

- [ ] **Step 3: Run e2e tests without update flag**

Run: `go test ./e2e/ -v`
Expected: PASS with new golden files

- [ ] **Step 4: Manually inspect golden file output**

Read at least one golden file to verify the lipgloss rendering
looks reasonable:

Run: `cat e2e/testdata/golden/simple.txt`
Expected: styled email output with proper formatting

- [ ] **Step 5: Commit**

```bash
git add e2e/
git commit -m "Update e2e golden tests for lipgloss renderer"
```

---

## Phase 5: Cleanup

### Task 14: Remove glamour dependency and dead code

**Files:**
- Modify: `go.mod`
- Modify: `internal/filter/html.go` — delete `collapseShortBlocks`,
  `compactLineRuns`, `reflowMarkdown`, `reflowParagraph`,
  `reflowBlockquote`, `markdownTokens`, `isParagraph`,
  `isShortPlain`, `isCompactLine`, `mergeBlockRuns`, `blockKey`,
  `stripANSI`, and the old `HTML` and `Markdown` exports. Remove
  glamour imports.
- Modify: `internal/filter/html_test.go` — delete tests for
  removed functions: `TestReflowParagraph`,
  `TestMarkdownTokensKeepsLinksAtomic`,
  `TestReflowKeepsLinkTextTogether`,
  `TestReflowMarkdownPreservesNonParagraphs`,
  `TestCollapseShortBlocks`, `TestCompactLineRuns`,
  `TestIsParagraphSkipsHardBreaks`, `TestStripANSI`
- Modify: `internal/filter/headers.go` — delete `ColorSet` type,
  `Headers` export, `colorizeValue`, `wrapAddresses` (these are
  now in `internal/content/`)
- Modify: `internal/filter/headers_test.go` — delete tests for
  removed functions
- Modify: `internal/filter/plain.go` — delete old `Plain` export,
  `findColorize`, subprocess code
- Delete: `.config/aerc/themes/nord.toml`,
  `.config/aerc/themes/solarized-dark.toml`,
  `.config/aerc/themes/gruvbox-dark.toml`

- [ ] **Step 1: Remove deleted rendering functions from html.go**

Delete: `collapseShortBlocks`, `isShortPlain`, `compactLineRuns`,
`isCompactLine`, `mergeBlockRuns`, `blockKey`, `reflowMarkdown`,
`reflowParagraph`, `reflowBlockquote`, `isParagraph`,
`markdownTokens`, `stripANSI`. Delete the old `HTML` and `Markdown`
functions. Remove `glamour` and `theme` imports. Remove unused
regexes (`reANSI`, `reOSC8`, `reOrderedList`, `reMdLink`).

Keep: `prepareHTML`, `stripHiddenElements`, `normalizeWhitespace`,
`deduplicateBlocks`, `stripEmptyLinks`, `unflattenQuotes`,
`findQuoteStart`, `isBlockquote`, `CleanHTML`, and their regexes.

- [ ] **Step 2: Remove deleted functions from headers.go**

Delete: `ColorSet`, `Headers`, `colorizeValue`, `wrapAddresses`.
Keep: `parseHeaders`, `stripBareAngles`, `headerBlock` (used by
`parseHeaders`). These unexported functions can be used by
`CleanPlain` or future cleanup rules.

Actually — `parseHeaders` is now in `internal/content/headers.go`.
The filter package no longer needs header parsing at all. Delete
the entire `headers.go` if no cleanup rules need it.

- [ ] **Step 3: Clean up plain.go**

Delete: old `Plain` function, `findColorize`. Keep: `detectHTML`,
`CleanPlain`, `htmlTagRe`, `reTabListItem`.

- [ ] **Step 4: Remove deleted tests**

Update `html_test.go`, `headers_test.go`, and `plain_test.go`
to only test the functions that remain.

- [ ] **Step 5: Remove glamour from go.mod**

Run: `go mod tidy`
Expected: glamour and its transitive deps removed. Lipgloss
promoted from indirect to direct.

- [ ] **Step 6: Delete TOML theme files**

```bash
rm .config/aerc/themes/nord.toml
rm .config/aerc/themes/solarized-dark.toml
rm .config/aerc/themes/gruvbox-dark.toml
```

- [ ] **Step 7: Run make check**

Run: `make check`
Expected: PASS — all vet and test clean

- [ ] **Step 8: Commit**

```bash
git add -u
git commit -m "Remove glamour, TOML themes, and dead rendering code"
```

---

### Task 15: Update docs and install

**Files:**
- Modify: `CLAUDE.md`
- Modify: `docs/styling.md`
- Modify: `docs/themes.md`

- [ ] **Step 1: Update CLAUDE.md theme references**

Update the "Theme System" section to reflect compiled themes:
- Remove references to `.toml` files
- Update `mailrender themes generate` description
- Note that themes are compiled Go values in `internal/theme/`

- [ ] **Step 2: Update docs/themes.md**

Update to document the new compiled theme system:
- `Palette` type with 16 color slots
- `NewTheme` constructor
- Three built-in themes
- How to add a new theme (add palette + `NewTheme` call)
- Styleset generator usage

- [ ] **Step 3: Update docs/styling.md**

Update rendering pipeline description:
- Three-layer architecture
- Block model types
- Lipgloss rendering
- Preview subcommand

- [ ] **Step 4: Run make install**

Run: `make install`
Expected: all binaries installed to `~/.local/bin/`

- [ ] **Step 5: Verify in aerc via tmux-testing**

Pipe a test email through the rebuilt `mailrender html` filter
and confirm the output renders correctly.

- [ ] **Step 6: Run /simplify**

- [ ] **Step 7: Commit and push**

```bash
git add CLAUDE.md docs/styling.md docs/themes.md
git commit -m "Update docs for lipgloss theme and rendering migration"
git push
```

package content

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

const maxBodyWidth = 78

// RenderBody renders blocks into a styled string using lipgloss.
// Width is capped at maxBodyWidth for readability.
func RenderBody(blocks []Block, t *theme.CompiledTheme, width int) string {
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

func renderBlock(block Block, t *theme.CompiledTheme, width int) string {
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
		return renderTable(b, t)

	case ListItem:
		text := renderSpans(b.Spans, t)
		prefix := "- "
		if b.Ordered {
			prefix = string(rune('0'+b.Index%10)) + ". "
		}
		return t.Paragraph.Render(prefix + text)

	default:
		return ""
	}
}

func renderSpans(spans []Span, t *theme.CompiledTheme) string {
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

func renderTable(table Table, t *theme.CompiledTheme) string {
	var rows []string

	if len(table.Headers) > 0 {
		var cells []string
		for _, hdr := range table.Headers {
			cells = append(cells, t.Bold.Render(renderSpans(hdr, t)))
		}
		rows = append(rows, strings.Join(cells, " | "))
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

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

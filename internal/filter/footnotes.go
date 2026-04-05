package filter

import (
	"fmt"
	"regexp"
	"strings"
)

// linkTextMarker wraps link text so styleFootnotes can identify it precisely.
const linkTextMarker = "\x00LT\x00"

// Regexes for reference-style link processing.
var (
	// reRefDef matches pandoc reference definitions: "  [label]: url" or "  [label]:" (empty URL)
	reRefDef = regexp.MustCompile(`(?m)^ {0,3}\[([^\]]+)\]:\s*(.*)$`)
	// reRefShortcut matches shortcut reference [text] in body, optionally with [N] or []
	reRefShortcut = regexp.MustCompile(`\[([^\]]+)\](?:\[(\d*)\])?`)
	// reAutolink matches autolinks <https://...>
	reAutolink = regexp.MustCompile(`<(https?://[^>]+)>`)
	// reImageRef matches standalone image refs: ![alt], ![alt][N], ![alt][]
	reImageRef = regexp.MustCompile(`!\[([^\]]*)\](?:\[(\d*)\])?`)
	// reImageLinkRef matches image wrapped in link: [![alt]][ref]
	reImageLinkRef = regexp.MustCompile(`\[!\[([^\]]*)\]\](?:\[(\d+)\])?`)
	// reEmptyTextRef matches empty-text reference links: [][ref]
	reEmptyTextRef = regexp.MustCompile(`\[\]\[(\d+)\]`)
)

// refDef holds a parsed reference definition.
type refDef struct {
	label string
	url   string
}

// footnoteRef holds a numbered footnote reference for the reference section.
type footnoteRef struct {
	num int
	url string
}

// convertToFootnotes transforms pandoc reference-style links into footnote
// syntax. Returns the transformed body text and a slice of footnote references.
// Self-referencing links (where label looks like a URL) become plain URLs.
// Image references, empty-text links, and empty-URL links are cleaned up.
func convertToFootnotes(text string) (string, []footnoteRef) {
	lines := strings.Split(text, "\n")
	var defs []refDef

	// Scan from bottom to find ref def block. Allow empty URL defs.
	i := len(lines) - 1
	for i >= 0 && strings.TrimSpace(lines[i]) == "" {
		i--
	}
	for i >= 0 {
		groups := reRefDef.FindStringSubmatch(lines[i])
		if groups == nil {
			break
		}
		defs = append(defs, refDef{label: groups[1], url: strings.TrimSpace(groups[2])})
		i--
	}
	bodyLines := lines[:i+1]

	// Reverse defs (collected bottom-up).
	for a, b := 0, len(defs)-1; a < b; a, b = a+1, b-1 {
		defs[a], defs[b] = defs[b], defs[a]
	}

	body := strings.Join(bodyLines, "\n")

	// Replace image references with alt text label or [image] placeholder.
	body = reImageLinkRef.ReplaceAllStringFunc(body, func(m string) string {
		groups := reImageLinkRef.FindStringSubmatch(m)
		if groups == nil {
			return m
		}
		alt := strings.TrimSpace(groups[1])
		if alt == "" {
			return ""
		}
		return "image: " + alt
	})
	body = reImageRef.ReplaceAllStringFunc(body, func(m string) string {
		groups := reImageRef.FindStringSubmatch(m)
		if groups == nil {
			return m
		}
		alt := strings.TrimSpace(groups[1])
		if alt == "" {
			return ""
		}
		return "image: " + alt
	})
	body = reEmptyTextRef.ReplaceAllString(body, "")

	// Build a set of labels still referenced in the body after image stripping.
	bodyRefs := make(map[string]bool)
	for _, m := range reRefShortcut.FindAllStringSubmatch(body, -1) {
		bodyRefs[m[1]] = true
		if m[2] != "" {
			bodyRefs[m[2]] = true
		}
	}

	// Categorize defs: only URL defs with surviving body references get footnotes.
	type numberedRef struct {
		num int
		url string
	}
	labelMap := make(map[string]numberedRef)
	stripLabels := make(map[string]bool)
	var refs []footnoteRef
	n := 0
	for _, d := range defs {
		if d.url == "" {
			stripLabels[d.label] = true
			continue
		}
		if !isURL(d.url) {
			continue
		}
		if !bodyRefs[d.label] {
			continue
		}
		if isSelfRef(d.label, d.url) {
			stripLabels[d.label] = true
			continue
		}
		n++
		labelMap[d.label] = numberedRef{num: n, url: d.url}
		refs = append(refs, footnoteRef{num: n, url: d.url})
	}

	// Replace body references with footnote markers or strip brackets.
	body = reRefShortcut.ReplaceAllStringFunc(body, func(m string) string {
		groups := reRefShortcut.FindStringSubmatch(m)
		if groups == nil {
			return m
		}
		label := groups[1]
		numericLabel := groups[2]
		display := stripEmphasis(label)

		if numericLabel != "" {
			if ref, ok := labelMap[numericLabel]; ok {
				return linkTextMarker + display + linkTextMarker + fmt.Sprintf("[^%d]", ref.num)
			}
		}

		if ref, ok := labelMap[label]; ok {
			return linkTextMarker + display + linkTextMarker + fmt.Sprintf("[^%d]", ref.num)
		}

		return display
	})

	// Convert autolinks to plain URLs.
	body = reAutolink.ReplaceAllString(body, "$1")

	return body, refs
}

// footnoteColors holds ANSI parameter strings for footnote styling.
type footnoteColors struct {
	LinkText string // body link text color
	Dim      string // footnote markers and ref labels
	LinkURL  string // reference section URLs
	Reset    string
}

// reFootnoteMarker matches "[^N]" markers in body text for dimming.
var reFootnoteMarker = regexp.MustCompile(`\[\^(\d+)\]`)

// styleFootnotes applies ANSI colors to footnote-annotated text.
// Link text (wrapped in linkTextMarker) gets link color, [^N] markers get dim color.
// A separator and colored reference section are appended.
func styleFootnotes(body string, refs []footnoteRef, cols int, colors *footnoteColors) string {
	if len(refs) == 0 {
		return strings.ReplaceAll(body, linkTextMarker, "")
	}

	lt := ""
	dim := ""
	lu := ""
	r := ""
	if colors.LinkText != "" {
		lt = "\033[" + colors.LinkText + "m"
	}
	if colors.Dim != "" {
		dim = "\033[" + colors.Dim + "m"
	}
	if colors.LinkURL != "" {
		lu = "\033[" + colors.LinkURL + "m"
	}
	if colors.Reset != "" {
		r = "\033[" + colors.Reset + "m"
	}

	body = replaceMarkerPairs(body, linkTextMarker, lt, r)
	body = reFootnoteMarker.ReplaceAllString(body, dim+"[^${1}]"+r)

	var sb strings.Builder
	sb.WriteString(body)
	sb.WriteString("\n" + dim + strings.Repeat("─", cols) + r + "\n")
	for _, ref := range refs {
		sb.WriteString(fmt.Sprintf("%s[^%d]:%s %s%s%s\n", dim, ref.num, r, lu, ref.url, r))
	}
	return sb.String()
}

// isSelfRef returns true when a reference label is effectively its own URL.
func isSelfRef(label, url string) bool {
	return strings.TrimPrefix(strings.TrimPrefix(label, "https://"), "http://") ==
		strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
}

// isURL returns true if s looks like a URL.
func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// stripEmphasis removes markdown emphasis markers from link display text.
// Pandoc wraps linked text in *...* or **...** when the HTML had <em>/<strong>.
// Link color already provides visual distinction, so emphasis is redundant.
func stripEmphasis(s string) string {
	s = strings.TrimPrefix(s, "**")
	s = strings.TrimSuffix(s, "**")
	s = strings.TrimPrefix(s, "*")
	s = strings.TrimSuffix(s, "*")
	return s
}

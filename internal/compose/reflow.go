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
// prefix, using minimum-raggedness line breaking. Returns a slice of
// lines, each starting with prefix.
func wrapText(text, prefix string, width int) []string {
	avail := width - runewidth.StringWidth(prefix)
	if avail <= 0 {
		return []string{prefix + text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{prefix}
	}

	n := len(words)
	wordLen := make([]int, n)
	for i, w := range words {
		wordLen[i] = runewidth.StringWidth(w)
	}

	// Minimum-raggedness DP: minimize sum of squared slack per line.
	const inf = 1 << 62
	cost := make([]int, n+1)
	breaks := make([]int, n)
	cost[n] = 0

	for i := n - 1; i >= 0; i-- {
		lineLen := -1
		best := inf
		bestBreak := n
		for j := i; j < n; j++ {
			lineLen += 1 + wordLen[j]
			if lineLen > avail && j > i {
				break
			}
			var c int
			if j == n-1 {
				c = cost[j+1]
			} else {
				slack := avail - lineLen
				c = slack*slack + cost[j+1]
			}
			if c < best {
				best = c
				bestBreak = j + 1
			}
		}
		cost[i] = best
		breaks[i] = bestBreak
	}

	var result []string
	for i := 0; i < n; {
		j := breaks[i]
		result = append(result, prefix+strings.Join(words[i:j], " "))
		i = j
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

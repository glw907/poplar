package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// FooterContext identifies which keybinding set to display.
type FooterContext int

const (
	// AccountContext is the unified one-pane account view (folders + messages).
	AccountContext FooterContext = iota
	ViewerContext
)

// Footer renders context-appropriate keybinding hints with group separators.
type Footer struct {
	styles     Styles
	context    FooterContext
	acctKeys   AccountKeys
	viewerKeys ViewerKeys
}

// NewFooter creates a Footer with the given styles.
func NewFooter(styles Styles) Footer {
	return Footer{
		styles:     styles,
		context:    AccountContext,
		acctKeys:   NewAccountKeys(),
		viewerKeys: NewViewerKeys(),
	}
}

// SetContext switches the displayed keybinding set.
func (f *Footer) SetContext(ctx FooterContext) {
	f.context = ctx
}

// View renders the footer at the given width.
//
// Hints are progressively dropped (highest dropRank first) when the
// full hint list would exceed width. Groups that become empty collapse
// along with their preceding separator. dropRank 0 hints are always
// kept — if they alone still exceed width, lipgloss will clip them.
func (f Footer) View(width int) string {
	var groups [][]footerHint
	switch f.context {
	case ViewerContext:
		groups = f.viewerKeys.FooterGroups()
	default:
		groups = f.acctKeys.FooterGroups()
	}

	visible := fitFooterHints(groups, width)

	sep := " " + f.styles.FooterSep.Render("┊") + "  "

	var parts []string
	for _, g := range visible {
		if len(g) == 0 {
			continue
		}
		var rendered []string
		for _, h := range g {
			k := f.styles.FooterKey.Render(h.key)
			d := f.styles.FooterHint.Render(" " + h.desc)
			rendered = append(rendered, k+d)
		}
		parts = append(parts, strings.Join(rendered, "  "))
	}

	line := " " + strings.Join(parts, sep)
	return lipgloss.NewStyle().Width(width).Render(line)
}

// fitFooterHints returns a trimmed copy of groups that fits within
// width, dropping hints with the highest dropRank first. When a group
// loses all of its hints, it vanishes from the output (taking its
// separator with it).
func fitFooterHints(groups [][]footerHint, width int) [][]footerHint {
	visible := make([][]footerHint, len(groups))
	for i, g := range groups {
		visible[i] = append([]footerHint(nil), g...)
	}
	for measureFooter(visible) > width {
		gi, hi := highestDropRank(visible)
		if gi < 0 {
			break
		}
		visible[gi] = append(visible[gi][:hi], visible[gi][hi+1:]...)
	}
	return visible
}

// highestDropRank returns the (group, hint) index of the highest
// dropRank hint across all groups, or (-1, -1) if nothing is droppable
// (everything left is dropRank 0).
func highestDropRank(groups [][]footerHint) (int, int) {
	bestGroup, bestHint, bestRank := -1, -1, 0
	for gi, g := range groups {
		for hi, h := range g {
			if h.dropRank == 0 {
				continue
			}
			if bestGroup < 0 || h.dropRank > bestRank {
				bestGroup, bestHint, bestRank = gi, hi, h.dropRank
			}
		}
	}
	return bestGroup, bestHint
}

// measureFooter returns the plain-text rune width of the rendered
// footer: leading space, hints joined by two spaces within a group,
// and non-empty groups joined by " ┊  " (4 runes) separators.
func measureFooter(groups [][]footerHint) int {
	total := 1 // leading space
	first := true
	for _, g := range groups {
		if len(g) == 0 {
			continue
		}
		if !first {
			total += 4 // " ┊  " sep (4 runes: space, ┊, space, space)
		}
		first = false
		for j, h := range g {
			if j > 0 {
				total += 2 // "  " between hints
			}
			total += utf8.RuneCountInString(h.key) + 1 + utf8.RuneCountInString(h.desc)
		}
	}
	return total
}

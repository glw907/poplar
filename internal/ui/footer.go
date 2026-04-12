package ui

import (
	"strings"

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
func (f Footer) View(width int) string {
	var groups []keyGroup
	switch f.context {
	case ViewerContext:
		groups = f.viewerKeys.Groups()
	default:
		groups = f.acctKeys.Groups()
	}

	sep := " " + f.styles.FooterSep.Render("┊") + "  "

	var parts []string
	for _, g := range groups {
		var bindings []string
		for _, b := range g {
			k := f.styles.FooterKey.Render(b.Help().Key)
			d := f.styles.FooterHint.Render(" " + b.Help().Desc)
			bindings = append(bindings, k+d)
		}
		parts = append(parts, strings.Join(bindings, "  "))
	}

	line := " " + strings.Join(parts, sep)
	return lipgloss.NewStyle().Width(width).Render(line)
}

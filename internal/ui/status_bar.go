package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders a one-row status line with folder info and connection state.
type StatusBar struct {
	styles    Styles
	icon      string
	folder    string
	total     int
	unread    int
	connected bool
}

// NewStatusBar creates a StatusBar with the given styles.
func NewStatusBar(styles Styles) StatusBar {
	return StatusBar{
		styles:    styles,
		connected: true,
	}
}

// SetFolder updates the displayed folder information.
func (sb *StatusBar) SetFolder(icon, name string, total, unread int) {
	sb.icon = icon
	sb.folder = name
	sb.total = total
	sb.unread = unread
}

// SetConnected updates the connection state.
func (sb *StatusBar) SetConnected(connected bool) {
	sb.connected = connected
}

// View renders the status bar at the given width.
func (sb StatusBar) View(width int) string {
	left := fmt.Sprintf(" %s  %s · %d messages", sb.icon, sb.folder, sb.total)
	if sb.unread > 0 {
		left += fmt.Sprintf(" · %d unread", sb.unread)
	}

	var right string
	if sb.connected {
		dot := sb.styles.StatusConnected.Render("●")
		right = dot + " connected "
	} else {
		dot := sb.styles.StatusOffline.Render("●")
		right = dot + " offline "
	}

	gap := maxInt(0, width-lipgloss.Width(left)-lipgloss.Width(right))
	middle := strings.Repeat(" ", gap)

	return sb.styles.StatusBar.Render(left + middle + right)
}

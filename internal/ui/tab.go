package ui

import tea "github.com/charmbracelet/bubbletea"

// Tab is a bubbletea model that renders inside the chrome shell.
// Each tab occupies the content area between the tab bar and status bar.
type Tab interface {
	tea.Model

	// Title returns the tab's display title (e.g., folder name).
	Title() string

	// Icon returns a Nerd Font icon for the tab bar.
	Icon() string

	// Closeable returns whether the user can close this tab.
	Closeable() bool
}

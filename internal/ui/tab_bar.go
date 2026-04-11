package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// tabInfo holds the data needed to render a single tab in the tab bar.
type tabInfo struct {
	title string
	icon  string
}

// maxTabTitle is the maximum display width for a tab title.
const maxTabTitle = 30

// truncateTitle caps a title at max runes, appending "…" if truncated.
func truncateTitle(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

// renderTabBar renders the 3-row bubble tab bar.
//
// The active tab is a rounded bubble that opens into the content area:
//
//	Row 1:  ╭───────────╮
//	Row 2:  │ 󰇰  Inbox  │  Re: Project update
//	Row 3: ─╯           ╰──────────────────────────╮
func renderTabBar(tabs []tabInfo, active, width int, s Styles) string {
	if len(tabs) == 0 || width < 20 {
		return ""
	}

	// Build the active tab content: " icon  title "
	activeTab := tabs[active]
	activeTitle := truncateTitle(activeTab.title, maxTabTitle)
	activeContent := " " + activeTab.icon + "  " + activeTitle + " "
	activeInner := lipgloss.Width(activeContent)

	// Compute the left offset: sum of widths of inactive tabs before active
	var beforeParts []string
	for i := 0; i < active; i++ {
		t := truncateTitle(tabs[i].title, maxTabTitle)
		beforeParts = append(beforeParts, " "+tabs[i].icon+"  "+t+" ")
	}
	leftOffset := 0
	for _, p := range beforeParts {
		leftOffset += lipgloss.Width(p)
		leftOffset += 3 // " · " separator
	}

	// Build inactive tabs after active
	var afterParts []string
	for i := active + 1; i < len(tabs); i++ {
		t := truncateTitle(tabs[i].title, maxTabTitle)
		afterParts = append(afterParts, tabs[i].icon+"  "+t)
	}
	afterStr := ""
	if len(afterParts) > 0 {
		afterStr = "  " + strings.Join(afterParts, "  ·  ")
	}

	border := s.TabActiveBorder
	activeText := s.TabActiveText
	inactiveText := s.TabInactiveText
	connectLine := s.TabConnectLine

	// Row 1: padding + ╭ + ─ fill + ╮
	row1Pad := strings.Repeat(" ", leftOffset)
	row1Inner := strings.Repeat("─", activeInner)
	row1 := row1Pad + border.Render("╭"+row1Inner+"╮")
	row1 += strings.Repeat(" ", maxInt(0, width-lipgloss.Width(row1)))

	// Row 2: inactive before + │ content │ + inactive after
	row2 := ""
	for i, p := range beforeParts {
		row2 += inactiveText.Render(p)
		if i < len(beforeParts)-1 {
			row2 += inactiveText.Render(" · ")
		} else {
			row2 += inactiveText.Render(" · ")
		}
	}
	row2 += border.Render("│") + activeText.Render(activeContent) + border.Render("│")
	if afterStr != "" {
		row2 += inactiveText.Render(afterStr)
	}
	row2 += strings.Repeat(" ", maxInt(0, width-lipgloss.Width(row2)))

	// Row 3: ─╯ + spaces + ╰ + ─ fill + ╮
	row3Left := connectLine.Render(strings.Repeat("─", maxInt(1, leftOffset)) + "╯")
	row3Mid := strings.Repeat(" ", activeInner)
	rightFill := maxInt(0, width-lipgloss.Width(row3Left)-activeInner-2)
	row3Right := connectLine.Render("╰" + strings.Repeat("─", rightFill) + "╮")
	row3 := row3Left + row3Mid + row3Right

	return row1 + "\n" + row2 + "\n" + row3
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

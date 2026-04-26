package ui

import "github.com/charmbracelet/lipgloss"

// Nerd Font icons live in the Supplementary Private Use Area-A
// (U+F0000–U+FFFFD). Every modern terminal renders them as 2 display
// cells, but go-runewidth (and lipgloss.Width, which uses it) reports
// them as 1 cell because the Unicode standard does not assign East
// Asian width to private-use codepoints.
//
// Use displayCells to measure any string that may contain Nerd Font
// icons. lipgloss.Width remains correct for icon-free strings.
const (
	spuaAStart = 0xF0000
	spuaAEnd   = 0xFFFFD
)

// displayCells returns the actual terminal display width of s,
// correcting runewidth's undercount of Nerd Font SPUA-A glyphs.
func displayCells(s string) int {
	return lipgloss.Width(s) + spuaACorrection(s)
}

// spuaACorrection counts SPUA-A runes in s. Each contributes one
// extra display cell beyond runewidth's reported width.
func spuaACorrection(s string) int {
	n := 0
	for _, r := range s {
		if r >= spuaAStart && r <= spuaAEnd {
			n++
		}
	}
	return n
}

package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	// aerc invokes filters as piped commands (no TTY on stdout), but the
	// output is displayed in aerc's terminal viewer. Force TrueColor so
	// lipgloss emits ANSI styles instead of stripping them.
	lipgloss.SetColorProfile(termenv.TrueColor)
}

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

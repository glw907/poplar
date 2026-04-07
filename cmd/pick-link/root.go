package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/glw907/beautiful-aerc/internal/filter"
	"github.com/glw907/beautiful-aerc/internal/palette"
	"github.com/glw907/beautiful-aerc/internal/picker"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "pick-link",
		Short:        "Interactive URL picker for aerc email messages",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := loadPalette()
			if err != nil {
				return err
			}

			cols := termCols()
			links, err := filter.HTMLLinks(os.Stdin, cols)
			if err != nil {
				return err
			}

			colors := picker.ColorsFromPalette(p)
			url, err := picker.Run(links, cols, colors)
			if err != nil {
				return err
			}
			if url != "" {
				name := "xdg-open"
				if strings.HasPrefix(url, "mailto:") {
					name = "aerc"
				}
				open := exec.Command(name, url)
				open.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
				return open.Start()
			}
			return nil
		},
	}
	return cmd
}

// loadPalette finds and loads the palette file relative to the binary location.
func loadPalette() (*palette.Palette, error) {
	binPath, _ := os.Executable()
	genDir := ""
	if binPath != "" {
		resolved, err := filepath.EvalSymlinks(binPath)
		if err == nil {
			binPath = resolved
		}
		binDir := filepath.Dir(binPath)
		genDir = filepath.Join(binDir, "..", "..", ".config", "aerc", "generated")
	}
	path, err := palette.FindPath(genDir)
	if err != nil {
		return nil, err
	}
	return palette.Load(path)
}

// termCols returns the terminal column count from AERC_COLUMNS or a default of 80.
func termCols() int {
	if s := os.Getenv("AERC_COLUMNS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return n
		}
	}
	return 80
}

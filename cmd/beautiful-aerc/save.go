package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/beautiful-aerc/internal/corpus"
	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save current email part to corpus for later analysis",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("reading stdin: %w", err)
			}

			binPath, _ := os.Executable()
			configHint := ""
			if binPath != "" {
				resolved, err := filepath.EvalSymlinks(binPath)
				if err == nil {
					binPath = resolved
				}
				binDir := filepath.Dir(binPath)
				configHint = filepath.Join(binDir, "..", "..", ".config", "aerc")
			}

			dir, err := corpus.FindDir(configHint)
			if err != nil {
				return err
			}

			path, err := corpus.Save(dir, data)
			if err != nil {
				return err
			}

			// Count pending corpus files.
			entries, _ := os.ReadDir(dir)
			count := 0
			for _, e := range entries {
				if !e.IsDir() {
					count++
				}
			}

			printSaveNotification(filepath.Base(path), count)
			return nil
		},
	}
	return cmd
}

func printSaveNotification(filename string, pending int) {
	p, _ := loadPalette()
	marker := ""
	heading := ""
	detail := ""
	dim := ""
	reset := ""
	if p != nil {
		marker = p.ANSI("C_MSG_MARKER")
		heading = p.ANSI("C_MSG_TITLE_SUCCESS")
		detail = p.ANSI("C_MSG_DETAIL")
		dim = p.ANSI("C_MSG_DIM")
		reset = p.Reset()
	}

	rows := termRows()
	pad := (rows - 4) / 3
	fmt.Print("\033[?25l")
	fmt.Print(strings.Repeat("\n", pad))

	fmt.Printf(" %s#%s %sSAVED TO CORPUS%s\n", marker, reset, heading, reset)
	fmt.Println()
	fmt.Printf(" %s%s%s\n", detail, filename, reset)
	fmt.Printf(" %s%d pending%s\n", dim, pending, reset)
}

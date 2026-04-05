package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

			fmt.Fprintf(os.Stderr, "saved %s\n", filepath.Base(path))
			return nil
		},
	}
	return cmd
}

package main

import (
	"fmt"
	"os"

	"github.com/glw907/beautiful-aerc/internal/picker"
	"github.com/spf13/cobra"
)

func newPickLinkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick-link",
		Short: "Interactive URL picker for email messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := loadPalette()
			if err != nil {
				return err
			}
			colors := picker.ColorsFromPalette(p)
			url, err := picker.Run(os.Stdin, os.Stderr, colors)
			if err != nil {
				return err
			}
			if url != "" {
				fmt.Fprintln(os.Stdout, url)
			}
			return nil
		},
	}
	return cmd
}

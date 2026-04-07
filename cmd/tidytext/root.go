package main

import "github.com/spf13/cobra"

var version = "0.1.0"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tidytext",
		Short:        "Tidy prose with AI-powered spelling, grammar, and punctuation fixes",
		Version:      version,
		SilenceUsage: true,
	}
	cmd.AddCommand(newFixCmd())
	cmd.AddCommand(newConfigCmd())
	return cmd
}

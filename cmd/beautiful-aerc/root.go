package main

import "github.com/spf13/cobra"

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "beautiful-aerc",
		Short:        "Themeable filters for the aerc email client",
		SilenceUsage: true,
	}
	cmd.AddCommand(newHeadersCmd())
	cmd.AddCommand(newHTMLCmd())
	cmd.AddCommand(newPlainCmd())
	cmd.AddCommand(newPickLinkCmd())
	return cmd
}

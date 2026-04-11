package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := newRootCmd()
	cmd.AddCommand(newThemesCmd())
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/glw907/beautiful-aerc/internal/compose"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type composeFlags struct {
	noCcBcc bool
	debug   bool
}

func newComposeCmd() *cobra.Command {
	f := composeFlags{}

	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Normalize aerc compose buffers for editing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompose(f)
		},
	}

	cmd.Flags().BoolVar(&f.noCcBcc, "no-cc-bcc", false, "do not inject empty Cc/Bcc headers")
	cmd.Flags().BoolVar(&f.debug, "debug", false, "write diagnostic messages to stderr")

	return cmd
}

func runCompose(f composeFlags) error {
	if f.debug {
		log.SetPrefix("mailrender compose: ")
		log.SetFlags(0)
	} else {
		log.SetOutput(io.Discard)
	}

	if term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("no input (pipe a compose buffer to stdin)")
	}

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	opts := compose.Options{
		InjectCcBcc: !f.noCcBcc,
	}

	output := compose.Prepare(input, opts)

	if _, err := os.Stdout.Write(output); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

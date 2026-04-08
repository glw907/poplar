package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/beautiful-aerc/internal/tidy"
	"github.com/spf13/cobra"
)

const defaultConfigPath = "~/.config/tidytext/config.toml"

// defaultConfigTOML is the template written by "config init".
const defaultConfigTOML = `[api]
# model = "claude-haiku-4-5-20251001"
# api_key = ""  # leave unset to use ANTHROPIC_API_KEY env var

[rules]
spelling            = true
grammar             = true
punctuation         = true
whitespace          = true
capitalization      = true
repeated_words      = true
missing_punctuation = true
oxford_comma        = "ignore"  # insert | remove | ignore

[style]
em_dash_spaces = false
ellipsis       = "character"  # character | dots
time_format    = "ignore"     # uppercase (10:00 AM) | lowercase (10:00am) | periods (10:00 a.m.) | ignore
# custom_instructions = [
#   "Prefer active voice.",
# ]
`

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "config",
		Short:        "Show or initialize tidytext configuration",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigShow()
		},
	}
	cmd.AddCommand(newConfigInitCmd())
	return cmd
}

func runConfigShow() error {
	path := expandHome(defaultConfigPath)
	cfg, err := tidy.LoadConfig(path)
	if err != nil {
		return err
	}
	fmt.Print(tidy.ConfigString(cfg))
	return nil
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "init",
		Short:        "Create default config file at " + defaultConfigPath,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigInit()
		},
	}
}

func runConfigInit() error {
	path := expandHome(defaultConfigPath)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// O_CREATE|O_EXCL atomically fails if the file already exists.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("config file already exists: %s", path)
		}
		return fmt.Errorf("create config: %w", err)
	}
	_, writeErr := f.WriteString(defaultConfigTOML)
	closeErr := f.Close()
	if writeErr != nil {
		os.Remove(path)
		return fmt.Errorf("write config: %w", writeErr)
	}
	if closeErr != nil {
		os.Remove(path)
		return fmt.Errorf("write config: %w", closeErr)
	}

	fmt.Printf("created %s\n", path)
	return nil
}

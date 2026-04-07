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

	// Error if file already exists.
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check config path: %w", err)
	}

	// Create directory if needed.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(defaultConfigTOML), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("created %s\n", path)
	return nil
}

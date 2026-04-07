package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/glw907/beautiful-aerc/internal/tidy"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type fixFlags struct {
	configPath string
	inPlace    bool
	noConfig   bool
	rules      []string
	styles     []string
}

func newFixCmd() *cobra.Command {
	f := fixFlags{}

	cmd := &cobra.Command{
		Use:          "fix [file]",
		Short:        "Fix spelling, grammar, and punctuation in a text file or stdin",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFix(f, args)
		},
	}

	cmd.Flags().StringVar(&f.configPath, "config", "~/.config/tidytext/config.toml", "path to config file")
	cmd.Flags().BoolVar(&f.inPlace, "in-place", false, "overwrite the input file with fixed text")
	cmd.Flags().BoolVar(&f.noConfig, "no-config", false, "skip loading config file, use defaults")
	cmd.Flags().StringSliceVar(&f.rules, "rule", nil, "rule override in key=value form (may be repeated)")
	cmd.Flags().StringSliceVar(&f.styles, "style", nil, "style override in key=value form (may be repeated)")

	return cmd
}

func runFix(f fixFlags, args []string) error {
	var input []byte
	var filePath string

	switch {
	case len(args) > 0:
		filePath = args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}
		input = data

	case !term.IsTerminal(int(os.Stdin.Fd())):
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read stdin: %w", err)
		}
		input = data

	default:
		return fmt.Errorf("no input provided (pipe text or pass a file argument)")
	}

	// Load config.
	cfg := tidy.DefaultConfig()
	if !f.noConfig {
		path := expandHome(f.configPath)
		loaded, err := tidy.LoadConfig(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tidytext: config error: %v, using defaults\n", err)
			// Safe fallback: output original text, return nil.
			fmt.Fprint(os.Stdout, string(input))
			return nil
		}
		cfg = loaded
	}

	// Apply overrides.
	if err := tidy.ApplyRuleOverrides(&cfg, f.rules); err != nil {
		return err
	}
	if err := tidy.ApplyStyleOverrides(&cfg, f.styles); err != nil {
		return err
	}

	// Resolve API credentials.
	apiKey := tidy.ResolveAPIKey(cfg)
	apiURL := os.Getenv("TIDYTEXT_API_URL")

	// Run tidy.
	result, err := tidy.Tidy(string(input), cfg, apiKey, apiURL)
	if err != nil {
		return err
	}

	// Print status message to stderr.
	fmt.Fprintln(os.Stderr, result.Message)

	// Write output.
	if f.inPlace && filePath != "" {
		if err := writeInPlace(filePath, result.Text); err != nil {
			return err
		}
	} else {
		fmt.Fprint(os.Stdout, result.Text)
	}

	return nil
}

// writeInPlace writes text to a temp file in the same directory as dst, then
// renames it over dst atomically.
func writeInPlace(dst, text string) error {
	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".tidytext-*")
	if err != nil {
		return fmt.Errorf("write in-place: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := fmt.Fprint(tmp, text); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write in-place: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("write in-place: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("write in-place: %w", err)
	}
	return nil
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) string {
	if len(path) >= 2 && path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

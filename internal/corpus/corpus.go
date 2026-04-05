package corpus

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindDir locates or creates the corpus directory. Resolution order:
// 1. $AERC_CONFIG/../../corpus/ (env override)
// 2. configHint/../../corpus/ (caller-supplied aerc config path)
// 3. ~/.config/aerc/../../corpus/ (default)
// Creates the directory if it does not exist.
func FindDir(configHint string) (string, error) {
	var candidates []string

	if aercConfig := os.Getenv("AERC_CONFIG"); aercConfig != "" {
		candidates = append(candidates, filepath.Join(aercConfig, "..", "..", "corpus"))
	}

	if configHint != "" {
		candidates = append(candidates, filepath.Join(configHint, "..", "..", "corpus"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "aerc", "..", "..", "corpus"))
	}

	for _, c := range candidates {
		c = filepath.Clean(c)
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c, nil
		}
	}

	if len(candidates) > 0 {
		dir := filepath.Clean(candidates[0])
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("creating corpus directory %s: %w", dir, err)
		}
		return dir, nil
	}

	return "", fmt.Errorf("cannot determine corpus directory")
}

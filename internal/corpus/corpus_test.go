package corpus

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (envVal string, binHint string)
		wantErr bool
	}{
		{
			"env override",
			func(t *testing.T) (string, string) {
				dir := t.TempDir()
				corpus := filepath.Join(dir, "corpus")
				os.MkdirAll(corpus, 0755)
				return dir, ""
			},
			false,
		},
		{
			"relative to binary hint",
			func(t *testing.T) (string, string) {
				dir := t.TempDir()
				aercDir := filepath.Join(dir, ".config", "aerc")
				os.MkdirAll(aercDir, 0755)
				corpus := filepath.Join(dir, "corpus")
				os.MkdirAll(corpus, 0755)
				return "", aercDir
			},
			false,
		},
		{
			"creates corpus dir if missing",
			func(t *testing.T) (string, string) {
				dir := t.TempDir()
				aercDir := filepath.Join(dir, ".config", "aerc")
				os.MkdirAll(aercDir, 0755)
				return "", aercDir
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envVal, binHint := tt.setup(t)
			if envVal != "" {
				t.Setenv("AERC_CONFIG", envVal)
			} else {
				t.Setenv("AERC_CONFIG", "")
			}
			dir, err := FindDir(binHint)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, err := os.Stat(dir); err != nil {
				t.Errorf("corpus dir does not exist: %v", err)
			}
		})
	}
}

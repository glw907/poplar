package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadUI(t *testing.T) {
	tests := []struct {
		name    string
		toml    string
		want    UIConfig
		wantErr string
	}{
		{
			name: "missing [ui] section uses defaults",
			toml: `[[account]]
name = "X"
source = "jmap://x@y"
`,
			want: UIConfig{
				Threading: true,
				Folders:   map[string]FolderConfig{},
			},
		},
		{
			name: "empty [ui] section uses defaults",
			toml: `[ui]
`,
			want: UIConfig{
				Threading: true,
				Folders:   map[string]FolderConfig{},
			},
		},
		{
			name: "global threading override",
			toml: `[ui]
threading = false
`,
			want: UIConfig{
				Threading: false,
				Folders:   map[string]FolderConfig{},
			},
		},
		{
			name: "per-folder rank and threading",
			toml: `[ui]
threading = true

[ui.folders.Inbox]
rank = 1
threading = false
sort = "date-asc"
`,
			want: UIConfig{
				Threading: true,
				Folders: map[string]FolderConfig{
					"Inbox": {
						Rank:         1,
						RankSet:      true,
						Threading:    false,
						ThreadingSet: true,
						Sort:         "date-asc",
					},
				},
			},
		},
		{
			name: "per-folder label and hide",
			toml: `[ui.folders."[Gmail]/All Mail"]
hide = true

[ui.folders."[Gmail]/Starred"]
label = "Starred"
rank = 5
`,
			want: UIConfig{
				Threading: true,
				Folders: map[string]FolderConfig{
					"[Gmail]/All Mail": {Hide: true},
					"[Gmail]/Starred":  {Label: "Starred", Rank: 5, RankSet: true},
				},
			},
		},
		{
			name: "invalid sort value rejected",
			toml: `[ui.folders.Inbox]
sort = "alphabetical"
`,
			wantErr: `invalid sort "alphabetical"`,
		},
		{
			name: "negative rank is valid",
			toml: `[ui.folders.Pinned]
rank = -10
`,
			want: UIConfig{
				Threading: true,
				Folders: map[string]FolderConfig{
					"Pinned": {Rank: -10, RankSet: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "accounts.toml")
			if err := os.WriteFile(path, []byte(tt.toml), 0644); err != nil {
				t.Fatal(err)
			}
			got, err := LoadUI(path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Threading != tt.want.Threading {
				t.Errorf("Threading = %v, want %v", got.Threading, tt.want.Threading)
			}
			if len(got.Folders) != len(tt.want.Folders) {
				t.Fatalf("Folders len = %d, want %d (got %+v)", len(got.Folders), len(tt.want.Folders), got.Folders)
			}
			for k, wantFC := range tt.want.Folders {
				gotFC, ok := got.Folders[k]
				if !ok {
					t.Errorf("missing folder %q", k)
					continue
				}
				if gotFC != wantFC {
					t.Errorf("folder %q = %+v, want %+v", k, gotFC, wantFC)
				}
			}
		})
	}
}

func TestLoadUIMissingFile(t *testing.T) {
	_, err := LoadUI("/nonexistent/accounts.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "reading ui config") {
		t.Errorf("expected 'reading ui config' in error, got %q", err.Error())
	}
}

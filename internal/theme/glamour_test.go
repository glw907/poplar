package theme

import (
	"os"
	"path/filepath"
	"testing"
)

const testThemeTOML = `
name = "test"

[colors]
bg_base = "#2e3440"
bg_elevated = "#3b4252"
bg_selection = "#394353"
bg_border = "#49576b"
fg_base = "#d8dee9"
fg_bright = "#e5e9f0"
fg_brightest = "#eceff4"
fg_dim = "#616e88"
accent_primary = "#81a1c1"
accent_secondary = "#88c0d0"
accent_tertiary = "#8fbcbb"
color_error = "#bf616a"
color_warning = "#d08770"
color_success = "#a3be8c"
color_info = "#ebcb8b"
color_special = "#b48ead"

[tokens]
heading = { color = "color_success", bold = true }
bold = { bold = true }
italic = { italic = true }
link_text = { color = "accent_secondary" }
rule = { color = "fg_dim" }
hdr_key = { color = "accent_primary", bold = true }
hdr_value = { color = "fg_base" }
hdr_dim = { color = "fg_dim" }
`

const minimalThemeTOML = `
name = "minimal"

[colors]
bg_base = "#2e3440"
bg_elevated = "#3b4252"
bg_selection = "#394353"
bg_border = "#49576b"
fg_base = "#d8dee9"
fg_bright = "#e5e9f0"
fg_brightest = "#eceff4"
fg_dim = "#616e88"
accent_primary = "#81a1c1"
accent_secondary = "#88c0d0"
accent_tertiary = "#8fbcbb"
color_error = "#bf616a"
color_warning = "#d08770"
color_success = "#a3be8c"
color_info = "#ebcb8b"
color_special = "#b48ead"
`

func writeTestTheme(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name+".toml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing test theme: %v", err)
	}
	return path
}

func TestGlamourStyle(t *testing.T) {
	t.Run("full theme", func(t *testing.T) {
		path := writeTestTheme(t, "test", testThemeTOML)
		th, err := Load(path)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}

		style := th.GlamourStyle()

		// Document margins should be 0.
		if style.Document.Margin == nil || *style.Document.Margin != 0 {
			t.Errorf("Document.Margin: got %v, want 0", style.Document.Margin)
		}
		if style.Document.Indent == nil || *style.Document.Indent != 0 {
			t.Errorf("Document.Indent: got %v, want 0", style.Document.Indent)
		}

		// H1 should use heading token: color_success (#a3be8c), bold.
		if style.H1.Color == nil || *style.H1.Color != "#a3be8c" {
			t.Errorf("H1.Color: got %v, want #a3be8c", style.H1.Color)
		}
		if style.H1.Bold == nil || !*style.H1.Bold {
			t.Errorf("H1.Bold: got %v, want true", style.H1.Bold)
		}

		// H6 should match H1 (same heading token).
		if style.H6.Color == nil || *style.H6.Color != "#a3be8c" {
			t.Errorf("H6.Color: got %v, want #a3be8c", style.H6.Color)
		}
		if style.H6.Bold == nil || !*style.H6.Bold {
			t.Errorf("H6.Bold: got %v, want true", style.H6.Bold)
		}

		// Strong should be bold only.
		if style.Strong.Bold == nil || !*style.Strong.Bold {
			t.Errorf("Strong.Bold: got %v, want true", style.Strong.Bold)
		}
		if style.Strong.Color != nil {
			t.Errorf("Strong.Color: got %v, want nil", style.Strong.Color)
		}

		// Emph should be italic only.
		if style.Emph.Italic == nil || !*style.Emph.Italic {
			t.Errorf("Emph.Italic: got %v, want true", style.Emph.Italic)
		}
		if style.Emph.Color != nil {
			t.Errorf("Emph.Color: got %v, want nil", style.Emph.Color)
		}

		// LinkText should use accent_secondary (#88c0d0).
		if style.LinkText.Color == nil || *style.LinkText.Color != "#88c0d0" {
			t.Errorf("LinkText.Color: got %v, want #88c0d0", style.LinkText.Color)
		}

		// HorizontalRule should use fg_dim (#616e88).
		if style.HorizontalRule.Color == nil || *style.HorizontalRule.Color != "#616e88" {
			t.Errorf("HorizontalRule.Color: got %v, want #616e88", style.HorizontalRule.Color)
		}
	})

	t.Run("minimal theme no tokens", func(t *testing.T) {
		path := writeTestTheme(t, "minimal", minimalThemeTOML)
		th, err := Load(path)
		if err != nil {
			t.Fatalf("Load: %v", err)
		}

		// Should not panic.
		style := th.GlamourStyle()

		// Margins still 0.
		if style.Document.Margin == nil || *style.Document.Margin != 0 {
			t.Errorf("Document.Margin: got %v, want 0", style.Document.Margin)
		}
		if style.Document.Indent == nil || *style.Document.Indent != 0 {
			t.Errorf("Document.Indent: got %v, want 0", style.Document.Indent)
		}

		// No tokens defined, so heading/strong/emph/link should be zero-value.
		if style.H1.Color != nil {
			t.Errorf("H1.Color: got %v, want nil", style.H1.Color)
		}
		if style.Strong.Bold != nil {
			t.Errorf("Strong.Bold: got %v, want nil", style.Strong.Bold)
		}
	})
}

package theme

import "testing"

func TestPaletteHex(t *testing.T) {
	tests := []struct {
		name string
		slot string
		want string
	}{
		{"bg_base", "bg_base", "#2e3440"},
		{"accent_primary", "accent_primary", "#81a1c1"},
		{"unknown", "nonexistent", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Nord.PaletteHex(tt.slot)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAllThemesHaveDistinctColors(t *testing.T) {
	if Nord.PaletteHex("bg_base") == SolarizedDark.PaletteHex("bg_base") {
		t.Error("Nord and SolarizedDark have same bg_base")
	}
	if Nord.PaletteHex("bg_base") == GruvboxDark.PaletteHex("bg_base") {
		t.Error("Nord and GruvboxDark have same bg_base")
	}
}

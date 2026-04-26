package ui

import "testing"

func TestDisplayCells(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"ascii", "Inbox", 5},
		{"single SPUA-A icon", "\U000f01f0", 2},
		{"icon + space + text", "\U000f01f0 Inbox", 8},
		{"two icons", "\U000f01f0\U000f045a", 4},
		{"box drawing not corrected", "│  ├─ ", 6},
		{"BMP private use not corrected", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := displayCells(tt.in)
			if got != tt.want {
				t.Errorf("displayCells(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

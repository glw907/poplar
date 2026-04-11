package ui

import (
	"strings"
	"testing"

	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestStatusBarView(t *testing.T) {
	styles := NewStyles(theme.Nord)

	t.Run("renders folder info and connection", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetFolder("󰇰", "Inbox", 10, 3)
		sb.SetConnected(true)
		result := sb.View(80)
		if !strings.Contains(result, "Inbox") {
			t.Error("missing folder name")
		}
		if !strings.Contains(result, "10") {
			t.Error("missing message count")
		}
		if !strings.Contains(result, "connected") {
			t.Error("missing connection indicator")
		}
	})

	t.Run("disconnected state", func(t *testing.T) {
		sb := NewStatusBar(styles)
		sb.SetFolder("󰇰", "Inbox", 10, 3)
		sb.SetConnected(false)
		result := sb.View(80)
		if !strings.Contains(result, "offline") {
			t.Error("missing offline indicator")
		}
	})
}

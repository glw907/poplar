package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/theme"
)

func TestApp(t *testing.T) {
	backend := mail.NewMockBackend()

	t.Run("initial state", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		if len(app.tabs) != 1 {
			t.Fatalf("expected 1 tab, got %d", len(app.tabs))
		}
		if app.activeTab != 0 {
			t.Errorf("activeTab = %d, want 0", app.activeTab)
		}
	})

	t.Run("quit on q", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app.width = 80
		app.height = 24
		_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		if cmd == nil {
			t.Fatal("expected quit command")
		}
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); !ok {
			t.Errorf("expected QuitMsg, got %T", msg)
		}
	})

	t.Run("quit on ctrl+c", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app.width = 80
		app.height = 24
		_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		if cmd == nil {
			t.Fatal("expected quit command")
		}
	})

	t.Run("window size stored", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		if app.width != 120 || app.height != 40 {
			t.Errorf("size = %dx%d, want 120x40", app.width, app.height)
		}
	})

	t.Run("tab delegates to account tab", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app.width = 80
		app.height = 24
		app, _ = app.Update(tea.KeyMsg{Type: tea.KeyTab})
		acct, ok := app.tabs[0].(AccountTab)
		if !ok {
			t.Fatal("tabs[0] is not AccountTab")
		}
		if acct.focused != MsgListPanel {
			t.Errorf("after Tab, focused = %d, want MsgListPanel", acct.focused)
		}
	})

	t.Run("view renders all sections", func(t *testing.T) {
		app := NewApp(theme.Nord, backend)
		app.width = 80
		app.height = 24
		view := app.View()
		if !strings.Contains(view, "Inbox") {
			t.Error("view missing Inbox in tab bar")
		}
		if !strings.Contains(view, "connected") {
			t.Error("view missing connection indicator")
		}
	})
}

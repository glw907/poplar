package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glw907/beautiful-aerc/internal/mail"
	"github.com/glw907/beautiful-aerc/internal/theme"
	"github.com/glw907/beautiful-aerc/internal/ui"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "poplar",
		Short:        "A bubbletea-based terminal email client",
		SilenceUsage: true,
		RunE:         runRoot,
	}
	return cmd
}

// appModel wraps ui.App to satisfy tea.Model (returns tea.Model, not App).
type appModel struct {
	app ui.App
}

func (m appModel) Init() tea.Cmd { return m.app.Init() }

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	app, cmd := m.app.Update(msg)
	m.app = app
	return m, cmd
}

func (m appModel) View() string { return m.app.View() }

func runRoot(_ *cobra.Command, _ []string) error {
	backend := mail.NewMockBackend()
	app := ui.NewApp(theme.Nord, backend)

	p := tea.NewProgram(appModel{app: app}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running poplar: %w", err)
	}
	return nil
}

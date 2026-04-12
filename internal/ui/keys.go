package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeys are handled by the root App model.
type GlobalKeys struct {
	Help key.Binding
	Cmd  key.Binding
	Quit key.Binding
}

// NewGlobalKeys returns the default global key bindings.
func NewGlobalKeys() GlobalKeys {
	return GlobalKeys{
		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Cmd:  key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// FolderJumpKeys jump to canonical folders.
type FolderJumpKeys struct {
	Inbox   key.Binding
	Drafts  key.Binding
	Sent    key.Binding
	Archive key.Binding
	Spam    key.Binding
	Trash   key.Binding
}

// NewFolderJumpKeys returns the default folder jump key bindings.
func NewFolderJumpKeys() FolderJumpKeys {
	return FolderJumpKeys{
		Inbox:   key.NewBinding(key.WithKeys("I"), key.WithHelp("I", "inbox")),
		Drafts:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "drafts")),
		Sent:    key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sent")),
		Archive: key.NewBinding(key.WithKeys("A"), key.WithHelp("A", "archive")),
		Spam:    key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "spam")),
		Trash:   key.NewBinding(key.WithKeys("T"), key.WithHelp("T", "trash")),
	}
}

// footerHint is one entry in the footer's keybinding display. These
// are display-only — actual key dispatch lives in each component's
// Update method. dropRank controls responsive behavior: when the
// footer can't fit the full hint list, hints with higher dropRank
// are dropped first. dropRank 0 hints are always kept (escape hatch
// so the user can always reach help/quit).
type footerHint struct {
	key      string
	desc     string
	dropRank int
}

// hint is a short constructor for footerHint literals.
func hint(key, desc string, dropRank int) footerHint {
	return footerHint{key: key, desc: desc, dropRank: dropRank}
}

// AccountKeys produces the unified one-pane account footer hints.
type AccountKeys struct{}

// NewAccountKeys returns the default account view key bindings.
func NewAccountKeys() AccountKeys { return AccountKeys{} }

// FooterGroups returns the footer hint groups in display order.
//
// Hints are tagged with a per-hint drop rank that drives responsive
// behavior. Rough drop order (highest rank first):
//
//   - nav entries (10, 9) — vim/arrow users don't need the hint
//   - v select (8), n/N results (7) — niche modes, discoverable via help
//   - . read (5), s star (4), f fwd (3), / find (3) — secondary actions
//   - r/R reply (2), c compose (2) — primary compose actions
//   - d del (1), a archive (1) — primary triage
//   - ? help (0), : cmd (0), q quit (0) — always kept
func (k AccountKeys) FooterGroups() [][]footerHint {
	return [][]footerHint{
		{
			hint("j/k/J/K", "nav", 10),
			hint("I/D/S/A", "folders", 9),
		},
		{
			hint("d", "del", 1),
			hint("a", "archive", 1),
			hint("s", "star", 4),
			hint(".", "read", 5),
		},
		{
			hint("r/R", "reply", 2),
			hint("f", "fwd", 3),
			hint("c", "compose", 2),
		},
		{
			hint("/", "find", 3),
			hint("n/N", "results", 7),
			hint("v", "select", 8),
		},
		{
			hint("?", "help", 0),
			hint(":", "cmd", 0),
			hint("q", "quit", 0),
		},
	}
}

// ViewerKeys produces the viewer footer hints.
type ViewerKeys struct{}

// NewViewerKeys returns the default viewer key bindings.
func NewViewerKeys() ViewerKeys { return ViewerKeys{} }

// FooterGroups returns the viewer footer hint groups in display order.
//
// Drop ranks: reply drops before triage (triage is more essential in
// the viewer). Tab/q/?/: are always kept.
func (k ViewerKeys) FooterGroups() [][]footerHint {
	return [][]footerHint{
		{
			hint("d", "del", 1),
			hint("a", "archive", 1),
			hint("s", "star", 4),
			hint(".", "read", 5),
		},
		{
			hint("r/R", "reply", 2),
			hint("f", "fwd", 3),
			hint("c", "compose", 2),
		},
		{
			hint("Tab", "links", 0),
			hint("q", "close", 0),
			hint("?", "help", 0),
			hint(":", "cmd", 0),
		},
	}
}

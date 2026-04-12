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

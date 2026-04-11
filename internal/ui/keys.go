package ui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeys are handled by the root App model.
type GlobalKeys struct {
	Tab1 key.Binding
	Tab2 key.Binding
	Tab3 key.Binding
	Tab4 key.Binding
	Tab5 key.Binding
	Tab6 key.Binding
	Tab7 key.Binding
	Tab8 key.Binding
	Tab9 key.Binding
	Help key.Binding
	Cmd  key.Binding
	Quit key.Binding
}

// NewGlobalKeys returns the default global key bindings.
func NewGlobalKeys() GlobalKeys {
	return GlobalKeys{
		Tab1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "tab 1")),
		Tab2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "tab 2")),
		Tab3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "tab 3")),
		Tab4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "tab 4")),
		Tab5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "tab 5")),
		Tab6: key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "tab 6")),
		Tab7: key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "tab 7")),
		Tab8: key.NewBinding(key.WithKeys("8"), key.WithHelp("8", "tab 8")),
		Tab9: key.NewBinding(key.WithKeys("9"), key.WithHelp("9", "tab 9")),
		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Cmd:  key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// MsgListKeys are shown in the footer when the message list is focused.
type MsgListKeys struct {
	Delete   key.Binding
	Archive  key.Binding
	Star     key.Binding
	Reply    key.Binding
	ReplyAll key.Binding
	Forward  key.Binding
	Compose  key.Binding
	Search   key.Binding
	Help     key.Binding
	Cmd      key.Binding
}

// NewMsgListKeys returns the default message list key bindings.
func NewMsgListKeys() MsgListKeys {
	return MsgListKeys{
		Delete:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "del")),
		Archive:  key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "archive")),
		Star:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "star")),
		Reply:    key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reply")),
		ReplyAll: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "all")),
		Forward:  key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "fwd")),
		Compose:  key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compose")),
		Search:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Cmd:      key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
	}
}

// ShortHelp implements help.KeyMap for the message list context.
func (k MsgListKeys) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Delete, k.Archive, k.Star, k.Reply, k.ReplyAll,
		k.Forward, k.Compose, k.Search, k.Help, k.Cmd,
	}
}

// FullHelp implements help.KeyMap for the message list context.
func (k MsgListKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

// SidebarKeys are shown in the footer when the sidebar is focused.
type SidebarKeys struct {
	Open    key.Binding
	Compose key.Binding
	Cmd     key.Binding
}

// NewSidebarKeys returns the default sidebar key bindings.
func NewSidebarKeys() SidebarKeys {
	return SidebarKeys{
		Open:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("Enter", "open")),
		Compose: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "compose")),
		Cmd:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "cmd")),
	}
}

// ShortHelp implements help.KeyMap for the sidebar context.
func (k SidebarKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Open, k.Compose, k.Cmd}
}

// FullHelp implements help.KeyMap for the sidebar context.
func (k SidebarKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

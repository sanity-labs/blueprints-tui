package tui

import "charm.land/bubbles/v2/key"

type appKeyMap struct {
	Quit    key.Binding
	Back    key.Binding
	Select  key.Binding
	Tab     key.Binding
	ShiftTab key.Binding
	Refresh key.Binding
	Help    key.Binding
}

var appKeys = appKeyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
}

func (k appKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Back, k.Tab, k.Refresh, k.Quit}
}

func (k appKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Select, k.Back, k.Tab, k.ShiftTab},
		{k.Refresh, k.Help, k.Quit},
	}
}

package ui

import "github.com/Nathan-ma/hubstaff-tui/internal/config"

// KeyMap holds the resolved keybindings used throughout the TUI.
// Values are Bubbletea key strings (e.g. "ctrl+e", "esc", "?").
type KeyMap struct {
	Quit         string
	Stop         string
	Refresh      string
	Filter       string
	Help         string
	Summary      string
	SwitchPane   string
	GlobalSearch string
}

// NewKeyMap creates a KeyMap from the user's keybindings configuration.
func NewKeyMap(cfg config.KeybindingsConfig) KeyMap {
	return KeyMap{
		Quit:         cfg.Quit,
		Stop:         cfg.Stop,
		Refresh:      cfg.Refresh,
		Filter:       cfg.Filter,
		Help:         cfg.Help,
		Summary:      cfg.Summary,
		SwitchPane:   cfg.SwitchPane,
		GlobalSearch: cfg.GlobalSearch,
	}
}

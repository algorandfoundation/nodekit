package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

type BootstrapMsg struct {
	Install bool
	Catchup bool
}

type BoostrapSelected BootstrapMsg

// EmitBootstrapSelection waits for and retrieves a new set of table rows from a given channel.
func EmitBootstrapSelection(selection BoostrapSelected) tea.Cmd {
	return func() tea.Msg {
		return selection
	}
}

package app

import (
	"github.com/algorandfoundation/algorun-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
)

type AccountSelected internal.Account

// EmitAccountSelected waits for and retrieves a new set of table rows from a given channel.
func EmitAccountSelected(account internal.Account) tea.Cmd {
	return func() tea.Msg {
		return AccountSelected(account)
	}
}

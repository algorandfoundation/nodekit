package accounts

import (
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
)

// EmitAccountSelected waits for and retrieves a new set of table rows from a given channel.
func EmitAccountSelected(account internal.Account) tea.Cmd {
	return func() tea.Msg {
		return account
	}
}

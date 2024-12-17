package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Page represents different pages that can be displayed in the application's viewport.
type Page string

const (

	// AccountsPage represents the page within the application used for managing and displaying account information.
	AccountsPage Page = "accounts"

	// KeysPage represents the page within the application used for managing and displaying key-related information.
	KeysPage Page = "keys"
)

// EmitShowPage returns a command that emits a tea.Msg containing the given Page to be displayed in the application's viewport.
func EmitShowPage(page Page) tea.Cmd {
	return func() tea.Msg {
		return page
	}
}

package app

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Page represents different pages that can be displayed in the application's viewport.
type Page string

const (
	AccountsPage Page = "accounts"
	KeysPage     Page = "keys"
)

func EmitShowPage(page Page) tea.Cmd {
	return func() tea.Msg {
		return page
	}
}

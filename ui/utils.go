package ui

import tea "github.com/charmbracelet/bubbletea"

// waitForUint64 handles an uint64 subscription channel as a tea.Message
func waitForUint64(sub chan uint64) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

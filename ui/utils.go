package ui

import tea "github.com/charmbracelet/bubbletea"

// waitForUint64 handles an uint64 subscription channel as a tea.Message
func waitForUint64(sub chan uint64) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}

// hidden returns 0 when the width is greater than the fill
func hidden(width int, fillSize int) int {
	if fillSize < width {
		return 0
	}
	return width
}

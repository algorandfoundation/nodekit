package accounts

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	return style.WithTitle("Accounts", lipgloss.JoinVertical(lipgloss.Center, style.ApplyBorder(m.Width-3, m.Height-4, "8").Render(m.table.View()), m.controls.View()))
}

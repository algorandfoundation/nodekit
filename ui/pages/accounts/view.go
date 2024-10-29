package accounts

import (
	"github.com/algorandfoundation/hack-tui/ui/pages"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	return pages.WithTitle("Accounts", lipgloss.JoinVertical(lipgloss.Center, pages.PageBorder(m.Width-3).Render(m.table.View()), m.controls.View()))
}

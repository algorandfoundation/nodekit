package keys

import (
	"github.com/algorandfoundation/hack-tui/ui/pages"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Center, pages.Padding1(m.table.View()), m.controls.View())
}

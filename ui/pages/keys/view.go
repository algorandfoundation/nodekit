package keys

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	return style.WithTitle("Keys", lipgloss.JoinVertical(lipgloss.Center, style.ApplyBorder(m.Width-3, m.Height-5, "8").Render(m.table.View()), m.controls.View()))
}

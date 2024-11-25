package controls

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the model's content if it is visible, aligning it horizontally and ensuring it fits within the specified width.
func (m Model) View() string {
	if !m.IsVisible {
		return ""
	}
	render := controlStyle.Render(m.Content)
	difference := m.Width - lipgloss.Width(render)
	line := strings.Repeat("─", max(0, difference/2))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, render, line)
}

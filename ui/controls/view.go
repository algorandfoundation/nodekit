package controls

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

var roundedBorder = func() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder())
}()
var controlStyle = func() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	b.Right = "├"
	return roundedBorder.BorderStyle(b)
}()

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

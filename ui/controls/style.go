package controls

import "github.com/charmbracelet/lipgloss"

var controlStyle = func() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	b.Right = "├"
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderStyle(b)
}()

package generate

import (
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	m.Input.Focused()
	render := m.Input.View()

	if lipgloss.Width(render) < 70 {
		return lipgloss.NewStyle().Width(70).Render(render)
	}
	return render
}

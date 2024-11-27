package generate

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	render := ""
	switch m.Step {
	case AddressStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"",
			"Create keys required to participate in Algorand consensus.",
			"",
			"Account address:",
			m.Input.View(),
			"",
		)
	case DurationStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"",
			"How long should the keys be valid for?",
			"",
			fmt.Sprintf("Duration in %ss:", m.Range),
			m.InputTwo.View(),
			"",
		)
	case WaitingStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"Generating Participation Keys...",
			"",
			"Please wait. This operation can take a few minutes.")
	}

	return lipgloss.NewStyle().Width(70).Render(render)
}

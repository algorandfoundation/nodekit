package generate

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	render := ""
	//m.Input.Focus()
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
			"Please wait. This operation can take a few minutes.")
	}

	if lipgloss.Width(render) < 70 {
		return lipgloss.NewStyle().Width(70).Render(render)
	}
	return render
}

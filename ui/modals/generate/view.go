package generate

import (
	"fmt"

	"github.com/algorandfoundation/algorun-tui/ui/style"
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
		if m.InputError != "" {
			render = lipgloss.JoinVertical(lipgloss.Left,
				render,
				style.Red.Render(m.InputError),
			)
		}
	case DurationStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"",
			"How long should the keys be valid for?",
			"",
			fmt.Sprintf("Duration in %ss:", m.Range),
			m.InputTwo.View(),
			"",
		)
		if m.InputTwoError != "" {
			render = lipgloss.JoinVertical(lipgloss.Left,
				render,
				style.Red.Render(m.InputTwoError),
			)
		}
	case WaitingStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"",
			"Generating Participation Keys...",
			"",
			"Please wait. This operation can take a few minutes.",
			"")
	}

	return lipgloss.NewStyle().Width(70).Render(render)
}

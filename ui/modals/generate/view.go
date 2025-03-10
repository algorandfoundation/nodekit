package generate

import (
	"fmt"

	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// Title returns a string representing the title based on the current step in the ViewModel.
func (m ViewModel) Title() string {
	switch m.Step {
	case DurationStep:
		return "Validity Range"
	case WaitingStep:
		return "Generating Keys"
	default:
		return "Generate Consensus Participation Keys"
	}
}

// BorderColor returns a string representing the border color based on the current step in the ViewModel.
func (m ViewModel) BorderColor() string {
	switch m.Step {
	case WaitingStep:
		return "9"
	default:
		return "2"
	}
}

// Controls returns a string representation of the available control options for the ViewModel.
func (m ViewModel) Controls() string {
	if m.Step == DurationStep {
		return "| " + style.Red.Render("(esc) to cancel") + " |"
	}
	return ""
}

// Navigation returns a string representing control instructions based on the current step in the ViewModel.
func (m ViewModel) Navigation() string {
	switch m.Step {
	case AddressStep:
		return style.Bold("( esc to cancel )")
	case DurationStep:
		return style.Bold("( (s)witch range )")
	default:
		return ""
	}
}

// Body returns a styled string representation of content based on the current step in the ViewModel.
func (m ViewModel) Body() string {
	render := ""
	switch m.Step {
	case AddressStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"",
			"Create keys required to participate in Algorand consensus.",
			"",
			"Account address:",
			m.AddressInput.View(),
			"",
		)
		if m.AddressInputError != "" {
			render = lipgloss.JoinVertical(lipgloss.Left,
				render,
				style.Red.Render(m.AddressInputError),
			)
		}
	case DurationStep:
		render = lipgloss.JoinVertical(lipgloss.Left,
			"",
			"How long should the keys be valid for?",
			"",
			fmt.Sprintf("Duration in %ss:", m.Range),
			m.DurationInput.View(),
			"",
		)
		if m.DurationInputError != "" {
			render = lipgloss.JoinVertical(lipgloss.Left,
				render,
				style.Red.Render(m.DurationInputError),
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

// View renders the ViewModel as a styled string, incorporating title, controls, and body content with dynamic borders.
func (m ViewModel) View() string {
	body := m.Body()
	width := lipgloss.Width(body)
	height := lipgloss.Height(body)
	return style.WithControls(m.Controls(), style.WithNavigation(
		m.Navigation(),
		style.WithTitle(
			m.Title(),
			// Apply the Borders with the Padding
			style.ApplyBorder(width+2, height-4, m.BorderColor()).
				PaddingRight(1).
				PaddingLeft(1).
				Render(m.Body()),
		),
	))

}

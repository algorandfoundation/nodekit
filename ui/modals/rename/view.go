package rename

import (
	"fmt"

	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/algorandfoundation/nodekit/ui/utils"
	"github.com/charmbracelet/lipgloss"
)

// Title returns the modal title.
func (m ViewModel) Title() string {
	return "Set Account Nickname"
}

// BorderColor returns the modal border color.
func (m ViewModel) BorderColor() string {
	return "2"
}

// Controls returns the available control hints for the modal.
func (m ViewModel) Controls() string {
	return "| " + style.Red.Render("(esc) to cancel") + " |"
}

// Navigation returns the navigation hint for the modal.
func (m ViewModel) Navigation() string {
	return style.Bold("( (enter) to save | empty to clear )")
}

// Body renders the modal contents.
func (m ViewModel) Body() string {
	render := lipgloss.JoinVertical(lipgloss.Left,
		"",
		fmt.Sprintf("Local nickname for %s:", utils.ShortAddress(m.Address)),
		"",
		m.Input.View(),
		"",
		lipgloss.NewStyle().Faint(true).Render("Stored locally on this machine in ~/.nodekit.json"),
		"",
	)
	if m.InputError != "" {
		render = lipgloss.JoinVertical(lipgloss.Left,
			render,
			style.Red.Render(m.InputError),
		)
	}
	return lipgloss.NewStyle().Width(70).Render(render)
}

// View renders the ViewModel as a styled string.
func (m ViewModel) View() string {
	body := m.Body()
	width := lipgloss.Width(body)
	height := lipgloss.Height(body)
	return style.WithControls(m.Controls(), style.WithNavigation(
		m.Navigation(),
		style.WithTitle(
			m.Title(),
			style.ApplyBorder(width+2, height-4, m.BorderColor()).
				PaddingRight(1).
				PaddingLeft(1).
				Render(m.Body()),
		),
	))
}

package keys

import (
	"fmt"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	if m.SelectedKeyToDelete != nil {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			renderDeleteConfirmationModal(m.SelectedKeyToDelete),
		)
	}
	table := style.ApplyBorder(m.Width, m.Height, "8").Render(m.table.View())
	return style.WithNavigation(
		m.navigation,
		style.WithControls(
			m.controls,
			style.WithTitle(
				"Keys",
				table,
			),
		),
	)
}

func renderDeleteConfirmationModal(partKey *api.ParticipationKey) string {
	modalStyle := lipgloss.NewStyle().
		Width(60).
		Height(7).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	modalContent := fmt.Sprintf("Participation Key: %v\nAccount Address: %v\nPress either y (yes) or n (no).", partKey.Id, partKey.Address)

	return modalStyle.Render("Are you sure you want to delete this key from your node?\n", modalContent)
}

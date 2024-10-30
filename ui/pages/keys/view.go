package keys

import (
	"fmt"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/ui/pages"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {
	if m.SelectedKeyToDelete != nil {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			renderDeleteConfirmationModal(m.SelectedKeyToDelete, m.DeleteLoading),
			m.controls.View(),
		)
	}
	return lipgloss.JoinVertical(
		lipgloss.Center,
		pages.Padding1(m.table.View()),
		m.controls.View(),
	)
}

func renderDeleteConfirmationModal(partKey *api.ParticipationKey, deleteLoading bool) string {
	modalStyle := lipgloss.NewStyle().
		Width(60).
		Height(7).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2)

	var modalContent string
	if deleteLoading {
		s := spinner.New()
		s.Spinner = spinner.Dot
		modalContent = fmt.Sprintf("Deleting key...\n%s", s.View())
	} else {
		modalContent = fmt.Sprintf("Are you sure you want to delete this key from your node?\nParticipation Key: %v\nAccount Address: %v\nPress either y (yes) or n (no).", partKey.Id, partKey.Address)
	}

	return modalStyle.Render(modalContent)
}

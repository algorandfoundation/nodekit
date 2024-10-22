package transaction

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func (m ViewModel) View() string {
	pad := strings.Repeat("\n", max(0, m.Height/2-1))
	return lipgloss.JoinVertical(lipgloss.Center, pad, "TODO", pad, m.controls.View())
}

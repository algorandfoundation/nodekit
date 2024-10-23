package transaction

import (
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {

	qrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("0"))

	red := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	if m.asciiQR == "" || m.urlTxn == "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"No QR Code available",
			m.controls.View())
	}
	if m.QRWontFit {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			red.Render("QR Code too large to display... Please adjust terminal dimensions or font."),
			m.controls.View())
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		qrStyle.Render(m.asciiQR),
		m.controls.View())
}

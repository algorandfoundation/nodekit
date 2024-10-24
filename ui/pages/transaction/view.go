package transaction

import (
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {

	qrStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("0"))

	urlStyle := lipgloss.NewStyle().
		Width(m.Width - 2)

	red := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Width(m.Width - 2)

	yellow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Width(m.Width - 2)

	var Padding1 = lipgloss.NewStyle().Padding().Render

	if m.asciiQR == "" || m.urlTxn == "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			"No QR Code or TxnURL available",
			"\n",
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
		yellow.Render(m.hint),
		Padding1(),
		qrStyle.Render(m.asciiQR),
		urlStyle.Render(m.urlTxn),
		m.controls.View())
}

package transaction

import (
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) View() string {

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
			red.Width(m.Width-2).Render("QR Code too large to display... Please adjust terminal dimensions or font."),
			m.controls.View())
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		yellow.Width(m.Width-2).Render(m.hint),
		Padding1.Render(),
		qrStyle.Render(m.asciiQR),
		urlStyle.Width(m.Width-2).Render(m.urlTxn),
		m.controls.View())
}

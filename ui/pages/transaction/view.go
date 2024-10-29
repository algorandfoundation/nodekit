package transaction

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func (m ViewModel) View() string {
	qrRender := lipgloss.JoinVertical(
		lipgloss.Center,
		style.Yellow.Render(m.hint),
		"",
		qrStyle.Render(m.asciiQR),
		urlStyle.Render(m.urlTxn),
	)

	if m.asciiQR == "" || m.urlTxn == "" {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			"No QR Code or TxnURL available",
			"\n",
			m.controls.View())
	}

	if lipgloss.Height(qrRender) > m.Height {
		padHeight := max(0, m.Height-lipgloss.Height(m.controls.View())-1)
		padHString := strings.Repeat("\n", padHeight/2)
		text := style.Red.Render("QR Code too large to display... Please adjust terminal dimensions or font.")
		padWidth := max(0, m.Width-lipgloss.Width(text))
		padWString := strings.Repeat(" ", padWidth/2)
		return lipgloss.JoinVertical(
			lipgloss.Left,
			padHString,
			lipgloss.JoinHorizontal(lipgloss.Left, padWString, style.Red.Render("QR Code too large to display... Please adjust terminal dimensions or font.")),
			padHString,
			m.controls.View())
	}

	qrRenderPadHeight := max(0, m.Height-(lipgloss.Height(qrRender)-lipgloss.Height(m.controls.View()))-1)
	qrPad := strings.Repeat("\n", qrRenderPadHeight/2)
	return lipgloss.JoinVertical(
		lipgloss.Center,
		qrPad,
		qrRender,
		qrPad,
		m.controls.View(),
	)
}

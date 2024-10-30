package transaction

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

func (m ViewModel) View() string {
	qrCode := qrStyle.Render(m.asciiQR)
	qrWidth := lipgloss.Width(qrCode) + 1
	qrHeight := lipgloss.Height(qrCode)
	title := ""
	if m.IsOnline {
		title = "Offline Transaction"
	} else {
		title = "Online Transaction"
	}

	url := ""
	if lipgloss.Width(m.urlTxn) > qrWidth {
		url = m.urlTxn[:(qrWidth-3)] + "..."
	} else {
		url = m.urlTxn
	}

	var render string
	if qrWidth > m.Width || qrHeight+2 > m.Height {
		text := style.Red.Render("QR Code too large to display... Please adjust terminal dimensions or font.")
		padHeight := max(0, m.Height-lipgloss.Height(text))
		padHString := strings.Repeat("\n", padHeight/2)
		padWidth := max(0, m.Width-lipgloss.Width(text))
		padWString := strings.Repeat(" ", padWidth/2)
		paddedStr := lipgloss.JoinVertical(
			lipgloss.Left,
			padHString,
			lipgloss.JoinHorizontal(lipgloss.Left, padWString, style.Red.Render("QR Code too large to display... Please adjust terminal dimensions or font.")),
		)
		render = style.ApplyBorder(m.Width, m.Height, "8").Render(paddedStr)
	} else {
		qRemainingWidth := max(0, (m.Width-lipgloss.Width(qrCode))/2)
		qrCode = lipgloss.JoinHorizontal(lipgloss.Left, strings.Repeat(" ", qRemainingWidth), qrCode, strings.Repeat(" ", qRemainingWidth))
		qRemainingHeight := max(0, (m.Height-2-lipgloss.Height(qrCode))/2)
		if qrHeight+2 == m.Height {
			qrCode = lipgloss.JoinVertical(lipgloss.Center, style.Yellow.Render(m.hint), qrCode, urlStyle.Render(url))
		} else {
			qrCode = lipgloss.JoinVertical(lipgloss.Center, strings.Repeat("\n", qRemainingHeight), style.Yellow.Render(m.hint), qrCode, urlStyle.Render(url))

		}
		render = style.ApplyBorder(m.Width, m.Height, "8").Render(qrCode)
	}
	return style.WithNavigation(
		m.navigation,
		style.WithControls(
			m.controls,
			style.WithTitle(
				title,
				render,
			),
		),
	)
}

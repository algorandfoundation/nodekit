package transaction

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m ViewModel) View() string {
	if m.ActiveKey == nil {
		return "No key selected"
	}
	if m.ATxn == nil {
		return "Loading..."
	}
	txn, err := m.ATxn.ProduceQRCode()
	if err != nil {
		return "Something went wrong"
	}

	render := qrStyle.Render(txn)

	width := lipgloss.Width(render)
	height := lipgloss.Height(render)

	if width > m.Width || height > m.Height {
		return style.Red.Render(ansi.Wordwrap("QR Code too large to display... Please adjust terminal dimensions or font.", m.Width, " "))

	}

	return render
}

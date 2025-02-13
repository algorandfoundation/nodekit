package transaction

import (
	"fmt"

	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	OnlineTitle  = "Register Online"
	OfflineTitle = "Register Offline"
)

func (m ViewModel) Title() string {
	if m.OfflineControls {
		return OfflineTitle
	} else {
		return OnlineTitle
	}
}
func (m ViewModel) BorderColor() string {
	return "9"
}
func (m ViewModel) Controls() string {
	escLegend := style.Red.Render("(esc) go back")
	if m.IsQREnabled() {
		otherView := "link"
		if m.ShowLink {
			otherView = "QR"
		}
		return "( " + style.Yellow.Render("(s)how "+otherView) + " | " + escLegend + " )"
	}
	return "( " + escLegend + " )"
}

func (m ViewModel) Body() string {
	if m.Participation == nil {
		return "No key selected"
	}
	if m.ATxn == nil || m.Link == nil {
		return "Loading..."
	}

	var adj string
	isOffline := m.ATxn.AUrlTxnKeyreg.VotePK == nil
	if isOffline {
		adj = "offline"
	} else {
		adj = "online"
	}

	intro := fmt.Sprintf("Sign this transaction to register your account as %s", adj)
	render := intro

	if !m.ShowLink {
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			style.Green.Render("Scan the QR code with Pera")+" or "+style.Yellow.Render("press S to show a link instead"),
		)
	}

	if m.ShouldAddIncentivesFee() {
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			"",
			style.Bold("Note: Transction fee set to 2 ALGO (opting in to rewards)"),
		)
	}

	if isOffline {
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			"",
			style.Bold("Note: this will take effect after 320 rounds (~15 min.)"),
			"Please keep your node running during this cooldown period.",
		)
	}

	if m.ShowLink {
		link := participation.ToShortLink(*m.Link, m.ShouldAddIncentivesFee())
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			"",
			"Open this URL in your browser:",
			"",
			style.WithHyperlink(link, link),
		)
	} else {
		// TODO: Refactor ATxn to Interface
		txn, err := m.ATxn.ProduceQRCode()
		if err != nil {
			return "Something went wrong"
		}
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			qrStyle.Render(txn),
		)
	}

	width := lipgloss.Width(render)
	height := lipgloss.Height(render)

	if !m.ShowLink && (width > m.Width || height > m.Height) {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			intro,
			"",
			style.Red.Render(ansi.Wordwrap("QR code is available but it does not fit on screen.", m.Width, " ")),
			style.Red.Render(ansi.Wordwrap("Adjust terminal dimensions/font size to display.", m.Width, " ")),
			"",
			ansi.Wordwrap("Or press S to switch to Link view.", m.Width, " "),
		)
	}

	return render
}

// View renders the ViewModel as a styled string, incorporating title, controls, and body content with dynamic borders.
func (m ViewModel) View() string {
	body := m.Body()
	width := lipgloss.Width(body)
	height := lipgloss.Height(body)
	return style.WithNavigation(
		m.Controls(),
		style.WithTitle(
			m.Title(),
			// Apply the Borders with the Padding
			style.ApplyBorder(width+2, height+2, m.BorderColor()).
				Padding(1).
				Render(m.Body()),
		),
	)

}

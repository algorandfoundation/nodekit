package transaction

import (
	"fmt"

	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m ViewModel) View() string {
	if m.Participation == nil {
		return "No key selected"
	}
	if m.ATxn == nil || m.Link == nil {
		// TODO make sure the link modal actually shows loading
		// currently it "hangs" in the previous screen when pressing R until the shortener service responds
		return "Loading..."
	}

	var adj string
	isOffline := m.ATxn.AUrlTxnKeyreg.VotePK == nil
	if isOffline {
		adj = "offline"
	} else {
		adj = "online"
	}

	textHeader := "Sign this transaction to register your account as %s"
	textOpenURLInBrowser := "Open this URL in your browser:"
	textPeraUserPressSForQr := style.Italics("Pera wallet user? Press S to scan a QR instead.")
	text2AFeeWarning := style.Yellow.Render(style.Bold(("⚠ Transaction fee set to 2 ALGO (opting in to rewards)")))
	textScanQrOrPressS := style.Green.Render("Scan the QR code with Pera") + " or " + style.Yellow.Render("press S to show a link instead")
	textOffline320Note := style.Bold("Note: this will take effect after 320 rounds (~15 min.)\n") + "Please keep your node running during this cooldown period."

	intro := fmt.Sprintf(textHeader, adj)
	render := intro

	if m.ShowLink {
		link := participation.ToShortLink(*m.Link, m.ShouldAddIncentivesFee())

		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			"",
			textOpenURLInBrowser,
			"",
			style.WithHyperlink(link, link),
		)
		if m.ShouldAddIncentivesFee() {
			render = lipgloss.JoinVertical(
				lipgloss.Center,
				render,
				"",
				text2AFeeWarning,
			)
		}
		if isOffline {
			render = lipgloss.JoinVertical(
				lipgloss.Center,
				render,
				"",
				textOffline320Note,
			)
		}
		if m.IsQREnabled() {
			render = lipgloss.JoinVertical(
				lipgloss.Center,
				render,
				"",
				textPeraUserPressSForQr,
			)
		}

	} else {
		// TODO: Refactor ATxn to Interface
		txn, err := m.ATxn.ProduceQRCode()

		// TODO responsive vertical spaces? calculate on modal/screen height diff
		if m.ShouldAddIncentivesFee() {
			render = lipgloss.JoinVertical(
				lipgloss.Center,
				render,
				text2AFeeWarning,
			)
		}
		if isOffline {
			render = lipgloss.JoinVertical(
				lipgloss.Center,
				render,
				textOffline320Note,
			)
		}
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			textScanQrOrPressS,
		)
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

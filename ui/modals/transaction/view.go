package transaction

import (
	"fmt"

	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var textHeader = "Sign this transaction to register your account as %s"
var textOpenURLInBrowser = "Open this URL in your browser:"
var textPeraUserPressSForQr = style.Italics("Pera wallet user? Press S to scan a QR instead.")
var text2AFeeWarning = style.Red.Render(style.Bold(("⚠ Transaction fee set to 2 ALGO (opting in to rewards)")))
var textScanQrOrPressS = style.Green.Render("Scan the QR code with Pera") + " or " + style.Yellow.Render("press S to show a link instead")
var textOffline320Note1 = style.Bold("Note: this will take effect after 320 rounds (~15 min.)")
var textOffline320Note2 = "Please keep your node running during this cooldown period."

func (m ViewModel) isOffline() bool {
	return m.ATxn.AUrlTxnKeyreg.VotePK == nil
}

func (m ViewModel) renderQRModal(renderIn string, txn string, spacing int) (render string) {
	remainingSpaces := spacing
	render = renderIn
	if m.ShouldAddIncentivesFee() {
		if remainingSpaces > 0 {
			render = lipgloss.JoinVertical(lipgloss.Center, render, "")
			remainingSpaces -= 1
		}
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			text2AFeeWarning,
		)
	}
	if m.isOffline() {
		if remainingSpaces > 0 {
			render = lipgloss.JoinVertical(lipgloss.Center, render, "")
			remainingSpaces -= 1
		}
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			render,
			textOffline320Note1,
			textOffline320Note2,
		)
	}
	// if we did not emit a space already but we can
	if remainingSpaces > 0 {
		render = lipgloss.JoinVertical(lipgloss.Center, render, "")
		remainingSpaces -= 1
	}
	render = lipgloss.JoinVertical(
		lipgloss.Center,
		render,
		textScanQrOrPressS,
	)
	if remainingSpaces > 0 {
		render = lipgloss.JoinVertical(lipgloss.Center, render, "")
		remainingSpaces -= 1
	}
	render = lipgloss.JoinVertical(
		lipgloss.Center,
		render,
		qrStyle.Render(txn),
	)
	return
}

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
	if m.isOffline() {
		adj = "offline"
	} else {
		adj = "online"
	}

	intro := fmt.Sprintf(textHeader, adj)
	render := ""
	qrSpacing := 3
	qrText, qrErr := m.ATxn.ProduceQRCode()

	if m.ShowLink {
		link := participation.ToShortLink(*m.Link, m.ShouldAddIncentivesFee())

		render = lipgloss.JoinVertical(
			lipgloss.Center,
			intro,
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
		if m.isOffline() {
			render = lipgloss.JoinVertical(
				lipgloss.Center,
				render,
				"",
				textOffline320Note1,
				textOffline320Note2,
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
		if qrErr != nil {
			return lipgloss.JoinVertical(
				lipgloss.Center,
				"Something went wrong while generating the QR code",
				"Press S to display a link",
			)
		}

		render = m.renderQRModal(intro, qrText, qrSpacing)
	}

	width := lipgloss.Width(render)
	height := lipgloss.Height(render)

	for qrSpacing > 0 && !m.ShowLink && (width > m.Width || height > m.Height) {
		qrSpacing -= 1
		render = m.renderQRModal(intro, qrText, qrSpacing)
		width = lipgloss.Width(render)
		height = lipgloss.Height(render)
	}

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

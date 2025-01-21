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
		return "Loading..."
	}
	// TODO: Refactor ATxn to Interface
	txn, err := m.ATxn.ProduceQRCode()
	if err != nil {
		return "Something went wrong"
	}

	var adj string
	isOffline := m.ATxn.AUrlTxnKeyreg.VotePK == nil
	if isOffline {
		adj = "offline"
	} else {
		adj = "online"
	}
	intro := fmt.Sprintf("Sign this transaction to register your account as %s", adj)
	link := participation.ToShortLink(*m.Link, m.ShouldAddIncentivesFee())
	loraText := lipgloss.JoinVertical(
		lipgloss.Center,
		"Open this URL in your browser:\n",
		style.WithHyperlink(link, link),
	)

	if m.ShouldAddIncentivesFee() {
		loraText = lipgloss.JoinVertical(
			lipgloss.Center,
			loraText,
			"",
			"Note: Transction fee set to 2 ALGO",
			"for staking rewards eligibility",
		)
	}

	if isOffline {
		loraText = lipgloss.JoinVertical(
			lipgloss.Center,
			loraText,
			"",
			"Note: this will take effect after 320 rounds (~15 min.)",
			"Please keep your node running during this cooldown period.",
		)
	}

	var render string
	if m.State.Status.Network == "testnet-v1.0" {
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			intro,
			"",
			"Scan the QR code with Pera or Defly",
			style.Yellow.Render("(make sure you use the "+m.State.Status.Network+" network)"),
			"",
			qrStyle.Render(txn),
			"-or-",
			"",
			loraText,
		)
	} else {
		render = lipgloss.JoinVertical(
			lipgloss.Center,
			"",
			intro,
			"",
			loraText,
			"",
		)
	}

	width := lipgloss.Width(render)
	height := lipgloss.Height(render)

	if width > m.Width || height > m.Height {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			intro,
			"",
			style.Red.Render(ansi.Wordwrap("Mobile QR is available but it does not fit on screen.", m.Width, " ")),
			style.Red.Render(ansi.Wordwrap("Adjust terminal dimensions or font size to display.", m.Width, " ")),
			"",
			"-or-",
			loraText,
		)
	}

	return render
}

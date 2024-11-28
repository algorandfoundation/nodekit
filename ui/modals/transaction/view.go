package transaction

import (
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func (m ViewModel) View() string {
	if m.Participation == nil {
		return "No key selected"
	}
	if m.ATxn == nil {
		return "Loading..."
	}
	// TODO: Refactor ATxn to Interface
	txn, err := m.ATxn.ProduceQRCode()
	if err != nil {
		return "Something went wrong"
	}

	var verb string
	isOffline := m.ATxn.AUrlTxnKeyreg.VotePK == nil
	if isOffline {
		verb = "deregister"
	} else {
		verb = "register"
	}
	intro := "Sign this transaction to " + verb + " your account keys:"

	link, _ := internal.ToLoraDeepLink(m.State.Status.Network, m.Active, *m.Participation)
	loraText := lipgloss.JoinHorizontal(
		lipgloss.Bottom,
		style.WithHyperlink("Click here", link),
		" to sign via Lora.",
	)
	if isOffline {
		loraText = lipgloss.JoinVertical(
			lipgloss.Center,
			loraText,
			"",
			"Note: this will take effect after 320 rounds (15 mins.)",
			"Please keep your node online during this cooldown period.",
		)
	}

	var render string
	if m.State.Status.Network == "testnet-v1.0" || m.State.Status.Network == "mainnet-v1.0" {
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

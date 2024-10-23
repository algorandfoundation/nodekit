package transaction

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/ui/controls"
	"github.com/charmbracelet/lipgloss"
)

var green = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

type ViewModel struct {
	// Width is the last known horizontal lines
	Width int
	// Height is the last known vertical lines
	Height int

	// QRWontFit is a flag to indicate the QR code is too large to display
	QRWontFit bool

	// Participation Key
	Data api.ParticipationKey

	// Components
	controls controls.Model

	// QR Code and URL
	asciiQR string
	urlTxn  string
}

// New creates and instance of the ViewModel with a default controls.Model
func New() ViewModel {
	return ViewModel{
		controls: controls.New(" (a)ccounts | (k)eys | " + green.Render("(t)xn ")),
	}
}

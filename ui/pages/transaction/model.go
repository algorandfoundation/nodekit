package transaction

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
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

	// Pointer to the State
	State *internal.StateModel

	// Components
	controls controls.Model

	// QR Code, URL and hint text
	asciiQR string
	urlTxn  string
	hint    string
}

// New creates and instance of the ViewModel with a default controls.Model
func New(state *internal.StateModel) ViewModel {
	return ViewModel{
		State:    state,
		controls: controls.New(" (a)ccounts | (k)eys | shift+tab: back "),
	}
}

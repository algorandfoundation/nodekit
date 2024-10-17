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

	// Participation Key
	Data api.ParticipationKey

	// Components
	controls controls.Model

	// TODO: add URL
	// urlTxn   string
}

// New creates and instance of the ViewModel with a default controls.Model
func New() ViewModel {
	return ViewModel{
		controls: controls.New(" (a)ccunts | (k)eys | " + green.Render("(t)xn ")),
	}
}

package transaction

import (
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/style"
)

type ViewModel struct {
	// Width is the last known horizontal lines
	Width int
	// Height is the last known vertical lines
	Height int

	// Participation Key
	Data api.ParticipationKey

	// Pointer to the State
	State    *internal.StateModel
	IsOnline bool

	// Components
	controls   string
	navigation string

	// QR Code, URL and hint text
	asciiQR string
	urlTxn  string
	hint    string
}

func (m ViewModel) FormatedAddress() string {
	return fmt.Sprintf("%s...%s", m.Data.Address[0:4], m.Data.Address[len(m.Data.Address)-4:])
}

// New creates and instance of the ViewModel with a default controls.Model
func New(state *internal.StateModel) ViewModel {
	return ViewModel{
		State:      state,
		IsOnline:   false,
		navigation: "| accounts | keys | " + style.Green.Render("txn") + " |",
		controls:   "( <- back )",
	}
}

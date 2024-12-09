package transaction

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/algorandfoundation/algourl/encoder"
)

type ViewModel struct {
	// Width is the last known horizontal lines
	Width int
	// Height is the last known vertical lines
	Height int

	Title string

	// Active Participation Key
	Participation *api.ParticipationKey
	Active        bool

	// Pointer to the State
	State    *internal.StateModel
	IsOnline bool

	// Components
	BorderColor string
	Controls    string
	navigation  string

	// QR Code
	ATxn *encoder.AUrlTxn
}

func (m ViewModel) FormatedAddress() string {
	return fmt.Sprintf("%s...%s", m.Participation.Address[0:4], m.Participation.Address[len(m.Participation.Address)-4:])
}

// New creates and instance of the ViewModel with a default controls.Model
func New(state *internal.StateModel) *ViewModel {
	return &ViewModel{
		State:       state,
		Title:       "Offline Transaction",
		IsOnline:    false,
		BorderColor: "9",
		navigation:  "| accounts | keys | " + style.Green.Render("txn") + " |",
		Controls:    "( " + style.Red.Render("esc") + " )",
		ATxn:        nil,
	}
}

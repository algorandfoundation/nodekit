package transaction

import (
	"fmt"
	"github.com/algorandfoundation/algourl/encoder"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/style"
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
	Link          *participation.ShortLinkResponse

	// Pointer to the State
	State    *algod.StateModel
	IsOnline bool

	// Components
	BorderColor string
	Controls    string
	navigation  string

	ShowLink bool

	// QR Code
	ATxn *encoder.AUrlTxn
}

func (m ViewModel) FormatedAddress() string {
	return fmt.Sprintf("%s...%s", m.Participation.Address[0:4], m.Participation.Address[len(m.Participation.Address)-4:])
}

func (m ViewModel) IsQREnabled() bool {
	return true // TODO
	// return m.State.Status.Network == "testnet-v1.0" || m.State.Status.Network == "mainnet-v1.0"
}

// New creates and instance of the ViewModel with a default controls.Model
func New(state *algod.StateModel) *ViewModel {
	return &ViewModel{
		State:       state,
		Title:       "Offline Transaction",
		ShowLink:    true,
		IsOnline:    false,
		BorderColor: "9",
		navigation:  "| accounts | keys | " + style.Green.Render("txn") + " |",
		Controls:    "",
		ATxn:        nil,
	}
}

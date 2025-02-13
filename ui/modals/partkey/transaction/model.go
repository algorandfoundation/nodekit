package transaction

import (
	"fmt"
	"github.com/algorandfoundation/algourl/encoder"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
)

type ViewModel struct {
	// Width is the last known horizontal lines
	Width int
	// Height is the last known vertical lines
	Height int

	//Title string

	// Active Participation Key
	Participation   *api.ParticipationKey
	OfflineControls bool
	Suspended       bool
	Link            *participation.ShortLinkResponse

	// Pointer to the State
	State    *algod.StateModel
	ShowLink bool

	// QR Code
	ATxn *encoder.AUrlTxn
}

func (m ViewModel) FormatedAddress() string {
	return fmt.Sprintf("%s...%s", m.Participation.Address[0:4], m.Participation.Address[len(m.Participation.Address)-4:])
}

func (m ViewModel) IsQREnabled() bool {
	return m.State.Status.Network == "testnet-v1.0" || m.State.Status.Network == "mainnet-v1.0" || m.State.Status.Network == "tuinet-v1"
}

// New creates and instance of the ViewModel with a default controls.Model
func New(state *algod.StateModel) *ViewModel {
	return &ViewModel{
		State:    state,
		ShowLink: true,
		ATxn:     nil,
	}
}
